#! /usr/bin/make -f

# Project variables.
PACKAGE := github.com/hermeznetwork/hermez-node
VERSION := $(shell git describe --tags --always)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
PROJECT_NAME := $(shell basename "$(PWD)")

# Go related variables.
GO_FILES := $(shell find . -type f -name '*.go' | grep -v vendor)
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/dist
GOPKG := $(.)
GOENVVARS := GOBIN=$(GOBIN)
GOCMD := $(GOBASE)/cmd/heznode
GOPROOF := $(GOBASE)/test/proofserver/cmd
GOBINARY := heznode
# Go 1.13+ do not require GOPATH to be set, but we use some binaries in $GOPATH/bin
GOPATH ?= $(shell go env GOPATH)
PACKR := $(GOPATH)/bin/packr2
GORELEASER := $(GOPATH)/bin/goreleaser
GOCILINT := $(GOPATH)/bin/golangci-lint

# Docker configs.
DOCKER_IMAGE = hermez-node-build
DOCKER_BUILD_FILE = build/Dockerfile

# Project configs.
MODE ?= sync
CONFIG ?= $(GOBASE)/cmd/heznode/cfg.buidler.toml
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

# GNU make is verbose. Make it silent.
MAKEFLAGS += --silent

# For go 1.13-1.15 compatibility
# https://maelvls.dev/go111module-everywhere/
export GO111MODULE=on

## build: Build the project.
build: clean migration-pack
	echo "  >  Building Hermez binary..."
	$(GOENVVARS) go build $(LDFLAGS) -o $(GOBIN)/$(GOBINARY) $(GOCMD)
	$(MAKE) migration-clean

.PHONY: help
help: Makefile
	echo
	echo " Choose a command run in "$(PROJECT_NAME)":"
	echo
	sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	echo

## test: Run the application check and all tests.
test: govet gocilint test-unit

## test-unit: Run all unit tests.
# NOTE: `-p 1` forces execution of package test in serial. Otherwise, they may be
# executed in parallel, and the test may find unexpected entries in the SQL database
# because it's shared among all tests.
test-unit:
	echo "  >  Running unit tests"
	$(GOENVVARS) $(PGENVVARS) go test -race -p 1 -failfast -timeout 300s -v ./...

## test-api-server: Run the API server using the Go tests.
test-api-server:
	echo "  >  Running unit tests"
	$(GOENVVARS) $(PGENVVARS) FAKE_SERVER=yes go test -race -timeout 0 ./api -p 1 -count 1 -v

## gofmt: Run `go fmt` for all go files.
gofmt: .stamp.gofmt
.stamp.gofmt: $(GO_FILES)
	echo "  >  Format all go files"
	$(GOENVVARS) gofmt -w $(GO_FILES)
	touch $@

## govet: Run go vet.
govet: .stamp.govet
.stamp.govet: $(GO_FILES)
	echo "  >  Running go vet"
	$(GOENVVARS) go vet ./...
	touch $@

$(GOCILINT):
	echo "  >  Installing gocilint"
	cd && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0

## gocilint: Run Golang CI Lint.
gocilint: .stamp.gocilint
.stamp.gocilint: $(GOCILINT) $(GO_FILES)
	echo "  >  Running Golang CI Lint"
	$(GOCILINT) run --timeout=5m -E whitespace -E gosec -E gci -E misspell -E gomnd -E gofmt -E goimports -E golint --exclude-use-default=false --max-same-issues 0
	touch $@

## clean: Clean build files. Runs `go clean` internally.
clean:
	-rm -r $(GOBIN) .stamp.* 2> /dev/null
	-mkdir -p $(GOBIN) .stamp.* 2> /dev/null
	echo "  >  Cleaning build cache"
	$(GOENVVARS) go clean ./...

## gomod-download: Install missing dependencies.
gomod-download:
	echo "  >  Checking if there is any missing dependencies..."
	$(GOENVVARS) go mod download

## docker-build: Build the binaries with Docker.
docker-build: clean $(GORELEASER)
	echo "  >  Building with docker..."
	docker build -t $(DOCKER_IMAGE) -f $(DOCKER_BUILD_FILE) .
	docker run -t --rm -v "$(PWD)/dist:/build/dist" $(DOCKER_IMAGE) make goreleaser

$(GORELEASER):
	echo "  >  Installing goreleaser"
	cd && go install github.com/goreleaser/goreleaser@latest

## goreleaser: Build the binaries with Goreleaser.
goreleaser: clean $(GORELEASER)
	echo "  >  Building with goreleaser..."
	cd $(GOBASE) && $(GORELEASER) --snapshot --skip-publish
	$(MAKE) migration-clean

$(PACKR):
	echo "  >  Installing packr2"
	cd && go get github.com/gobuffalo/packr/v2/packr2

## migration-pack: Pack the database migrations into the binary.
migration-pack: $(PACKR)
	echo "  >  Packing the migrations..."
	cd $(GOBASE)/db && $(PACKR)

## migration-clean: Clean the database migrations pack.
migration-clean:
	echo "  >  Cleaning the migrations..."
	cd $(GOBASE)/db && $(PACKR) clean

## run-node: Run Hermez node.
run-node: build
	echo "  >  Running $(PROJECT_NAME)"
	$(GOBIN)/$(GOBINARY) run --mode $(MODE) --cfg $(CONFIG)

## wipedbs: Clean Hermez node databases.
wipedbs: build
	echo "  >  Cleaning local databases"
	$(GOBIN)/$(GOBINARY) wipedbs --mode $(MODE) --cfg $(CONFIG)

## run-proof-mock: Run proof server mock API.
run-proof-mock: stop-proof-mock
	echo "  >  Running Proof Server Mock"
	$(GOENVVARS) go build -o $(GOBIN)/proof $(GOPROOF)
	$(GOBIN)/proof 2>&1 & echo $$! > $(PID_PROOF_MOCK)
	cat $(PID_PROOF_MOCK) | sed "/^/s/^/  \>  Proof Server Mock PID: /"

## stop-proof-mock: Stop proof server mock API.
stop-proof-mock:
	touch $(PID_PROOF_MOCK)
	-kill -s INT `cat $(PID_PROOF_MOCK)` 2> /dev/null
	-rm $(PID_PROOF_MOCK) $(GOBIN)/proof 2> /dev/null

## run-database-container: Run the Postgres container
run-database-container:
	echo "  >  Running the postgreSQL DB..."
	docker run --rm --name hermez-l2db-$(PGPORT) -p $(PGPORT):5432 -e POSTGRES_DB=$(PGDATABASE) -e POSTGRES_USER=$(PGUSER) -e POSTGRES_PASSWORD="$(PGPASSWORD)" -d postgres

## stop-database-container: Stop the Postgres container
stop-database-container:
	echo "  >  Stopping the postgreSQL DB..."
	docker stop hermez-l2db-$(PGPORT)

## exec: Run given command. e.g; make exec run="go test ./..."
exec:
	$(GOENVVARS) $(run)

## install: Install the heznode cli as a service in coord mode
install:
	echo "  > Installing heznode as a service"
	echo "  > Checking requirements"
ifneq ("$(wildcard ./dist/heznode)","")
	echo "  - heznode file found!"
else
	echo "  - heznode file not found!"
	echo "  - please, run make build before make install!"
	test -f ./dist/heznode
endif
ifneq ("$(wildcard ./cmd/heznode/cfg.builder.toml)","")
	echo "  - config template found!"
else
	echo "  - config template not found!"
	echo "  - please, check the ./cmd/heznode/cfg.builder.toml!"
	test -f ./cmd/heznode/cfg.builder.toml
endif
ifneq ("$(wildcard /etc/hermez/config.toml)","")
	echo "  > Config file already exists - ignored."
else
	echo "  > Copying config file to /etc/hermez"
	mkdir -p /etc/hermez
	cp cmd/heznode/cfg.builder.toml /etc/hermez/config.toml
endif
ifeq ("$(wildcard /etc/systemd/system/heznode.service)", "")
	echo "  > Registering as a service"
	touch /etc/systemd/system/heznode.service
	echo "[Unit]" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "Description=Hermez Node" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "After=postgresql.service" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "StartLimitBurst=5" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "StartLimitIntervalSec=60" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "[Service]" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "Type=simple" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "Restart=always" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "RestartSec=1" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "ExecStart=/usr/local/bin/heznode run --mode coord --cfg /etc/hermez/config.toml" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "KillMode=process" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "StandardOutput=append:/var/log/hermez-node.log" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "StandardError=append:/var/log/hermez-node.log" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "[Install]" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "WantedBy=multi-user.target" | tee -a /etc/systemd/system/heznode.service > /dev/null
	echo "" | tee -a /etc/systemd/system/heznode.service > /dev/null
	systemctl daemon-reload
else
	echo "  > Service is already registered. Will be stoped!"
	service heznode stop
	rm /usr/local/bin/heznode
endif
	echo "  > Copying hez binary to /usr/local/bin"
	cp dist/heznode /usr/local/bin/heznode
	echo "  > Service is ready. Please update the configs at /etc/hermez/config.toml"
	echo "  > You can use the service with service heznode status|start|stop"
	echo "  Bye."
