## 变更说明

（做了什么、为什么做）

## 变更类型

- [ ] 新增题库 / 题目
- [ ] Bug 修复
- [ ] 新功能
- [ ] 重构 / 文档

## 自查清单

- [ ] `cd backend && go test ./... -race -count=1` 通过
- [ ] `cd frontend && npm test && npm run build` 通过（涉及前端时）
- [ ] `python3 scripts/audit_bank.py` 0 ERROR（涉及题库时）
- [ ] 行为变更已同步 README / e2e 断言
