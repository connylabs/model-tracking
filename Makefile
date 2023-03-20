export GO111MODULE=on
.PHONY: build migrate down fmt lint lint-openapi mock vendor sql-gen

OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
ALL_ARCH := amd64 arm arm64
BIN_DIR := bin
BINS := $(BIN_DIR)/$(OS)/$(ARCH)/model-tracking
PROJECT := model-tracking
PKG := github.com/connylabs/$(PROJECT)

TAG := $(shell git describe --abbrev=0 --tags HEAD 2>/dev/null)
COMMIT := $(shell git rev-parse HEAD)
VERSION := $(COMMIT)
ifneq ($(TAG),)
    ifeq ($(COMMIT), $(shell git rev-list -n1 $(TAG)))
        VERSION := $(TAG)
    endif
endif
DIRTY := $(shell test -z "$$(git diff --shortstat 2>/dev/null)" || echo -dirty)
VERSION := $(VERSION)$(DIRTY)
LD_FLAGS := -ldflags "-s -w -X $(PKG)/version.Version=$(VERSION)"
SRC := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

GO_FILES ?= $$(find . -name '*.go' -not -path './vendor/*')
GO_PKGS ?= $$(go list ./... | grep -v "$(PKG)/vendor")

MIGRATE_BINARY = $(BIN_DIR)/goose
GOLANGCI_LINT_BINARY := $(BIN_DIR)/golangci-lint
OAPI_CODEGEN_BINARY := $(BIN_DIR)/oapi-codegen
JET_BINARY := $(BIN_DIR)/jet

BUILD_IMAGE ?= golang:1.20.2-alpine
CONTAINERIZE_BUILD ?= true
BUILD_SUFIX :=
ifeq ($(CONTAINERIZE_BUILD), true)
	BUILD_PREFIX := docker run --rm \
	    -u $$(id -u):$$(id -g) \
	    -v $$(pwd):/$(PROJECT) \
	    -w /$(PROJECT) \
	    $(BUILD_IMAGE) \
	    /bin/sh -c '
	BUILD_SUFIX := '
endif

build: $(BINS)

build-%:
	@$(MAKE) --no-print-directory OS=$(word 1,$(subst -, ,$*)) ARCH=$(word 2,$(subst -, ,$*)) build

all-build: $(addprefix build-$(OS)-, $(ALL_ARCH))

$(BINS): $(SRC) go.mod
	@mkdir -p $(BIN_DIR)/$(word 2,$(subst /, ,$@))/$(word 3,$(subst /, ,$@))
	@echo "building: $@"
	@$(BUILD_PREFIX) \
	        GOARCH=$(word 3,$(subst /, ,$@)) \
	        GOOS=$(word 2,$(subst /, ,$@)) \
	        GOCACHE=$$(pwd)/.cache \
		CGO_ENABLED=0 \
		go build -mod=vendor -o $@ \
		    $(LD_FLAGS) \
		    ./cmd/$(@F) \
	$(BUILD_SUFIX)

$(BIN_DIR):
	mkdir -p $@

api/v1alpha1/v1alpha1.go: api/v1alpha1/v1alpha1.yaml $(OAPI_CODEGEN_BINARY)
	$(OAPI_CODEGEN_BINARY) -generate types,client,chi-server,spec -package v1alpha1 -o $@ $<

sql-gen: $(JET_BINARY)
	@docker run --rm --name generate -d -p 5433:5432 -e POSTGRES_USER=$(PROJECT) -e POSTGRES_PASSWORD=$(PROJECT) -e POSTGRES_DB=$(PROJECT) postgres
	until docker exec generate /usr/bin/pg_isready -d $(PROJECT) -h localhost -p 5432 -U $(PROJECT) -q; do sleep 1 ; done
	DATABASE_URL="user=$(PROJECT) password=$(PROJECT) dbname=$(PROJECT) host=localhost port=5433 sslmode=disable" $(MAKE) --no-print-directory migrate
	$(JET_BINARY) --source=postgresql --user=$(PROJECT) --password=$(PROJECT) --host=localhost --port=5433 --dbname=$(PROJECT) --sslmode=disable --schema=public --path=./store
	@docker kill generate

fmt:
	@echo $(GO_PKGS)
	gofmt -w -s $(GO_FILES)

lint: $(GOLANGCI_LINT_BINARY) lint-openapi
	$(GOLANGCI_LINT_BINARY) run

	@echo 'gofmt -d -s $(GO_FILES)'
	@fmt_res=$$(gofmt -d -s $(GO_FILES)); if [ -n "$$fmt_res" ]; then \
		echo ""; \
		echo "Gofmt found style issues. Please check the reported issues"; \
		echo "and fix them if necessary before submitting the code for review:"; \
		echo "$$fmt_res"; \
		exit 1; \
	fi

lint-openapi:
	docker run --rm -v $$(pwd):/var/lib/$(PROJECT) stoplight/spectral:5.9 lint /var/lib/$(PROJECT)/api/v1alpha1/v1alpha1.yaml

test:
	go test ./...

vendor:
	go mod tidy
	go mod vendor

mock:
	docker run --init --rm -v $$(pwd):/var/lib/$(PROJECT) -p 4010:4010 stoplight/prism:4 mock -h 0.0.0.0 /var/lib/$(PROJECT)/api/v1/v1.yaml

DATABASE_URL ?= user=$(PROJECT) password=$(PROJECT) dbname=$(PROJECT) host=localhost port=5432 sslmode=disable

migrate: $(MIGRATE_BINARY)
	$(MIGRATE_BINARY) -dir db/migrations postgres "$(DATABASE_URL)" up

down: $(MIGRATE_BINARY)
	$(MIGRATE_BINARY) -dir db/migrations postgres "$(DATABASE_URL)" down

$(GOLANGCI_LINT_BINARY): | $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v1.52.0

$(MIGRATE_BINARY): | $(BIN_DIR)
	go build -mod=vendor -o $@ github.com/pressly/goose/cmd/goose

$(OAPI_CODEGEN_BINARY): | $(BIN_DIR)
	go build -mod=vendor -o $@ github.com/deepmap/oapi-codegen/cmd/oapi-codegen

$(JET_BINARY): | $(BIN_DIR)
	go build -mod=vendor -o $@ github.com/go-jet/jet/v2/cmd/jet
