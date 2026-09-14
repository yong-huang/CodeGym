# Changelog

本项目的所有重要变更记录于此。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added
- 数据结构系列题库（60 题，Python，6 个子分类各 10 题）：数组（双指针/二分/前缀和/差分/滑动窗口）、链表（反转/快慢指针/合并/环检测/两数相加）、栈与队列（括号匹配/逆波兰/最小栈/循环队列/单调栈与单调队列）、哈希表（手写哈希/两数之和/异位词分组/开放寻址/LRU 缓存）、树（遍历/BST 插入查找/验证/翻转/路径总和/最小堆/Trie）、图（邻接表/BFS/DFS/连通分量/最短路/拓扑排序/环检测/并查集/岛屿数量/Dijkstra），全部真实执行判题
- Go 类型系统题库（30 题，Go 子分类「类型系统」）：零值与类型推断/iota/定义类型与别名/显式转换/rune 与字节/类型断言与 type switch/结构体方法集/嵌入与标签/接口隐式实现/nil 接口陷阱/any/error/接口组合/泛型约束/指针与引用语义，真实 `go run` 执行判题
- Agentic AI 开发题库（50 题，Python 子分类「Agent 开发」）：消息与采样参数、工具 Schema/注册分派/异常兜底、Agent 循环与 ReAct 轨迹、滑动窗口/摘要/向量检索/RAG、结构化输出校验与重试、多智能体路由与 handoff、退避重试/成本核算/日志，全部真实执行判题
- 训练点 `order_group` 顺序可互换分组：同组判题做双射匹配，行序无关的配置行（如 HCL 块内属性）写反同样判对；重复答案仍只能匹配一个训练点
- SQL 基础题库（50 题）：SELECT/WHERE/排序分页/聚合分组/JOIN/子查询/DML/DDL，统一示例库出题
- Python 协程专项题库（20 题）：async/await/create_task/gather/超时异常/异步迭代器/Queue/Semaphore
- Python 多线程专项题库（20 题）：Thread/join/Lock·RLock/Semaphore/Event/Barrier/threading.local/queue.Queue/Timer/线程池
- BigQuery 题库（61 题）拆为两个子库：「SQL 基础」（35 题通用语法）与「BigQuery 特有」（26 题：数组/STRUCT/UNNEST、JSON、窗口·QUALIFY、分区聚类、通配表、INFORMATION_SCHEMA、脚本变量、采样、时间旅行）
- DevOps 题库（140 题）拆为四个子库：「terraform」（30）、「ansible」（30）及并入的「docker」（50）、「prometheus」（30）
- AWS 题库（180 题）六个子库各 30 题：ec2/iam/eks/s3/vpc/rds
- 教练端接口令牌鉴权：`-admin-token` 或 `CODEGYM_ADMIN_TOKEN` 启用后所有 `/api/v1/admin/*` 需携带令牌
- Docker 部署：`Dockerfile.backend` / `Dockerfile.frontend` / `docker-compose.yml`
- GitHub Actions CI：后端 -race 测试、前端测试与构建、题库审计

### Changed
- Bug 修复题型训练点提示改用题目级四级辅助（原按类别生成的提示与填空内容错位）
- channel 类训练点 L2 提示补充 comma-ok 说明；新增 map_write/critical_section 等类别映射

### Fixed
- 执行失败路径 `fill_results` 返回 null 导致前端评分卡崩溃（现恒为空数组）
- 前端填空答案提取被后续代码段文字截断的回归问题
- 编辑器填空区间追踪三连修：删除答案后重新输入不被识别、退格多删破坏代码结构后答案错位、单填空整段粘贴吞并固定代码
- 编辑器超长行被裁剪：按内容撑开宽度出横向滚动，行号栏 sticky 钉在可视区左缘；同时修复超过 560px 的长代码纵向滚动后高亮与行号消失

## [0.2.0] - 2026-08

### Added
- 题库全面 JSON 化到 `backend/data/bank/`，不改代码即可增删改题目
- Kubernetes 调度专项（30 题）、OpenShift 对比专项（30 题）、Prometheus 基础（30 题）
- Go 并发专项（50 题，真实 go run 判题）+ 自动补训练点机制
- 教练端 AI 出题（Ollama）+ Ollama 设置页面化 + 执行环境管理
- 侧栏收藏视图、按代码位置排列评分表

## [0.1.0] - 2026-08

### Added
- CodeGym MVP：填空-运行-验证闭环、负重分级、四级辅助、评分评级
- Python / Go / JavaScript / Docker / OpenShift / Kubernetes 首批题库
- 训练进度、薄弱部位、收藏、导入导出、教练端 CRUD
