# 🎵 Portify

*The Universal Playlist Converter*

[![CI/CD Pipeline](https://github.com/debalin/portify/actions/workflows/ci.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/debalin/portify)](https://goreportcard.com/report/github.com/debalin/portify)
[![CodeQL Analysis](https://github.com/debalin/portify/actions/workflows/codeql.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/codeql.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Portify** is an open-source tool that breaks down the walled gardens of music streaming. It allows users to seamlessly convert and synchronize their carefully curated playlists across multiple music platforms — moving your music freely between services like Spotify, YouTube Music, and more.

> ✨ **Built with AI:** This entire application was vibe-coded from scratch with the help of **Antigravity** and the **Gemini 3.1 Pro** model.

## ✨ Key Features

* **Real-Time Streaming Conversion:** ConnectRPC leverages Server-Sent Events (SSE) to stream granular track-by-track conversion progress directly into a live React progress bar.
* **Stateless Authentication:** Completely database-free. OAuth 2.0 flows are securely brokered through Go and pinned to the browser's lightweight `sessionStorage`.
* **Dynamic Playlist Generation:** `Create New` destination playlists, or fetch your existing provider playlists to `Append` tracks without overwriting.
* **Robust Error Handling:** Unmatched songs are precisely flagged in a collapsible UI element, and expired OAuth credentials are silently invalidated under the hood.

## 🚀 Current Status

- [x] **Core Canonical Model:** Protobuf-based universal data structures for cross-platform track translation.
- [x] **Spotify Adapter (Source):** Full support for fetching tracks from public and private Spotify playlists.
- [x] **YouTube Adapter (Destination):** Algorithmic track matching and YouTube playlist generation.
- [x] **React Frontend:** Web UI with OAuth login, provider selection, playlist browsing, and streaming conversion progress.
- [x] **CI/CD Pipeline:** Automated GitHub Actions for linting, testing, and build verification on every push and PR.
- [x] **Security Scanning:** CodeQL Advanced Security analysis and Dependabot dependency monitoring.
- [ ] **Additional Providers:** Apple Music, Tidal, Amazon Music, etc.

## 🛠️ Technology Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25, ConnectRPC, Protocol Buffers |
| **Frontend** | React 19, Vite 7, TypeScript 5.9, Vitest |
| **Authentication** | OAuth 2.0 (Spotify Web API, Google/YouTube Data API v3) |
| **Linting** | `go vet` + `gofmt` (Backend), ESLint + TypeScript-ESLint (Frontend) |
| **CI/CD** | GitHub Actions, CodeQL, Dependabot |
| **Protobuf Tooling** | Buf (generation, linting, formatting) |

## 📁 Project Structure

```
portify/
├── cmd/
│   ├── server/          # Main Go backend entrypoint
│   ├── testspotify/     # Spotify adapter integration test harness
│   └── testyoutube/     # YouTube adapter integration test harness
├── internal/
│   ├── adapters/
│   │   ├── spotify/     # Spotify API adapter (source)
│   │   ├── youtube/     # YouTube API adapter (destination)
│   │   └── mock/        # Mock adapter for unit testing
│   ├── domain/          # Core provider interface & canonical model
│   └── server/          # ConnectRPC service handler (converter.go)
├── proto/               # Protobuf service & model definitions
├── gen/                 # Auto-generated Go & TypeScript code (buf generate)
├── frontend/            # React + Vite + TypeScript SPA
│   └── src/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml       # CI/CD Pipeline (Go + React)
│   │   └── codeql.yml   # CodeQL security scanning
│   └── dependabot.yml   # Automated dependency updates
├── .githooks/
│   └── pre-commit       # Local pre-commit hook (format + lint)
├── Makefile             # Developer commands (format, lint, test, build)
├── setup.sh             # Bootstrap script (macOS/Linux/WSL)
├── setup.ps1            # Bootstrap script (Windows PowerShell)
└── .env.example         # Template for backend OAuth credentials
```

## 🔐 Environment Configuration

The application uses `.env` files for credentials and feature toggles.

### Backend (`/.env`)

Create a `.env` file in the project root with your OAuth developer keys:

```env
SPOTIFY_ID="your_spotify_client_id"
SPOTIFY_SECRET="your_spotify_client_secret"
YOUTUBE_ID="your_google_cloud_client_id"
YOUTUBE_SECRET="your_google_cloud_client_secret"
FRONTEND_URL="http://127.0.0.1:5175/"
```

> **Important:** The `FRONTEND_URL` must exactly match the authorized Redirect URI configured in both the Spotify and Google Cloud developer consoles (including the trailing slash).

### Frontend (`/frontend/.env`)

```env
# Set to 'true' to display a debug overlay showing raw state and sessionStorage.
VITE_SHOW_DEBUG_PANEL=false
```

## 🏃 Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/doc/install)
- [Node.js (v18+)](https://nodejs.org/) & `npm`

### 1. Quick Setup (Recommended)

The bootstrap script installs developer tools (`buf`), configures git hooks, scaffolds `.env` files, and installs frontend dependencies:

```bash
make setup
```

### 2. Generate Protobufs

Generate Go server stubs and TypeScript ConnectRPC client code from `.proto` definitions:

```bash
buf generate
```

### 3. Start the Backend

```bash
go run ./cmd/server
```

The ConnectRPC server starts on `http://localhost:8080`.

### 4. Start the Frontend

In a separate terminal:

```bash
cd frontend
npm run dev
```

The React app is available at `http://127.0.0.1:5175`. Vite proxies all `/converter.v1.ConverterService/*` requests to the Go backend automatically.

## 🧪 Testing & Quality

### Makefile Commands

| Command | Description |
|---|---|
| `make setup` | Bootstrap the full developer environment |
| `make format` | Format Go files (`gofmt -s`) and Protobufs (`buf format`) |
| `make lint` | Run `go vet` and `buf lint` |
| `make test` | Run Go backend unit tests with coverage |
| `make build` | Compile the Go server binary |
| `make frontend-lint` | Run ESLint on the React frontend |
| `make frontend-test` | Run Vitest unit tests |
| `make frontend-build` | Production build of the React app |
| `make all` | Run format → lint → test → build |

### CI/CD Pipeline

Every push and pull request to `master` triggers two parallel GitHub Actions jobs:

1. **Go Backend & Lint** — Sets up Go 1.25, runs `buf lint`, `gofmt` formatting check, `go vet`, and `go test` with coverage.
2. **React Frontend & Lint** — Sets up Node.js 20, runs `npm ci`, ESLint, Vitest, and a production build compilation check.

### Pre-Commit Hook

The `make setup` script configures a local git hook (`.githooks/pre-commit`) that runs formatting, `go vet`, `buf lint`, and frontend ESLint checks before every commit. Use `--no-verify` to skip when needed.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
