#!/usr/bin/env bash
# macOS / Linux / WSL Bootstrapper

echo "=> Checking dependencies..."
if ! command -v go &> /dev/null; then
    echo "❌ ERROR: Go is not installed. Please download it from https://go.dev/dl/ and run this script again."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ ERROR: Node.js/npm is not installed. Please download it from https://nodejs.org/ and run this script again."
    exit 1
fi

echo "[1/4] Configuring strict Git Hooks..."
git config core.hooksPath .githooks

echo "[2/4] Scaffolding environment files..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "  -> Created template .env file in root"
fi
if [ ! -f frontend/.env ]; then
    echo "VITE_SHOW_DEBUG_PANEL=false" > frontend/.env
    echo "  -> Created template frontend/.env"
fi

echo "[3/4] Installing Go Developer tooling..."
go install github.com/bufbuild/buf/cmd/buf@v1.40.0
go mod tidy

echo "[4/4] Installing Node.js frontend dependencies..."
cd frontend && npm install

echo "✅ Bootstrap Complete! Be sure to fill out your .env file credentials!"
