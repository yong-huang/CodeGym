# CodeGym · 交互式代码训练平台

> 练出来的，才是自己的。

![CI](https://github.com/hyhit/CodeGym/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue)
![Data License](https://img.shields.io/badge/Data-CC%20BY--SA%204.0-lightgrey)

依据 `docs/design.md` 实现的 **Phase 1+2 可运行 MVP**：一个「填空-运行-验证」的代码训练闭环。
学员选择训练动作 → 选择负重级别 → 随机留空生成训练变体 → 填写训练点 → 运行验证 → 即时评分（S/A/B/C/D）→ 记录薄弱部位与成长进度。

## 界面预览

| 首页看板（含题库目录树） | 训练页（填空 + 负重 + 辅助提示） | 判题结果（真实执行 + 评分卡） |
|:---:|:---:|:---:|
| ![首页](docs/screenshots/home.png) | ![训练页](docs/screenshots/training.png) | ![判题结果](docs/screenshots/result.png) |

## 功能总览

- 📚 **训练计划**：浏览 / 按技术栈筛选训练动作（Python / Go / JavaScript / Kubernetes / …）
- 🐍 **Python 题库**：50 道基础题（语法/f-string/流程/列表推导/函数/递归/异常/类……），真实 `python3` 执行判题
- ⚡ **Python 协程题库**：20 道专项题（async/await/create_task/gather/超时与异常/异步迭代器/Queue/Semaphore……），真实 `asyncio` 执行判题
- 🧵 **Python 多线程题库**：20 道专项题（Thread/join/共享状态/Lock·RLock/Semaphore/Event/Barrier/threading.local/queue.Queue/Timer/线程池……），全部按确定性输出设计并真实执行判题
- 🤖 **Agentic AI 开发题库**：50 道 Python 题（消息与采样参数、工具 Schema 与注册分派、Agent 循环与 ReAct 轨迹、滑动窗口/摘要/向量检索/RAG、结构化输出校验、多智能体路由与 handoff、退避重试/成本核算/日志），真实执行判题
- 🐹 **Go 题库**：50 道基础题（语法/函数/切片/map/结构体/接口/并发/泛型……），真实 `go run` 执行判题
- 🔷 **Go 类型系统题库**：30 道专项题（零值与类型推断/iota/定义类型与别名/转换与 rune/类型断言与 type switch/结构体与方法集/嵌入与标签/接口隐式实现与 nil 陷阱/any/error/接口组合/泛型约束/指针语义），真实 `go run` 执行判题
- 🧱 **数据结构系列题库**（60 题，Python 真实执行判题，6 个子分类各 10 题）：数组（双指针/二分/前缀和/差分/滑动窗口）、链表（反转/快慢指针/合并/环检测/两数相加）、栈与队列（括号匹配/逆波兰/最小栈/循环队列/单调栈与单调队列）、哈希表（手写哈希/两数之和/异位词分组/开放寻址/LRU）、树（遍历/BST/翻转/堆/Trie）、图（BFS/DFS/拓扑排序/并查集/岛屿数量/Dijkstra）
- ☸️ **Kubernetes 题库**：50 道基础题（kubectl 命令补全 + YAML 配置补全），支持别名答案（如 `-A` = `--all-namespaces`）
- 🗄️ **SQL 题库**：50 道基础题（SELECT/WHERE 排序分页/聚合分组/多表 JOIN/子查询/DML/DDL 与约束……），围绕统一的示例库（员工/部门/商品/客户/订单）出题
- 📊 **BigQuery 题库**：61 道题，拆两个子库——「SQL 基础」（35 题：通用 SQL 语法）与「BigQuery 特有」（26 题：数组·STRUCT·UNNEST/JSON/窗口·QUALIFY/分区聚类/通配表/INFORMATION_SCHEMA/脚本变量/采样/时间旅行等 BQ 专属），围绕 sales 数据集出题
- 🚀 **DevOps 题库**：140 道题，拆四个子库——「terraform」（30）、「ansible」（30）、「docker」（50：命令 + Dockerfile/compose）与「prometheus」（30：promtool/HTTP API/PromQL/告警配置，答案匹配判题）
- ☁️ **AWS 题库**：180 道题，六个子库各 30 题——ec2（实例/镜像/EBS/安全组/弹性 IP）、iam（用户·组·角色/策略 JSON/信任与边界/临时凭证）、eks（集群/节点组/Fargate/OIDC·kubectl）、s3（桶/对象/存储级别/生命周期/加密/桶策略）、vpc（子网/路由/IGW·NAT/对等·端点/流日志）、rds（建库/快照恢复/只读副本/Multi-AZ/Aurora/参数组）
- ⭐ **收藏**：给需要加强的题目点 ★ 收藏，训练页与首页可随时取消，支持「只看收藏」筛选，有空多练练
- 🎯 **负重分级**：🌱入门 → 🌿简单 → 🌳中等 → 🔥困难 → 💎专家，随机留空数量随级别变化（设计 §5）
- 🧑‍💻 **四种题型**：填空补全、命令补全、配置补全、Bug 修复（设计 §4）
- 💡 **四级辅助保护**：轻提示 → 中提示 → 重提示 → 显答案，逐级扣分（设计 §6）
- ▶️ **运行验证**：Python / Go 训练点本地子进程执行（Go 通过临时 module `go run`），自动判题（设计 §7/§8）
- 📦 **题库导入 / 导出**：教练端一键导出整个题库为 JSON，可粘贴导入迁移/共享
- 📊 **训练成绩**：逐训练点评分、总分、评级、薄弱分析（设计 §7.5）
- 📝 **薄弱部位**：错题自动进错题本，支持一键重新加练
- 🔥 **成长追踪**：打卡连续天数、累计作答、完成进度（设计 §13）
- 🧑‍🏫 **教练端**：训练动作 JSON 创建 / 更新 / 删除，平台统计

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.26 · net/http（标准库） |
| 数据库 | SQLite（modernc.org/sqlite，纯 Go） |
| 执行引擎 | 本地 Python 子进程（`PYTHON_BIN` 可配置，MVP 阶段） |
| 前端 | React 18 + TypeScript + Vite |
| 存储结构 | 与设计文档 §11 的 exercises / attempt_records / user_progress 表对应 |

> 设计文档推荐的 Docker/K8s 沙箱、PostgreSQL、Ollama 出题为后续 Phase，MVP 使用本地子进程与 SQLite 保证开箱即跑。

## 快速开始

### 方式一：Docker Compose（推荐）

```bash
docker compose up -d --build          # 前端 http://localhost:8088
# 生产环境务必设置教练端令牌：
CODEGYM_ADMIN_TOKEN=$(openssl rand -hex 16) docker compose up -d --build
```

### 方式二：本地开发

#### 1. 启动后端

```bash
cd backend
go run ./cmd/server -addr :8090        # 首次启动自动建库并 seed data/bank 下的全部内置题库（996 题）
```

可选参数：`-db data/codegym.db` 指定数据库文件；`PYTHON_BIN` 指定 Python 解释器；
`-admin-token <token>`（或环境变量 `CODEGYM_ADMIN_TOKEN`）为教练端接口启用鉴权。

#### 2. 启动前端

```bash
cd frontend
npm install
npm run dev                             # http://localhost:5190 （/api 已代理到 :8090）
```

打开 http://localhost:5190 即可开始训练。

## ⚠️ 安全提示

CodeGym 的定位是**个人/内网训练工具**，默认配置**未启用任何鉴权与沙箱隔离**：

- 教练端接口（创建/删除题目、导入导出、AI 出题）可被任意调用；
- 训练代码会在服务器本地以子进程真实执行（Python 直接运行、Go 经 `go run` 编译执行）。

**请勿将未加防护的实例暴露到公网。** 对外部署时：

1. 用 `-admin-token` / `CODEGYM_ADMIN_TOKEN` 启用教练端令牌（前端「教练端」页可保存令牌）；
2. 置于带认证的反向代理之后，限制来源 IP；
3. 如需强隔离，把后端放进容器并限制 CPU/内存/网络（K8s 沙箱在后续路线中）。


### 3. 运行测试

```bash
# 单元 + 集成测试（judge / variant / 全量 API 端到端，含 -race 竞态检测）
cd backend && go test ./... -race -count=1

# 前端类型检查、编辑区答案提取/渲染单测、生产构建
cd frontend && npm test && npm run build

# 实时端到端冒烟测试（需先启动后端 :8090 与前端 :5190）
python3 scripts/e2e_live.py
```

`backend/internal/api/e2e_test.go` 内置真实 SQLite + 真实 Python 子进程的完整接口测试
（列表/变体/执行/判题/进度/薄弱部位/教练 CRUD/导入/统计/超时/404）。
`frontend/src/components/CodeEditor.test.tsx` 覆盖 VSCode 风格编辑器的答案提取与渲染。
`scripts/e2e_live.py` 则通过前端代理对运行中的服务做全链路冒烟验证（65 项检查）。

## API 一览（设计 §12）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/exercises` | 训练列表（`?tech_stack=` 筛选） |
| `GET` | `/api/v1/exercises/:id` | 训练动作详情（含答案字段，教练用） |
| `GET` | `/api/v1/exercises/:id/variant?level=easy` | 随机变体（训练点已留空，不含答案） |
| `POST` | `/api/v1/execute` | 提交答案 → 执行 → 判题 → 评分 |
| `GET` | `/api/v1/progress` | 学员成长进度 + 最近作答 |
| `GET` | `/api/v1/weak-points` | 薄弱部位列表 |
| `GET` | `/api/v1/environments` | 可用执行环境 |
| `POST` | `/api/v1/admin/exercises` | 创建训练动作（教练，需令牌） |
| `PUT` | `/api/v1/admin/exercises/:id` | 更新训练动作（教练，需令牌） |
| `DELETE` | `/api/v1/admin/exercises/:id` | 删除训练动作（教练，需令牌） |
| `GET` | `/api/v1/admin/export` | 导出整个题库（JSON，教练，需令牌） |
| `POST` | `/api/v1/admin/import` | 批量导入训练动作（教练，需令牌） |
| `GET` | `/api/v1/admin/stats` | 平台统计（教练，需令牌） |

> 教练端接口在启用 `-admin-token` 后需要携带 `X-Admin-Token: <token>` 或 `Authorization: Bearer <token>`；未启用时开放（仅限本地开发）。

## 目录结构

```
CodeGym/
├── docs/design.md              # 设计文档（需求来源）
├── backend/
│   ├── cmd/server/             # 入口
│   └── internal/
│       ├── model/              # 数据模型（设计 §11）
│       ├── store/              # SQLite 存储
│       ├── seed/               # 内置训练动作
│       ├── variant/            # 负重随机留空引擎（设计 §5）
│       ├── engine/             # 执行引擎（设计 §8）
│       ├── judge/              # 判题评分（设计 §7）
│       └── api/                # HTTP 接口（设计 §12）
└── frontend/
    └── src/
        ├── components/         # 代码编辑器 / 辅助面板 / 评分卡 / 负重选择
        └── pages/              # 看板 / 训练 / 薄弱部位 / 教练端
```

## 后续路线（对应设计 §16）

- Phase 3：Kubernetes / Docker 执行引擎、真实环境配置管理
- Phase 4：用户认证、权限、训练报告
- Phase 5：JSON/YAML 批量导入导出、AI 出题（Ollama）、数据仪表盘
- Phase 6：生产部署（Dockerfile、监控、安全加固）

## 贡献

欢迎提 Issue 与 PR！出题、修 Bug、加技术栈都可以，见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

- 代码：[MIT](LICENSE)
- 题库内容（`backend/data/bank/*.json`）：[CC BY-SA 4.0](LICENSE-DATA)
