Write-Host "=> Checking dependencies..." -ForegroundColor Cyan
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "❌ ERROR: Go is not installed. Please download it from https://go.dev/dl/ and run this script again." -ForegroundColor Red
    exit 1
}

if (-not (Get-Command "npm" -ErrorAction SilentlyContinue)) {
    Write-Host "❌ ERROR: Node.js/npm is not installed. Please download it from https://nodejs.org/ and run this script again." -ForegroundColor Red
    exit 1
}

Write-Host "[1/4] Configuring strict Git Hooks..." -ForegroundColor Cyan
git config core.hooksPath .githooks

Write-Host "[2/4] Scaffolding environment files..." -ForegroundColor Cyan
if (!(Test-Path -Path ".env")) {
    Copy-Item ".env.example" -Destination ".env"
    Write-Host "  -> Created template .env file in root" -ForegroundColor Green
}
if (!(Test-Path -Path "frontend/.env")) {
    Set-Content -Path "frontend/.env" -Value "VITE_SHOW_DEBUG_PANEL=false"
    Write-Host "  -> Created template frontend/.env" -ForegroundColor Green
}

Write-Host "[3/4] Installing Go Developer tooling..." -ForegroundColor Cyan
go install github.com/bufbuild/buf/cmd/buf@v1.40.0
go mod tidy

Write-Host "[4/4] Installing Node.js frontend dependencies..." -ForegroundColor Cyan
Push-Location frontend
npm install
Pop-Location

Write-Host "✅ Bootstrap Complete! Be sure to fill out your .env file credentials!" -ForegroundColor Green
