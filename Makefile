.PHONY: backend frontend test build dev

backend:
	cd backend && go run ./cmd/server -addr :8090

frontend:
	cd frontend && npm run dev

test:
	cd backend && go test ./...

build:
	cd backend && go build ./...
	cd frontend && npm run build

# 一键启动前后端（前台运行，Ctrl+C 停止）
dev:
	cd backend && go run ./cmd/server -addr :8090 & \
	cd frontend && npm run dev
