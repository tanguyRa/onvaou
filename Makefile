PROJECT_NAME ?= project
DEV_COMPOSE=docker --log-level error compose -f compose.yml -p ${PROJECT_NAME}
SECURE_COMPOSE=op run --env-file=".env.dev" -- docker compose -f compose.yml -p ${PROJECT_NAME}

BACKEND_SERVICE := back

.PHONY: help build start stop restart logs clean \
        check-backend-running migrate-up migrate-down migration sqlc \
        front back test ingest ingest-artifacts ingest-replay

# ============================================
# Help
# ============================================
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development:"
	@echo "  start         Start all services (open app at 127.0.0.1:3000)"
	@echo "  stop          Stop all services"
	@echo "  restart       Restart all services"
	@echo "  logs          View logs from all services"
	@echo "  clean         Remove containers and volumes"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up    Run database migrations (N optional, default: all)"
	@echo "  migrate-down  Rollback migrations (N optional, default: 1)"
	@echo "  migration     Create new migration (name=\"...\")"
	@echo "  sqlc          Generate SQL code"
	@echo ""
	@echo "Utilities:"
	@echo "  front         SSH into frontend container"
	@echo "  back          SSH into backend container"
	@echo "  test          Run tests"
	@echo "  ingest        Run Go ETL worker in backend container (SOURCE=openagenda|datagouv|vacuum or ALL=1)"
	@echo "  ingest-artifacts List recent ETL debug artifacts (LIMIT optional, default: 20)"
	@echo "  ingest-replay Replay an ETL debug artifact (ARTIFACT_ID required)"
	@echo "  stats         Show container statistics"
	@echo "  pull          Pull latest images"
	@echo "  db_env        Show database environment variable to verify 1password/docker setup"
	@echo ""
	@echo "Production:"
	@echo "  build-prod    Build production images"

# ============================================
# Development
# ============================================
start: build start-all migrate
	@echo "Started all services and applied migrations. Access the app at http://127.0.0.1:3000"

build:
	${SECURE_COMPOSE} build --pull

start-all:
	${SECURE_COMPOSE} up -d --remove-orphans

stop:
	${DEV_COMPOSE} down

restart: stop start

logs:
	${DEV_COMPOSE} logs -f

clean:
	${DEV_COMPOSE} down -v --remove-orphans
	@echo "Cleaned up containers and volumes"

# ============================================
# Database Management
# ============================================
check-backend-running:
	@${DEV_COMPOSE} ps --status running --services | grep -qx "${BACKEND_SERVICE}" || (echo "backend service '${BACKEND_SERVICE}' is not running. Start it with 'make start'."; exit 1)

migrate: migrate-up
migrate-up: check-backend-running
	@echo "Applying migrations..."
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'migrate -path /db/migrations -database "$$DATABASE_URL" up ${N}'

migrate-down: check-backend-running
	@echo "Rolling back $(if $(N),$(N),1) migrations..."
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'migrate -path /db/migrations -database "$$DATABASE_URL" down $(if $(N),$(N),1)'

migration:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migration name=\"migration_name\""; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@mkdir -p db/migrations
	@touch db/migrations/$$(date +%Y%m%d%H%M%S)_$(name).up.sql
	@touch db/migrations/$$(date +%Y%m%d%H%M%S)_$(name).down.sql
	@echo "Created migration files"

sqlc: check-backend-running
	@echo "Generating SQL code..."
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} go run -mod=mod github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate -f /db/sqlc.yml
	@echo "Done"

# ============================================
# Utilities
# ============================================
front:
	${DEV_COMPOSE} exec -it front sh

back:
	${DEV_COMPOSE} exec -it ${BACKEND_SERVICE} sh

stats:
	docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"

test:
	${DEV_COMPOSE} exec -it ${BACKEND_SERVICE} go test ./...

ingest: check-backend-running
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'if [ "$(ALL)" = "1" ]; then go run ./cmd/ingest --all; elif [ -n "$(SOURCE)" ]; then go run ./cmd/ingest --source="$(SOURCE)"; else go run ./cmd/ingest --all; fi'

ingest-artifacts: check-backend-running
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'go run ./cmd/ingest --list-artifacts --limit="$(if $(LIMIT),$(LIMIT),20)"'

ingest-replay: check-backend-running
	@if [ -z "$(ARTIFACT_ID)" ]; then \
		echo "Usage: make ingest-replay ARTIFACT_ID=<artifact-id>"; \
		exit 1; \
	fi
	${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'go run ./cmd/ingest --replay-artifact="$(ARTIFACT_ID)"'

db_env:
	op run --env-file=".env.dev" -- ${DEV_COMPOSE} exec -T ${BACKEND_SERVICE} sh -c 'printenv | grep DATABASE_URL'

# ============================================
# Production
# ============================================
build-prod:
	docker compose --env-file .env.prod -f compose.yml -p ${PROJECT_NAME} build
