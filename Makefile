MIN_MAKE_VERSION := 3.81

# Min version
ifneq ($(MIN_MAKE_VERSION),$(firstword $(sort $(MAKE_VERSION) $(MIN_MAKE_VERSION))))
	$(error GNU Make $(MIN_MAKE_VERSION) or higher required)
endif

GO_LDFLAGS ?= -w -extldflags "-static" -X main.GitRevision=$(GIT_REVISION) -X main.Version=$(GIT_TAG_VERSION)
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_TAG_VERSION := $(shell git tag -l --points-at HEAD | grep -v latest)

ifeq ($(CI),true)
	GO_TEST_EXTRAS ?= "-coverprofile=c.out"
endif

##@ Help
.PHONY: help
help: ## Show all available commands (you are looking at it)
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: build run up down up-detached

generate-swagger-docs: ## Generate
	swag init --generalInfo main.go --dir "cmd/,pkg/health/,pkg/net/http/ginprometheus" --output cmd/app/swagger --parseDependency

build: ## Build in docker
	./script/build.sh

run: up ## alias

up: ## Start up application container
	docker compose up --build

up-detached: # Start up application in the background
	docker compose up --build -d

down: ## Stop and remove the application containers
	docker compose down --volumes

##@ Migrations
.PHONY: migration-create migration-up migration-down-by-one migration-down-all

migration-create: ## Create a new migration (usage: make migratíon-create name=your_migration_name)
	@if [ -z "$(name)" ]; then echo "Migration name not provided. Usage: make migration-create name=your_migration_name"; exit 1; fi
	docker compose run --rm \
	  -e DATABASE_URL=${DATABASE_URL} \
	  whalebone-clients \
	  goose -dir db/migrations create $(name) sql

migration-up: ## Apply all up migrations
	docker compose run --rm \
	  -e DATABASE_URL=${DATABASE_URL} \
	  whalebone-clients \
	  goose postgres "$$DATABASE_URL" -dir db/migrations -v up status

migration-down-by-one: ## Roll back the last migration
	docker compose run --rm \
	  -e DATABASE_URL=${DATABASE_URL} \
	  whalebone-clients \
	  goose postgres "$$DATABASE_URL" -dir db/migrations -v down status

migration-down-all: ## Roll back all migrations
	docker compose run --rm \
	  -e DATABASE_URL=${DATABASE_URL} \
	  whalebone-clients \
	  goose postgres "$$DATABASE_URL" -dir db/migrations -v reset status


##@ Test
.PHONY: test mtest test-pattern mtest-pattern
test: ## Run tests via docker compose
	./script/test.sh

mtest: ## Run tests via docker compose, supports arm arch
	./script/arm/test.sh

test-pattern: ## Run test with pattern
	./script/test-pattern.sh $(pattern)

mtest-pattern: ## Run test with pattern, supports arm arch
	./script/arm/test-pattern.sh $(pattern)
