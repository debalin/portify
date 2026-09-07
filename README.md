# 🎵 Portify

*The Universal Playlist Converter*

[![CI/CD Pipeline](https://github.com/debalin/portify/actions/workflows/ci.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/ci.yml)
[![CodeQL Analysis](https://github.com/debalin/portify/actions/workflows/codeql.yml/badge.svg)](https://github.com/debalin/portify/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/debalin/portify/graph/badge.svg)](https://codecov.io/gh/debalin/portify)
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
- [x] **Spotify Adapter:** Full bidirectional support for fetching tracks from Spotify and generating playlists.
- [x] **YouTube Adapter:** Full bidirectional support for fetching tracks and generating YouTube playlists with algorithmic track matching.
- [x] **React Frontend:** Web UI with OAuth login, provider selection, playlist browsing, and streaming conversion progress.
- [x] **CI/CD Pipeline:** Automated GitHub Actions for linting, testing, and build verification on every push and PR.
- [x] **Security Scanning:** CodeQL Advanced Security analysis and Dependabot dependency monitoring.
- [x] **Tidal Adapter:** Full support for converting to and from Tidal via Open API v2.
- [ ] **Additional Providers:** Apple Music, Amazon Music, SoundCloud, etc.

## 🎧 Supported Platforms

Portify currently supports seamless conversion between the following music streaming platforms:
- **Spotify**: Full support (Source & Destination)
- **YouTube Music**: Full support (Source & Destination)
- **Tidal**: Full support (Source & Destination)

*Note: More providers like Apple Music and Amazon Music are on the roadmap!*

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
TIDAL_ID="your_tidal_client_id"
TIDAL_SECRET="your_tidal_client_secret"
FRONTEND_URL="http://127.0.0.1:5175/"
PORTIFY_ENV="local" # Defines telemetry environment (e.g. local, staging, production)
```

> **Important:** The `FRONTEND_URL` must exactly match the authorized Redirect URI configured in both the Spotify and Google Cloud developer consoles (including the trailing slash).

### Frontend (`/frontend/.env`)

```env
# Set to 'true' to display a debug overlay showing raw state and sessionStorage.
VITE_SHOW_DEBUG_PANEL=false
```

## 📊 Observability & Telemetry

Portify is fully instrumented with **OpenTelemetry (OTel)** for metrics and tracing:

### 1. Metrics Exposed Locally (Pull Model)
The Go backend exposes standard Prometheus pull metrics at `/metrics` (default port `8080`).

### 2. Grafana Cloud Integration (Push Model)
To push traces and metrics automatically to an OTLP-compatible receiver (like Grafana Cloud), configure these environment variables in your `.env` file:

```env
OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp-gateway-prod-us-west-0.grafana.net/otlp"
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic <base64_encoded_token>"
OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf"
PORTIFY_ENV="staging" # Labels all metrics/spans with 'environment' (defaults to 'local' if omitted)
```

### Metrics Collected
* `portify_conversions_total`: Total playlist conversion attempts, labeled by `source`, `destination`, and `status` (`success`, `failed`).
* `portify_tracks_processed_total`: Total number of tracks matched or failed, labeled by `provider` and `status` (`success`, `match_failed`, `insert_failed`).
* `portify_api_requests_total`: Total outgoing HTTP requests to third-party APIs, labeled by `provider`, `operation` (e.g. `Search`, `PlaylistModify`), and `status_code`.
* `portify_api_latency_seconds`: High-precision latency histogram of outgoing API requests, configured with custom sub-second bucket views.
* `portify_api_retries_total`: Total retry attempts triggered by backoff loops, labeled by `provider`, `operation`, `attempt`, and `status_code`.



## ⚠️ Provider Quotas & Rate Limits

### Shared Developer Account Architecture
Portify operates on a **centralized developer application model**:
- The backend instance is configured with a single set of OAuth application developer credentials per provider (`SPOTIFY_ID`/`SPOTIFY_SECRET`, `YOUTUBE_ID`/`YOUTUBE_SECRET`, `TIDAL_ID`/`TIDAL_SECRET`).
- While individual users authenticate with their personal accounts via OAuth 2.0 to read and write to their own private libraries, **all outgoing API requests route through and are credited against the host server's registered developer credentials**.
- Consequently, the API quotas and rate limits are **shared across all users and all conversions** running through that Portify deployment.

Below is a detailed breakdown of how quotas and rate limits work for each supported service:

### 1. YouTube Music (YouTube Data API v3)
* **Budget Model:** Daily points-based quota allocated per Google Cloud project (resets daily at **midnight Pacific Time (PST/PDT)**).
* **Default Free Allocation:** **10,000 quota units / day** per Google Cloud project.
* **Quota Cost Per Operation:**
  * **Search (`youtube.search.list` / track matching):** **100 units** per query.
  * **Add to Playlist (`youtube.playlistItems.insert`):** **50 units** per track.
  * **Like Song (`youtube.videos.rate`):** **50 units** per track.
  * **Inspect Playlist Tracks (`youtube.playlistItems.list`):** **1 unit** per call (up to 50 tracks per page).
  * **List Playlists (`youtube.playlists.list`):** **1 unit** per call.
  * **Create Playlist (`youtube.playlists.insert`):** **50 units** (one-time).
* **Effective Conversion Capacity:** A single track conversion (Search + Insert) requires **150 units**. Under the free 10,000-unit tier, a single developer key can convert **~66 tracks per day across all users** before receiving `403 quotaExceeded`.
* **Mitigations & Handling in Portify:**
  * **Deduplication / Smart Resume (Issue #92):** When appending to an existing destination playlist, Portify fetches existing tracks (costing only 1 unit per 50 tracks) and skips already-present songs without making search or insert calls. This allows multi-day conversion runs for large playlists (e.g. 700+ songs) without duplicates or wasted quota.
  * **Quota Extension:** For multi-user or high-volume usage, developers must apply for a free [YouTube API Quota Extension](https://console.cloud.google.com/) (e.g. requesting 100,000–150,000 units).
  * **Project Rotation:** For self-hosted instances, rotating between multiple Google Cloud projects (swapping Client ID/Secret) provides an additional 10,000 units per project immediately.

### 2. Spotify (Spotify Web API)
* **Budget Model:** Rolling-window rate limiting (requests per time window, typically evaluated over ~30-second windows) rather than a hard daily cap.
* **Tiers & Restrictions:**
  * **Development Mode:** The default mode for newly created Spotify developer applications. Limited to **up to 25 explicitly allow-listed Spotify user accounts** registered in the Spotify Developer Dashboard.
  * **Extended Quota Mode:** Requires submitting an application to Spotify to remove the 25-user restriction for open public access.
* **Rate Limits & Behavior:**
  * Typically allows **~100–180 requests per 30-second window** per client ID.
  * When exceeded, Spotify responds with `HTTP 429 Too Many Requests` along with a `Retry-After: <seconds>` response header.
* **Portify's Handling:** Portify includes an automatic exponential backoff client (`internal/adapters/common/retry_client.go`) that dynamically respects `Retry-After` headers and pauses execution before automatically resuming, preventing dropped tracks.
* **Effective Conversion Capacity:** Significantly higher throughput than YouTube. A single developer account can easily convert hundreds to thousands of tracks per day, provided calls are paced within Spotify's rolling rate windows.

### 3. Tidal (Tidal Developer API)
* **Budget Model:** Concurrency and token-bucket request throttling per OAuth Client ID.
* **Default Allocation:** Standard developer accounts are throttled at approximately **5–10 requests per second (RPS)** (~300 requests per minute).
* **Rate Limits & Behavior:**
  * Exceeding burst limits results in an `HTTP 429 Too Many Requests` response.
  * Each operation (search track, fetch playlist, add track) counts as a standard 1 HTTP request without points-based multipliers.
* **Portify's Handling:** Handled automatically by the backend retry and rate-limiting wrapper, which buffers requests and retries with jitter if burst limits are momentarily reached.
* **Effective Conversion Capacity:** High capacity for personal and self-hosted use. Migrating typical playlists of several hundred songs completes smoothly within minutes.

---

### Quota & Rate Limit Comparison Summary

| Service | Rate Limit Model | Default Developer Tier | Cost per Converted Track | Effective Daily Capacity | Over-Limit Response |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **YouTube Music** | Daily Points Budget | 10,000 units / day | ~150 units (100 search + 50 insert) | **~66 tracks / day** | `403 Forbidden` (`quotaExceeded`) |
| **Spotify** | Rolling Window (30s) | ~100–180 req / 30s (Max 25 users in Dev Mode) | 1–2 API calls | **Thousands of tracks / day** (paced) | `429 Too Many Requests` (`Retry-After`) |
| **Tidal** | Token Bucket (RPS) | ~5–10 req / sec | 1–2 API calls | **Thousands of tracks / day** (paced) | `429 Too Many Requests` |

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

### 3. Run the App Locally

Start both the Go backend and React frontend with a single command:

```bash
make dev
```

This launches the ConnectRPC server on `http://localhost:8080` and the Vite dev server on `http://127.0.0.1:5175` concurrently. Vite proxies all `/converter.v1.ConverterService/*` requests to the Go backend automatically. Press `Ctrl+C` to stop both. You do *not* need Docker or Cloudflare for basic local development!

You can also start them individually with `make dev-backend` or `make dev-frontend`.

### 4. Staging Deployment (Docker + Cloudflare)

We utilize a headless server to host Portify's containerized infrastructure securely on the public internet.

*   **URL:** [https://staging-portify.debalin.dev](https://staging-portify.debalin.dev)
*   **Infrastructure:** A local Ubuntu server running `docker-compose`. Nginx acts as a reverse proxy serving the compiled React frontend on port 80 and piping all API requests to the isolated Go backend container, completely resolving CORS cross-origin concerns natively.
*   **Networking:** Instead of exposing local server ports to the web, `cloudflared` runs as a sidecar container, creating a secure Zero Trust tunnel from the internal Docker network out to the public internet, providing fully-managed HTTPS out of the box.

To spin up the staging environment on your host machine:
```bash
docker compose up --build -d
```
*(Ensure your `CLOUDFLARE_TUNNEL_TOKEN` and `FRONTEND_URL` are set inside `.env` first!)*

## 🧪 Testing & Quality

### Makefile Commands

| Command | Description |
|---|---|
| `make setup` | Bootstrap the full developer environment |
| `make dev` | **Start both backend + frontend in one terminal** |
| `make dev-backend` | Start only the Go backend server |
| `make dev-frontend` | Start only the Vite dev server |
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
