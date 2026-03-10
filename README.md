# 🎵 Portify

*The Universal Playlist Converter*

**Portify** is an open-source tool aimed at breaking down the walled gardens of music streaming. It allows users to seamlessly convert and synchronize their carefully curated playlists across multiple music platforms—moving your music freely between services like Spotify, YouTube Music, Apple Music, and more!

> ⚠️ **Note:** Portify is currently a **Work in Progress (WIP)**. The core backend architecture is active, but the frontend UI and additional streaming providers are still being developed.

> ✨ **Built with AI:** This entire application was "vibe coded" from scratch with the help of **Antigravity** and the **Gemini 3.1 Pro High** model. 

## 🚀 Current Status

- [x] **Core Canonical Model:** Protobuf-based universal data structures for seamless translation.
- [x] **Spotify Adapter (Source):** Full support for fetching tracks from public and private Spotify playlists.
- [x] **YouTube Adapter (Destination):** Full support for algorithmically matching tracks and generating YouTube playlists.
- [ ] **React Frontend:** Web UI for user login, provider selection, and conversion triggers.
- [ ] **Automated Testing:** Testcontainer integration for E2E flow validation.

## 🛠️ Technology Stack

* **Core Engine:** Go (Golang)
* **API Framework:** ConnectRPC & Protocol Buffers
* **Frontend:** React / Vite (Coming Soon)
* **Authentication:** OAuth 2.0 

## 🏃 Getting Started & Build Instructions

### Prerequisites
- [Go 1.25+](https://go.dev/doc/install)
- [Node.js (v18+)](https://nodejs.org/) & `npm`
- [Buf CLI](https://buf.build/docs/installation) (for compiling Protocol Buffers)

### 1. Generating Protobufs
We use `buf` to generate Go models and TypeScript ConnectRPC client code from our `.proto` definitions.
```bash
# Run this from the root of the project every time you change a .proto file
buf generate
```

### 2. Starting the Backend (Go)
Fetch dependencies and start the ConnectRPC Go server (runs on `localhost:8080`).
```bash
go mod tidy
go run ./cmd/server
```

### 3. Starting the Frontend (React + Vite)
In a new terminal window, start the frontend development server:
```bash
cd frontend
npm install
npm run dev
```
The React app will be available at `http://localhost:5173`. It acts as a proxy to tunnel `/converter.v1.ConverterService/*` requests to the Go backend.

### 4. Running Tests (E2E Validation)
Run all tests, including our E2E flow tests utilizing Testcontainers.
```bash
go test ./... -v
```

### 5. Pushing Changes
Ensure you generate protobufs cleanly, format your code, and run tests before pushing:
```bash
buf format -w
go fmt ./...
go test ./...
git add .
git commit -m "Your descriptive commit message"
git push
```
