# 🎵 Portify

*The Universal Playlist Converter*

[![Build Status](https://github.com/debalin/portify/actions/workflows/ci.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/debalin/portify)](https://goreportcard.com/report/github.com/debalin/portify)
[![CodeQL Analysis](https://github.com/debalin/portify/actions/workflows/codeql.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/codeql.yml)

**Portify** is an open-source tool aimed at breaking down the walled gardens of music streaming. It allows users to seamlessly convert and synchronize their carefully curated playlists across multiple music platforms—moving your music freely between services like Spotify, YouTube Music, Apple Music, and more!

> ⚠️ **Note:** Portify is currently a **Work in Progress (WIP)**. The core backend architecture is active, but the frontend UI and additional streaming providers are still being developed.

> ✨ **Built with AI:** This entire application was "vibe coded" from scratch with the help of **Antigravity** and the **Gemini 3.1 Pro High** model. 

## 🚀 Current Status

- [x] **Core Canonical Model:** Protobuf-based universal data structures for seamless translation.
- [x] **Spotify Adapter (Source):** Full support for fetching tracks from public and private Spotify playlists.
- [x] **YouTube Adapter (Destination):** Full support for algorithmically matching tracks and generating YouTube playlists.
- [x] **React Frontend:** Web UI for user login, provider selection, and server-streaming conversion triggers.
- [ ] **Automated Testing:** Testcontainer integration for E2E flow validation.

## ✨ Key Features

* **Stateless Authentication:** Completely database-free architecture. OAuth 2.0 flows are securely brokered through Go and dynamically pinned to your browser's lightweight `sessionStorage`.
* **Streaming Progress Tracking:** ConnectRPC leverages Server-Sent Events (SSE) to pipe granular track-by-track conversion progress dynamically into a React progress bar UI.
* **Dynamic Playlist Generation:** Seamlessly `Create New` distinct destination playlists, or dynamically fetch your existing provider playlists to `Append` tracks onto the end of them without overwriting!
* **Robust Error Handling:** Granular failure tracking precisely flags unmatched songs inside a collapsible DOM element to prevent massive playlists from locking up the browser, whilst seamlessly invalidating expired OAuth credentials quietly under the hood.

## 🛠️ Technology Stack

* **Core Engine:** Go (Golang)
* **API Framework:** ConnectRPC & Protocol Buffers
* **Frontend:** React / Vite (Coming Soon)
* **Authentication:** OAuth 2.0 

## 🔐 Environment Configuration (.env)

The application relies on `.env` files to securely manage credentials and toggle UI features.

### Backend (`/.env`)
Create a `.env` file in the root repository directory to supply your OAuth Developer keys:
```env
SPOTIFY_ID="your_spotify_client_id"
SPOTIFY_SECRET="your_spotify_client_secret"
YOUTUBE_ID="your_google_cloud_client_id"
YOUTUBE_SECRET="your_google_cloud_client_secret"
FRONTEND_URL="http://127.0.0.1:5175/"
```
> **Important:** The `FRONTEND_URL` must exactly match the authorized Redirect URI you configured in the Spotify and Google Cloud developer dashboards (including the trailing slash).

### Frontend (`/frontend/.env`)
Create a separate `.env` file inside the `/frontend` directory to toggle Vite-specific features:
```env
# Set to 'true' to display a persistent red overlay printing raw App.tsx memory and sessionStorage states.
VITE_SHOW_DEBUG_PANEL=false
```

## 🏃 Getting Started & Build Instructions

### Prerequisites
- [Go 1.25+](https://go.dev/doc/install)
- [Node.js (v18+)](https://nodejs.org/) & `npm`

### 1. Developer Environment Bootstrapper (Recommended)
You can entirely automate installing all required Go developer tools (like `buf`), configuring the `pre-commit` git hooks, scaffolding out your local `.env` skeleton files, and fetching all frontend packages with a single cross-platform make command:
```bash
make setup
```

### 2. Generating Protobufs
We use `buf` to generate Go models and TypeScript ConnectRPC client code from our `.proto` definitions.
```bash
buf generate
```

### 3. Starting the Backend (Go)
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
