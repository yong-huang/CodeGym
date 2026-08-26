# Changelog

本项目的所有重要变更记录于此。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added
- SQL 基础题库（50 题）：SELECT/WHERE/排序分页/聚合分组/JOIN/子查询/DML/DDL，统一示例库出题
- Python 协程专项题库（20 题）：async/await/create_task/gather/超时异常/异步迭代器/Queue/Semaphore
- 教练端接口令牌鉴权：`-admin-token` 或 `CODEGYM_ADMIN_TOKEN` 启用后所有 `/api/v1/admin/*` 需携带令牌
- Docker 部署：`Dockerfile.backend` / `Dockerfile.frontend` / `docker-compose.yml`
- GitHub Actions CI：后端 -race 测试、前端测试与构建、题库审计

### Changed
- Bug 修复题型训练点提示改用题目级四级辅助（原按类别生成的提示与填空内容错位）
- channel 类训练点 L2 提示补充 comma-ok 说明；新增 map_write/critical_section 等类别映射

### Fixed
- 执行失败路径 `fill_results` 返回 null 导致前端评分卡崩溃（现恒为空数组）
- 前端填空答案提取被后续代码段文字截断的回归问题

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
