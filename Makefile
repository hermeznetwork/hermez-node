#! /usr/bin/make -f

# Project variables.
PACKAGE := github.com/hermeznetwork/hermez-node
VERSION := $(shell git describe --tags --always)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
PROJECT_NAME := $(shell basename "$(PWD)")

# Go related variables.
GO_FILES ?= $$(find . -name '*.go' | grep -v vendor)
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOPKG := $(.)
GOENVVARS := GOBIN=$(GOBIN)
GOCMD := $(GOBASE)/cli/node
GOPROOF := $(GOBASE)/test/proofserver/cli
GOBINARY := heznode
PACKR := $(GOPATH)/bin/packr2

# Project configs.
MODE ?= sync
CONFIG ?= $(GOBASE)/cli/node/cfg.buidler.toml
PGHOST ?= localhost
PGPORT ?= 4012
PGUSER ?= hermez
PGPASSWORD ?= yourpasswordhere
PGDATABASE ?= hermez
PGENVVARS :=  PGHOST=$(PGHOST) PGPORT=$(PGPORT) PGUSER=$(PGUSER) PGPASSWORD=$(PGPASSWORD) PGDATABASE=$(PGDATABASE)

# Use linker flags to provide version/build settings.
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# PID file will keep the process id of the server.
PID_PROOF_MOCK := /tmp/.$(PROJECT_NAME).proof.pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## test: Run the application check and all tests.
test: govet gocilint test-unit

## test-unit: Run all unit tests.
# NOTE: `-p 1` forces execution of package test in serial. Otherwise, they may be
# executed in parallel, and the test may find unexpected entries in the SQL database
# because it's shared among all tests.
test-unit:
	@echo "  >  Running unit tests"
	$(GOENVVARS) $(PGENVVARS) go test -race -p 1 -failfast -timeout 300s -v ./...

## test-api-server: Run the API server using the Go tests.
test-api-server:
	@echo "  >  Running unit tests"
	$(GOENVVARS) $(PGENVVARS) FAKE_SERVER=yes go test -race -timeout 0 ./api -p 1 -count 1 -v

## gofmt: Run `go fmt` for all go files.
gofmt:
	@echo "  >  Format all go files"
	$(GOENVVARS) gofmt -w ${GO_FILES}

## govet: Run go vet.
govet:
	@echo "  >  Running go vet"
	$(GOENVVARS) go vet ./...

## golint: Run default golint.
golint:
	@echo "  >  Running golint"
	$(GOENVVARS) golint -set_exit_status ./...

## gocilint: Run Golang CI Lint.
gocilint:
	@echo "  >  Running Golang CI Lint"
	$-golangci-lint run --timeout=5m -E whitespace -E gosec -E gci -E misspell -E gomnd -E gofmt -E goimports -E golint --exclude-use-default=false --max-same-issues 0

## exec: Run given command. e.g; make exec run="go test ./..."
exec:
	@$(GOENVVARS) $(run)

## clean: Clean build files. Runs `go clean` internally.
clean:
	@-rm -r $(GOBIN) 2> /dev/null
	@echo "  >  Cleaning build cache"
	@$(GOENVVARS) go clean

## build: Build the project.
build: clean install
	@echo "  >  Building Hermez binary..."
	@bash -c "$(MAKE) migration-pack"
	@$(GOENVVARS) go build $(LDFLAGS) -o $(GOBIN)/$(GOBINARY) $(GOCMD)
	@bash -c "$(MAKE) migration-clean"

## install: Install missing dependencies.
install:
	@echo "  >  Checking if there is any missing dependencies..."
	@$(GOENVVARS) go mod download

## go-get: Install specific dependency. Runs `go get` internally. e.g; make install get=github.com/foo/bar
go-get:
	@echo "  >  Adding the missing dependency: $(get)"
	@$(GOENVVARS) go get $(GOCMD)/... $(get)

## run-node: Run Hermez node.
run-node:
	@bash -c "$(MAKE) clean build"
	@echo "  >  Running $(PROJECT_NAME)"
	@$(GOBIN)/$(GOBINARY) run --mode $(MODE) --cfg $(CONFIG)

## run-proof-mock: Run proof server mock API.
run-proof-mock: stop-proof-mock
	@echo "  >  Running Proof Server Mock"
	$(GOENVVARS) go build -o $(GOBIN)/proof $(GOPROOF)
	@$(GOBIN)/proof 2>&1 & echo $$! > $(PID_PROOF_MOCK)
	@cat $(PID_PROOF_MOCK) | sed "/^/s/^/  \>  Proof Server Mock PID: /"

## stop-proof-mock: Stop proof server mock API.
stop-proof-mock:
	@-touch $(PID_PROOF_MOCK)
	@-kill -s INT `cat $(PID_PROOF_MOCK)` 2> /dev/null || true
	@-rm $(PID_PROOF_MOCK) $(GOBIN)/proof 2> /dev/null || true

## install-packr: Install the packr for the database migrations.
install-packr:
ifeq ($(wildcard $(PACKR)),)
	@echo "  >  Installing packr2"
	@-bash -c "go get github.com/gobuffalo/packr/v2/packr2"
else
	@echo "  >  packr2 already installed"
endif

## migration-pack: Pack the database migrations into the binary.
migration-pack: install-packr
	@echo "  >  Packing the migrations..."
	@cd $(GOBASE)/db && $(PACKR) && cd -

## migration-clean: Clean the database migrations pack.
migration-clean:
	@echo "  >  Cleaning the migrations..."
	@cd $(GOBASE)/db && $(PACKR) clean && cd -

## run-database-container: Run the Postgres container
run-database-container:
	echo "  >  Running the postgreSQL DB..."
	docker run --rm --name hermez-l2db-$(PGPORT) -p $(PGPORT):5432 -e POSTGRES_DB=$(PGDATABASE) -e POSTGRES_USER=$(PGUSER) -e POSTGRES_PASSWORD="$(PGPASSWORD)" -d postgres

## stop-database-container: Stop the Postgres container
stop-database-container:
	echo "  >  Stopping the postgreSQL DB..."
	docker stop hermez-l2db-$(PGPORT)
