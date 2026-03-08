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

## 🏃 Getting Started 

*(Detailed setup instructions will be published as the project approaches its first alpha release.)*
