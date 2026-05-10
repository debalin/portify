# Portify Deployment Guide

This guide walks through deploying the Portify application on a fresh Ubuntu Linux server using Docker, Docker Compose, and Cloudflare Tunnels for secure HTTPS exposure.

## Prerequisites

- A fresh Ubuntu server (e.g., 22.04 LTS or newer)
- SSH access to the server
- A Cloudflare account and a registered domain (for the tunnel)
- OAuth Developer Credentials for Spotify and YouTube/Google Cloud

## 1. Initial Server Setup & Docker Installation

SSH into your Ubuntu server and update your packages:

```bash
sudo apt update
sudo apt upgrade -y
```

Install Docker and Docker Compose:

```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to the docker group so you don't need sudo for every docker command
sudo usermod -aG docker $USER
newgrp docker
```

## 2. Clone the Repository

Install Git and clone the Portify repository:

```bash
sudo apt install git -y
git clone https://github.com/debalin/portify.git
cd portify
```

## 3. Configure Environment Variables

Portify requires several environment variables for OAuth and the Cloudflare tunnel. Copy the example environment file:

```bash
cp .env.example .env
```

Edit the `.env` file with your favorite text editor (e.g., `nano .env`) and fill in your credentials:

```env
SPOTIFY_ID="your_spotify_client_id"
SPOTIFY_SECRET="your_spotify_client_secret"
YOUTUBE_ID="your_google_cloud_client_id"
YOUTUBE_SECRET="your_google_cloud_client_secret"
FRONTEND_URL="https://your-public-domain.com/"
CLOUDFLARE_TUNNEL_TOKEN="your_tunnel_token"
```

> **Important:** The `FRONTEND_URL` must exactly match the authorized Redirect URI configured in both the Spotify and Google Cloud developer consoles (including the trailing slash).

## 4. Deploying the Application

With Docker installed and the `.env` file configured, you can spin up the entire stack (Backend, Frontend, and Cloudflare Tunnel) using Docker Compose:

```bash
docker compose up --build -d
```

This command will:
1. Build the lightweight Go backend container.
2. Compile the React frontend and serve it via an Nginx reverse proxy, avoiding CORS issues entirely.
3. Launch the `cloudflared` sidecar container to establish a secure tunnel to the internet without opening any firewall ports on the server.

## 5. Maintenance Commands

To view the live logs of your running containers:
```bash
docker compose logs -f
```

To gracefully stop the application:
```bash
docker compose down
```

To pull the latest code changes from GitHub and cleanly redeploy:
```bash
git pull
docker compose up --build -d
```
