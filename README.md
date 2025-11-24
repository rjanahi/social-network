# Social Network (Next.js + Go backend)

This repository contains a full-stack social network application:

- Frontend: Next.js (Pages Router) in `frontend-next/`.
- Backend: Go HTTP server in `backend/` with SQLite and a WebSocket hub for realtime features.

The project includes Dockerfiles and `docker-compose.yml` to run the two services together.

## Quick start (recommended: Docker Compose)

1. Build and start services:

```bash
docker-compose up -d --build
```

2. Open the frontend in your browser:

```text
http://localhost:3000
```

The backend listens on port 8080 and serves APIs and static assets for the frontend.

## Run locally without Docker

Start the backend:

```bash
bash back.sh
```

Start the frontend:

```bash
bash front.sh
```

By default the frontend expects the backend at `http://localhost:8080`. When running in Docker Compose the frontend is configured to use the `backend` service address via `BACKEND_URL` / `NEXT_PUBLIC_BACKEND_URL`.

## Database & Migrations

- SQLite database file: `backend/data/socialnetwork.db` (created at runtime).
- Migrations live at `backend/pkg/db/migrations/sqlite` and are applied automatically by the backend on startup using `github.com/golang-migrate/migrate/v4`.



## Main features to test

- User registration and login (sessions via cookie)
- Profiles, avatars, date-of-birth and age handling
- Follow/unfollow & follow requests (private/public profiles)
- Posts, comments, likes and privacy levels
- Groups: create, invite, join, posts and events
- Realtime chat and notifications using WebSockets (`/ws`)



## Project layout (high-level)

- `backend/` — Go server, route wiring in `backend/web/connectWeb.go`, api logic in `backend/pkg/apis/*`, db helpers in `backend/pkg/db/*`.
- `frontend-next/` — Next.js app: `pages/`, `components/`, `contexts/`, `hooks/`, `public/`, `styles/`.
- `docker-compose.yml` — orchestrates `backend` and `frontend` services for local development.

## Prerequisites

- Go (for backend, optional if using Docker)
- Node.js / npm (for frontend, optional if using Docker)
- Docker & Docker Compose (recommended)
- sqlite3 (optional, for inspecting DB)

## Team

- syusuf
- sahahmed
- aramadha
- rjanahi
- omkhalid


