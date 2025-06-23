# Chat-App

A full-stack real-time chat application built with:

* **Backend** – Go 1.24 · Gorilla WebSocket · uber-zap logging
* **Frontend** – React 19 · TypeScript · Vite 6 · Tailwind CSS 4 · Radix-UI · Socket.IO client
* **Dev tooling** – Bun / npm · ESLint 9 · Makefile · Air hot-reload for Go

---

## Table of Contents
1. Features
2. Architecture overview
3. Technology stack
4. Local setup
5. Environment variables
6. Development workflow
7. Production build & deploy
8. Project structure
9. Makefile targets
10. Troubleshooting
11. Contributing
12. License

---

## 1  Features
* Instant one-to-many messaging over WebSocket
* Typed, fully hot-reloading front-end and back-end dev experience
* Clean, minimal UI (Tailwind CSS + Radix-UI)
* Structured logging with uber-zap
* Makefile orchestration for single-command dev / prod builds

---

## 2  Architecture Overview

```
chat-app
├── backend              (Go 1.24 service)
│   ├── main.go          – entrypoint; HTTP + WS server
│   ├── routes/          – HTTP route registration
│   ├── websocket/       – connection manager & client abstractions
│   └── logger/          – singleton zap sugared logger
├── frontend             (React 19 + Vite app)
│   ├── src/             – TSX components & utilities
│   ├── public/          – static assets
│   └── vite.config.ts   – Tailwind + React plugin setup
└── Makefile             – dev/prod helpers
```

The backend exposes `/ws`, handled by a **Manager** that maintains an in-memory `ClientList`. Each incoming message is broadcast to all connected peers (except the sender) through buffered `egress` channels, avoiding concurrent write issues.

---

## 3  Technology Stack

| Layer | Libraries / Tools | Purpose |
|-------|-------------------|---------|
| Runtime | Go 1.24, Node 20 / Bun | Core languages |
| WebSocket | `github.com/gorilla/websocket` | Real-time transport |
| Logging | `go.uber.org/zap` | Fast, structured logging |
| Frontend | React 19, Vite 6, TypeScript 5.8 | SPA scaffold |
| Styling | Tailwind CSS 4, `class-variance-authority`, `clsx` | Utility-first styling |
| UI | Radix-UI Avatar, Lucide Icons | Accessible primitives & icons |
| Realtime client | `socket.io-client` (TS types via `@types/socket.io-client`) | WebSocket abstraction |
| Tooling | ESLint 9, Air hot-reload, Makefile | DX & automation |

---

## 4  Local Setup

```bash
# Clone & enter the repo
git clone https://github.com/<you>/chat-app.git
cd chat-app

# Install Go deps
cd backend && go mod download   # requires Go ≥ 1.24
# (optional) install Air for live-reload
go install github.com/air-verse/air@latest
cd ..

# Install JS deps (Node 20 or Bun)
cd frontend && bun install      # or: npm install
cd ..
```

---

## 5  Environment Variables

Create two `.env` files (they are git-ignored):

### backend/.env
```
ALLOWED_ORIGINS=http://localhost:5173
PORT=8080            # optional – defaults to 8080
```

### frontend/.env
```
VITE_WS_URL=ws://localhost:8080/ws
```

---

## 6  Development Workflow

The **Makefile** abstracts everything:

```bash
# Start Go Air + Vite dev servers concurrently
make dev
```

* `air` watches and hot-reloads Go code on changes.
* `vite` (port 5173) serves & HMRs the React app.

---

## 7  Production Build & Deploy

```bash
# Bundle React, then compile & run the Go binary
make prod
```

1. `npm run build` creates `frontend/dist` (static assets).
2. Go compiles `backend/main.go` into `chat-app` (port 8080).
3. Serve `dist` via CDN / Nginx & proxy `/ws` to the Go service in production.

---

## 8  Detailed Project Structure

| Path | Description |
|------|-------------|
| `backend/main.go` | Initializes zap logger, registers routes, starts `http.ListenAndServe`. |
| `backend/routes/routes.go` | Binds `/ws` → `websocket.Manager.ServeWs`. |
| `backend/websocket/manager.go` | Upgrader config (compression, CORS via `ALLOWED_ORIGINS`), add/remove clients. |
| `backend/websocket/client.go` | Per-connection goroutines: `readMessage`, `writeMessage`, graceful close. |
| `frontend/src/components` | `Home.tsx`, `Chat-room.tsx` and shared UI components. |
| `frontend/src/lib` | Utility hooks / helpers (e.g., socket hook). |
| `vite.config.ts` | Tailwind plugin & path alias config. |

---

## 9  Makefile Targets

| Target | Purpose |
|--------|---------|
| `make dev` | Hot-reloading development mode (`air` + `vite`). |
| `make prod` | Build React & run compiled Go binary locally. |
| `make preview` | Serve built React app (static) on port 4173. |
| `make install-air` | Installs Air globally for Go live-reload. |

---

## 10  Troubleshooting

| Symptom | Fix |
|---------|-----|
| `websocket: request origin not allowed` | Ensure `ALLOWED_ORIGINS` matches the frontend origin (`scheme://host:port`). |
| Port clash on 8080 / 5173 | Export `PORT` (backend) or edit `vite.config.ts` server port. |
| Go changes not reloaded | Verify `air` is installed and detects `backend/*.go`. |


---

## 11  License

Released under the MIT License © 2025.
