# 贡献指南

感谢关注 CodeGym！欢迎以任何形式参与贡献：修 Bug、补题、加技术栈、改进体验。

## 开发环境

```bash
# 后端（Go 1.26+，SQLite 纯 Go 驱动无需 cgo）
cd backend && go run ./cmd/server -addr :8090

# 前端（Node 22+）
cd frontend && npm install && npm run dev   # http://localhost:5190

# 测试（提交前请确保全部通过）
cd backend && go test ./... -race -count=1
cd frontend && npm test && npm run build
python3 scripts/audit_bank.py               # 题库结构 + 真实执行校验
```

## 提交题目

内置题库全部是 `backend/data/bank/*.json` 文件，改 JSON 即可增删改题目，无需动 Go 代码。

出题规范：

1. **可执行题**（python/go 的 fill 题）必须有确定性的 `expected_output`——判题会真实运行代码并比对 stdout；
2. **训练点**（points）的 `text` 必须是 `full_code` 中的原文（span 由后端按文本定位），同一题内 id 不重复；
3. 每题提供四级提示（level_1 点拨 → level_4 完整答案）与简短解析；
4. 提交前跑 `python3 scripts/audit_bank.py`，0 ERROR 才能合并。

新增技术栈：在 `data/bank/` 新建 `<stack>.json`；若该栈不本地执行（答案匹配判题），无需改代码即可生效。

## 提交约定

- Commit message 用中文一行主题 + 明细列表（参考既有历史），例如：`题库: 新增 SQL 基础题库 50 题`
- 一个 PR 聚焦一件事；涉及题库的 PR 请在描述中附 audit 输出摘要
- 行为变更请同步更新 README 与 e2e 断言

## 报告问题

提交 issue 时请附：复现步骤、相关题目 id（如 `sql-basics-029`）、后端日志片段。
