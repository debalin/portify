# 🎵 Portify

*The Universal Playlist Converter*

**Portify** is an open-source tool aimed at breaking down the walled gardens of music streaming. It allows users to seamlessly convert and synchronize their carefully curated playlists across multiple music platforms—moving your music freely between services like Spotify, YouTube Music, Apple Music, and more!

> ⚠️ **Note:** Portify is currently a **Work in Progress (WIP)**. The core backend architecture is active, but the frontend UI and additional streaming providers are still being developed.

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
- [Buf CLI](https://buf.build/docs/installation) (for compiling Protocol Buffers)

### 1. Generating Protobufs
We use `buf` to generate Go models and ConnectRPC routing code from our `.proto` definitions.
```bash
# Run this from the root of the project every time you change a .proto file
buf generate
```

### 2. Building the Project
Once the generated code is in place, fetch dependencies and verify the build:
```bash
go mod tidy
go build ./...
```

### 3. Running the Server
```bash
go run ./cmd/server
```
*(Note: To actually test the adapters, use the isolated `testspotify` and `testyoutube` scripts in the `/cmd` directory, which rely on local `.env` variables).*
