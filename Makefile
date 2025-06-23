# Paths
BACKEND_DIR=./backend
FRONTEND_DIR=./frontend
BINARY_NAME=chat-app
# Scripts
AIR_CMD=cd $(BACKEND_DIR) && air
FRONTEND_DEV=cd $(FRONTEND_DIR) && npm run dev
FRONTEND_BUILD=cd $(FRONTEND_DIR) && npm run build
FRONTEND_PREVIEW=cd $(FRONTEND_DIR) && npm run preview

# ===== DEV =====

dev:
	$(AIR_CMD) & \
	$(FRONTEND_DEV)

# ===== PROD =====

prod: build-frontend run-backend

build-frontend:
	$(FRONTEND_BUILD)

run-backend:
	cd $(BACKEND_DIR) && go build -o $(BINARY_NAME) . && ./$(BINARY_NAME)

preview:
	$(FRONTEND_PREVIEW)

# ===== UTIL =====

install-air:
	go install github.com/air-verse/air@latest

.PHONY: dev prod build-frontend run-backend preview install-air
