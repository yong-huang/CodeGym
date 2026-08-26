.PHONY: backend frontend dev test test-backend test-frontend audit build clean

# 一键启动前后端（前台运行，Ctrl+C 停止）
dev:
	cd backend && go run ./cmd/server -addr :8090 & \
	cd frontend && npm run dev

backend:
	cd backend && go run ./cmd/server -addr :8090

frontend:
	cd frontend && npm run dev

# 后端全量测试（含竞态检测）
test-backend:
	cd backend && go test ./... -race -count=1

# 前端单测 + 生产构建
test-frontend:
	cd frontend && npm test && npm run build

# 题库审计：结构 / span / 提示 / 可执行题真实运行输出
audit:
	python3 scripts/audit_bank.py

# 全部测试
test: test-backend test-frontend audit

build:
	cd backend && go build ./...
	cd frontend && npm run build

clean:
	rm -rf frontend/dist backend/bin
	find . -name "*.tsbuildinfo" -not -path "*/node_modules/*" -delete
