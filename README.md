# Chat-App

A full-stack real-time chat application built with:

* **Backend** – Go 1.24 · Gorilla WebSocket · uber-zap logging · Redis · Prometheus metrics
* **Frontend** – React 19 · TypeScript · Vite 6 · Tailwind CSS 4 · Radix-UI · Socket.IO client
* **Dev tooling** – Docker Compose · Air hot-reload · Prometheus · Grafana

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
* Docker Compose for development and production environments
* Monitoring with Prometheus and Grafana
* Redis for pub/sub and state management
* Health check endpoints
* Room-based chat functionality

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

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development without Docker)
- Node.js 20+ or Bun (for frontend development)

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/KaranMali12/chat-app.git
cd chat-app/backend-v2

# Copy and configure environment variables
cp .env.example .env.dev
# Edit .env.dev with your configuration

# Start the development stack
docker-compose -f dev.docker-compose.yml up --build
```

The application will be available at:
- Chat App: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### Local Development (Without Docker)

```bash
# Install Go dependencies
cd backend-v2
go mod download

# Install Air for live-reload
go install github.com/air-verse/air@latest

# Start the application
air
```

For production deployment, use the production Docker setup:
```bash
cd backend-v2
docker-compose -f prod.docker-compose.yml up --build -d
```

---

## 5  Environment Variables

### Development (.env.dev)
```
# Application
APP_ENV=development
PORT=:8080
APP_NAME=chat-app

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# CORS
ALLOWED_ORIGINS=http://localhost:3000

# Prometheus
PROMETHEUS_ENABLED=true
```

### Production (.env.prod)
```
# Application
APP_ENV=production
PORT=:8080
APP_NAME=chat-app-prod

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=your_secure_password
REDIS_DB=0

# CORS - Update with your production domain
ALLOWED_ORIGINS=https://yourdomain.com

# Prometheus
PROMETHEUS_ENABLED=true
```

---

## 6  Development Workflow

### Using Docker (Recommended)

```bash
# Start the development stack
docker-compose -f dev.docker-compose.yml up --build
```

This will start:
- Go application with hot-reload (Air)
- Redis instance
- Prometheus for metrics
- Grafana for visualization

### Local Development

```bash
# Start the Go application with hot-reload
air

# In a separate terminal, start Redis
docker run -p 6379:6379 redis
```

### Monitoring
- **Prometheus**: http://localhost:9090
  - Configured to scrape Go application metrics
- **Grafana**: http://localhost:3000
  - Default credentials: admin/admin
  - Pre-configured with a dashboard for monitoring the chat application



---

## 8  Detailed Project Structure (Backend v2)

### Core Components

| Path | Description |
|------|-------------|
| `cmd/` | Application entry points |
| `cmd/main.go` | Main application entry point |
| `cmd/bootstrap.go` | Application initialization and dependency injection |
| `internal/` | Private application code |
| `internal/config/` | Configuration management |
| `internal/handler/` | HTTP request handlers |
| `internal/hub/` | WebSocket hub implementation |
| `internal/metrics/` | Prometheus metrics |
| `pkg/` | Reusable packages |
| `pkg/logger/` | Logging utilities |
| `pkg/redis/` | Redis client wrapper |

### Infrastructure

| Path | Description |
|------|-------------|
| `dev.Dockerfile` | Development Dockerfile with hot-reload |
| `prod.Dockerfile` | Production-optimized Dockerfile |
| `dev.docker-compose.yml` | Development environment with monitoring |
| `prod.docker-compose.yml` | Production environment |
| `prometheus.yml` | Prometheus configuration |
| `.env.example` | Example environment variables |

### API Endpoints
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics
- `GET /ws` - WebSocket endpoint
- `POST /api/v1/create-room` - Create a new chat room
- `GET /api/v1/room-stats` - Get room statistics

---

## 9  Docker Compose Commands

### Development

| Command | Description |
|---------|-------------|
| `docker-compose -f dev.docker-compose.yml up --build` | Start development stack |
| `docker-compose -f dev.docker-compose.yml down` | Stop development stack |
| `docker-compose -f dev.docker-compose.yml logs -f` | View logs |

### Production

| Command | Description |
|---------|-------------|
| `docker-compose -f prod.docker-compose.yml up -d --build` | Deploy production stack |
| `docker-compose -f prod.docker-compose.yml down` | Stop production stack |
| `docker-compose -f prod.docker-compose.yml logs -f` | View logs |

### Monitoring

| Service | URL | Default Credentials |
|---------|-----|---------------------|
| Prometheus | http://localhost:9090 | None |
| Grafana | http://localhost:3000 | admin/admin |

---

## 10  Troubleshooting

| Symptom | Fix |
|---------|-----|
| `websocket: request origin not allowed` | Ensure `ALLOWED_ORIGINS` in `.env` matches the frontend origin. |
| `connection refused` errors | Make sure Redis is running and the connection details in `.env` are correct. |
| Port conflicts | Check if ports 8080, 3000, 9090 are already in use. |
| Go changes not reloading | Ensure `air` is running and watching the correct directories. |
| Prometheus targets down | Check if the application is running and the `metrics` endpoint is accessible. |
| Grafana login issues | Default credentials are admin/admin. Reset with `docker-compose -f dev.docker-compose.yml exec grafana grafana-cli admin reset-admin-password newpassword` |


---

## 11  License

Released under the MIT License © 2025.
