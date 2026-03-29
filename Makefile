.PHONY: all test lint format build frontend-test frontend-lint setup

setup:
	@echo "=> Bootstrapping local developer environment..."
	@bash setup.sh || pwsh -File setup.ps1 || powershell -File setup.ps1

all: format lint test build

# Backend (Go & Buf) Commands
format:
	@echo "=> Formatting Go files and Protobuf..."
	buf format -w
	gofmt -s -w .

lint:
	@echo "=> Running golangci-lint..."
	golangci-lint run
	@echo "=> Linting Protobufs..."
	buf lint

test:
	@echo "=> Running Go backend tests with coverage..."
	go test -coverprofile=coverage.out ./...

build:
	@echo "=> Building Go binary..."
	go build -o server.exe ./cmd/server

# Frontend Commands
frontend-install:
	@echo "=> Installing Frontend packages..."
	cd frontend && npm install

frontend-lint:
	@echo "=> Linting React frontend..."
	cd frontend && npm run lint

frontend-test:
	@echo "=> Running React frontend tests..."
	cd frontend && npm run test

frontend-build:
	@echo "=> Building React frontend..."
	cd frontend && npm run build
