.PHONY: all test lint format build frontend-test frontend-lint setup dev dev-backend dev-frontend

setup:
	@echo "=> Bootstrapping local developer environment..."
	@bash setup.sh || pwsh -File setup.ps1 || powershell -File setup.ps1

all: format lint test build

clean: backend-clean frontend-clean
	@echo "=> Full clean complete."

rebuild: backend-rebuild frontend-rebuild
	@echo "=> Full rebuild complete."

ifeq ($(OS),Windows_NT)
# Run both backend and frontend on Windows using PowerShell.
dev:
	@echo "=> Starting Portify (backend + frontend) on Windows..."
	@powershell -ExecutionPolicy Bypass -NoProfile -Command "Start-Process go -ArgumentList 'run ./cmd/server'; cd frontend; npm run dev"
else
# Run both backend and frontend on Unix. The Go server is backgrounded; Ctrl+C kills both processes.
dev:
	@echo "=> Starting Portify (backend + frontend)..."
	@trap 'kill %1 2>/dev/null; exit' INT TERM; \
	go run ./cmd/server & \
	cd frontend && npm run dev
endif

dev-backend:
	@echo "=> Starting Go backend server..."
	go run ./cmd/server

dev-frontend:
	@echo "=> Starting Vite dev server..."
	cd frontend && npm run dev

# Backend (Go & Buf) Commands
ifeq ($(OS),Windows_NT)
backend-clean:
	@echo "=> Cleaning up Go build cache on Windows..."
	@powershell -ExecutionPolicy Bypass -NoProfile -Command "if (Test-Path 'server.exe') { Remove-Item -Force 'server.exe' }"
	go clean -cache
else
backend-clean:
	@echo "=> Cleaning up Go build cache..."
	rm -f server.exe
	go clean -cache
endif

backend-rebuild: backend-clean format lint test build

format:
	@echo "=> Formatting Go files and Protobuf..."
	buf format -w
	gofmt -s -w .

lint:
	@echo "=> Running Go Vet..."
	go vet ./...
	@echo "=> Linting Protobufs..."
	buf lint

test:
	@echo "=> Running Go backend tests..."
	go test ./...

build:
	@echo "=> Building Go binary..."
	go build -o server.exe ./cmd/server

# Frontend Commands
ifeq ($(OS),Windows_NT)
frontend-clean:
	@echo "=> Cleaning up Node Modules on Windows..."
	@powershell -ExecutionPolicy Bypass -NoProfile -Command "if (Test-Path 'frontend\node_modules') { Remove-Item -Recurse -Force 'frontend\node_modules' }"
	@powershell -ExecutionPolicy Bypass -NoProfile -Command "if (Test-Path 'frontend\package-lock.json') { Remove-Item -Force 'frontend\package-lock.json' }"
else
frontend-clean:
	@echo "=> Cleaning up Node Modules..."
	rm -rf frontend/node_modules frontend/package-lock.json
endif

frontend-install:
	@echo "=> Installing Frontend packages..."
	cd frontend && npm install

frontend-rebuild: frontend-clean frontend-install frontend-build

frontend-lint:
	@echo "=> Linting React frontend..."
	cd frontend && npm run lint

frontend-test:
	@echo "=> Running React frontend tests..."
	cd frontend && npm run test

frontend-build:
	@echo "=> Building React frontend..."
	cd frontend && npm run build
