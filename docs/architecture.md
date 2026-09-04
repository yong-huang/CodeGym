# CodeGym 架构文档

> 练出来的，才是自己的。

本文档描述 CodeGym 的实际实现架构（依据 commit `fed1e5f` 的代码），配合三张交互式架构图阅读：

| 图 | 类型 | 文件 |
|----|------|------|
| 组件架构图 | architecture | [codegym-architecture.html](diagrams/codegym-architecture.html) |
| 训练闭环时序图 | sequence | [codegym-sequence.html](diagrams/codegym-sequence.html) |
| 题库与判题数据流图 | dataflow | [codegym-dataflow.html](diagrams/codegym-dataflow.html) |

三张图均为自包含 HTML，支持主题切换、缩放、聚焦与关系追踪，直接用浏览器打开即可。

---

## 1. 概述

CodeGym 是一个「填空-运行-验证」的交互式代码训练平台：学员选择训练动作与负重级别，后端从题库随机挖空生成训练变体（掩码代码不含答案），学员在编辑器中填写训练点后提交，平台对 Python/Go 真实执行判题、对其余技术栈做答案匹配，产出逐点评分与成长记录。

| 层 | 技术 | 说明 |
|----|------|------|
| 前端 | React 18 + TypeScript + Vite | 无路由库（自研 hash 路由）、无状态库、无 UI 组件库 |
| 后端 | Go 1.26 · `net/http` 标准库 | 单二进制，无 Web 框架，Go 1.22 方法路由 |
| 数据库 | SQLite（`modernc.org/sqlite`） | 纯 Go 驱动，单连接（`SetMaxOpenConns(1)`）单写者 |
| 执行引擎 | 本地子进程 | `python3 -c` 直跑；Go 写入临时 module 后 `go run .`，带硬超时 |
| AI（可选） | Ollama | 教练端出题，endpoint/model 存 DB，页面可配 |

```
浏览器 (React SPA :5190)
   │  REST /api/v1（Vite 代理 / nginx 反代）
   ▼
API Server (Go :8090) ──► variant / judge / engine / store / seed / llm
   │                         │            │
   ▼                         ▼            ▼
SQLite (data/codegym.db)   题库 JSON    执行子进程 (python3 / go run)
                           (data/bank)  Ollama (:11434，可选)
```

## 2. 组件架构

详见 [组件架构图](diagrams/codegym-architecture.html)。系统分为四层：

### 2.1 前端（`frontend/src/`）

- **App.tsx**：自研 hash 路由（`#/exercise/:id` 等 5 条路由），顶栏导航 + 可折叠侧栏。
- **Sidebar.tsx**：技术栈 → 子分类 → 题目的三级目录树，展示完成进度与收藏。
- **TrainingPage.tsx**：训练闭环页面——负重选择、代码编辑区、辅助保护、运行验证、评分卡。
- **CodeEditor.tsx**：核心组件。透明 `<textarea>` 与高亮 `<pre>` 叠加实现编辑/高亮对齐；内部维护一套**填空区间增量追踪**状态机（见 §4.3）。
- **api/client.ts**：唯一 fetch 封装；教练端请求自动附带 localStorage 中的 `X-Admin-Token`。

### 2.2 HTTP 层（`backend/internal/api/`）

- `handlers.go`：约 27 条路由。学员端（列表/变体/执行/进度/薄弱/收藏/环境）无鉴权，学员身份取 `X-User-Id` 头（缺省 `demo-user`）；教练端 `/api/v1/admin/*` 由 `adminOnly` 中间件保护——启用 `-admin-token` 后要求 `X-Admin-Token` 或 `Bearer`，常数时间比较。
- `ai.go`：Ollama 出题。prompt 要求返回 JSON 数组，解析后在 `full_code` 里 `strings.Index` 定位训练点生成 span，定位失败的题逐条报错不落库。
- 关键安全设计：变体接口 `buildVariantResponse` 发给前端的 `marked_code` 已把训练点替换为 `____FILL_<id>____` 掩码，不携带 `original_text`。

### 2.3 训练内核（四个纯逻辑包）

| 包 | 职责 | 关键函数 |
|----|------|---------|
| `variant` | 负重随机留空：按级别过滤合格训练点（`min_weight` 排名 + `always_fill`），随机选 N 个、跳过重叠 span；`ApplyAnswers` 把答案按 span 从右到左回填 | `SelectPoints` / `ApplyAnswers` |
| `judge` | 双路径判题（见 §4.4）；`GradeFill` 逐点加权评分 + 辅助扣分（L2/L3/L4 = -5/-15/-30）；`orderFreeAssign` 同组顺序互换双射匹配 | `GradeFill` / `GradeBugfix` |
| `engine` | 子进程执行。Python 直跑 `python3 -c`；Go 写临时 module `go run .`；`context.WithTimeout` 硬超时；`EffectiveOutput` 取 stdout 首个非空行 | `Run` / `RunPython` / `RunGo` |
| `seed` | 启动时扫描 `data/bank/*.json` 全量 upsert（只动 exercises 表）；按训练点原文在 `full_code` 中搜索**重算 span 坐标**；`ensureBeginnerPoint` 保证每题入门负重至少一个可填点 | `BuildAll` / `buildExercise` |

### 2.4 数据层（`backend/internal/store/`）

SQLite 7 张表：`exercises`、`attempt_records`、`user_progress`、`favorites`、`categories`、`environments`、`settings`。JSON 字段（fill_points/test_cases 等）整体序列化存 TEXT。打卡 streak 按日历日比较（当天重复作答不加天）。

## 3. 训练闭环（时序）

详见 [训练闭环时序图](diagrams/codegym-sequence.html)。

1. **取变体**：`GET /api/v1/exercises/:id/variant?level=medium` → `store` 读题目 → `variant.SelectPoints` 按负重抽训练点 → 返回掩码代码 + 训练点元数据（含四级提示文本，不含答案原文）。
2. **填写**：CodeEditor 把掩码 token 替换为 `____` 占位符渲染；学员输入，前端实时做**答案提取**并上报父组件。
3. **运行验证**：`POST /api/v1/execute` 携带 `fill_answers`（`FILL_<id>` 键）与 `hint_usage`。后端重建完整代码；`python`/`go` 栈真实子进程执行并比对 stdout，其余技术栈（命令/配置/SQL/K8s/DevOps/AWS 等）直接答案匹配；逐点评分加权得总分，映射 S/A/B/C/D。
4. **成长**：作答写入 `attempt_records`，更新 `user_progress`（累计/正确数/薄弱部位/打卡 streak），错题进薄弱部位页支持一键加练。

## 4. 数据流与关键机制

详见 [题库与判题数据流图](diagrams/codegym-dataflow.html)。

### 4.1 题库即数据源

`backend/data/bank/*.json`（19 个文件、856 题）是题库唯一数据源。JSON 中**不存储任何坐标**——只存训练点原文 `text`，seed 加载时用 `occurrenceIndex` 在 `full_code` 里按出现次序定位、重算 `span_start/span_end`。因此题库可以纯手工编辑，改代码不必算偏移量。教练端与 AI 生成的题目则绕过 seed，直接 upsert（seed 只增不删，绝不改进度与收藏）。

### 4.2 双路径判题

- **真实执行**（`python` / `go` 栈，FastAPI 子类除外）：回填答案 → 子进程运行 → stdout 与 `expected_output` 去空白比对 + 逐训练点匹配评分。
- **答案匹配**（命令/配置/SQL/K8s/OpenShift/DevOps/AWS 等）：不执行，直接逐点比对答案，总分 ≥75 通关。
- 答案匹配先 `normalize`（去全部空白）再比较，支持 `acceptable_answers` 别名（如 `-A` = `--all-namespaces`）与 `order_group` 顺序互换分组（见 §4.5）。

### 4.3 前端填空区间追踪（CodeEditor 最复杂部分）

编辑器是自由文本（整段可编辑），需要从用户编辑后的文本反推每个填空的答案区间。策略为三向路由：

1. **普通编辑**（退格、输入、删光重输、多删破坏代码结构）→ 信任**增量追踪**：`diffEdit` 定位本次编辑的旧/新区间，`applyEdit` 精确伸缩受影响的填空区间；纯插入落在零宽区间（空答案）上时扩张而非平移。
2. **结构性替换**（编辑区越过固定代码段，如全选粘贴）且新文本仍与模板对齐 → 按模板代码段**从零提取**（回溯搜索，边界最准）。
3. 其余（区间被挤叠、结构残缺）→ 从零提取的贪婪兜底。

### 4.4 评分模型

逐点得分 = 基础分（精确/别名 100，部分正确 50，错误 0）+ 辅助扣分，按 `weight_score` 加权平均得总分；`GradeBugfix`（Bug 修复题型）为 测试通过 70 + 定位 20 + 质量 10 − 辅助扣分。总分映射 S/A/B/C/D。

### 4.5 顺序可互换分组（order_group）

训练点可标 `order_group`（如 HCL 块内先后无关的属性行）。判题时同组内做**双射匹配**：每个答案只分配给一个训练点，穷举（组 ≤7，阶乘上限 5040）取总分最高的分配。效果：行写反判对；两空填同样内容只能对一个点；未分组的题保持顺序敏感。

## 5. 关键设计决策

| 决策 | 理由 | 代价 |
|------|------|------|
| 题库 JSON 不存坐标，加载时重算 span | 题库纯手工可维护，改代码不用算偏移 | 训练点原文必须在代码中唯一可定位（用 `occurrence` 区分重复） |
| `net/http` 标准库、无框架 | MVP 依赖最小化、开箱即跑 | 中间件（鉴权/CORS/日志）手写，约 40 行 |
| SQLite 单连接 | 纯 Go 零 CGO 部署简单；单机训练工具写入低频 | 无并发写扩展；多实例部署需换 PostgreSQL（后续 Phase） |
| 非可执行栈走答案匹配 | K8s/Docker/SQL 真实执行需要沙箱（后续 Phase 3） | 这类题目考记忆而非实操 |
| 变体接口剥离答案 | 答案不进前端内存，L3/L4 提示点开才展示 | — |
| 增量区间追踪 + 三向路由 | 自由编辑下答案提取必须既准又稳（历次回归：空格分隔相邻填空、零宽重输、多删、整段粘贴） | 逻辑复杂，靠 29 个单测守护 |

## 6. 目录结构

```
CodeGym/
├── docs/
│   ├── design.md                 # 需求规格（v2.0）
│   ├── architecture.md           # 本文档
│   └── diagrams/                 # 交互式架构图（自包含 HTML）
│       ├── codegym-architecture.html
│       ├── codegym-sequence.html
│       └── codegym-dataflow.html
├── backend/
│   ├── cmd/server/               # 入口：flag 解析、seed、装配
│   ├── data/bank/*.json          # 题库（唯一数据源，19 文件 856 题）
│   └── internal/
│       ├── model/                # 数据结构与负重配置
│       ├── store/                # SQLite 7 表
│       ├── seed/                 # 题库装载与 span 重算
│       ├── variant/              # 负重随机留空
│       ├── engine/               # 子进程执行
│       ├── judge/                # 双路径判题与评分
│       ├── llm/                  # Ollama 客户端
│       └── api/                  # HTTP 路由与处理器（含 AI 出题）
├── frontend/src/
│   ├── App.tsx                   # hash 路由 + 布局
│   ├── pages/                    # 看板/训练/薄弱部位/题库管理/教练端
│   ├── components/               # CodeEditor/Sidebar/HintPanel/ScoreCard/WeightSelector
│   └── api/client.ts             # fetch 封装
└── scripts/                      # e2e_live.py 冒烟、audit_bank.py 题库审计
```

## 7. API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/exercises` | 训练列表（`?tech_stack=` 筛选，剥离答案字段） |
| GET | `/api/v1/exercises/:id` | 题目详情（含答案，教练用） |
| GET | `/api/v1/exercises/:id/variant` | 随机变体（掩码，不含答案） |
| POST | `/api/v1/execute` | 提交答案 → 执行 → 判题 → 评分 → 记录 |
| GET | `/api/v1/progress` · `/weak-points` · `/favorites` · `/environments` | 成长数据 |
| POST | `/api/v1/favorites` | 收藏开关 |
| POST/PUT/DELETE | `/api/v1/admin/exercises…` | 教练端题目 CRUD（需令牌） |
| GET/POST | `/api/v1/admin/export` · `/import` | 题库导出/导入 |
| * | `/api/v1/admin/categories` · `/environments` · `/llm-settings` · `/ai-generate` · `/stats` | 分类/环境/LLM 配置/AI 出题/统计 |

## 8. 安全边界与已知限制

- 定位**个人/内网训练工具**：学员端无鉴权（`X-User-Id` 可任意伪造），训练代码在服务器本地真实执行。对外部署必须启用 `-admin-token` 并置于带认证的反代之后。
- 非 Python/Go 题不真实执行（K8s/Docker 沙箱在后续 Phase 3）。
- 用户认证体系未建（Phase 4）；SQLite 单实例限制了水平扩展。
- `handlers.go` 中 `categoryHint2Map` 是 240+ 行的静态 L2 提示映射，随题库增长需考虑外置。

---

*本文档由代码分析生成（commit `fed1e5f`）。三张架构图的图源 JSON 位于 [diagrams/src/](diagrams/src/)，用 archify 校验并渲染，修改后可重新生成 HTML。*
