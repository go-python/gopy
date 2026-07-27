# Basic Go makefile

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

DIRS=`go list ./...`

PYTHON=python3
PIP=$(PYTHON) -m pip

GIT_COMMIT=`git rev-parse --short HEAD`
VERS_DATE=`date -u +%Y-%m-%d\ %H:%M`
LDFLAGS=-X 'main.GitCommit=$(GIT_COMMIT)' -X 'main.VersionDate=$(VERS_DATE) UTC'

all: build

build:
	@echo "GO111MODULE = $(value GO111MODULE)"
	$(GOBUILD) -v -ldflags "$(LDFLAGS)" $(DIRS)

test: 
	@echo "GO111MODULE = $(value GO111MODULE)"
	$(GOTEST) -v $(DIRS)

clean: 
	@echo "GO111MODULE = $(value GO111MODULE)"
	$(GOCLEAN) ./...

fmts:
	gofmt -s -w .
	
vet:
	@echo "GO111MODULE = $(value GO111MODULE)"
	$(GOCMD) vet $(DIRS) | grep -v unkeyed

tidy: export GO111MODULE = on
tidy:
	@echo "GO111MODULE = $(value GO111MODULE)"
	go mod tidy
	
mod-update: export GO111MODULE = on
mod-update:
	@echo "GO111MODULE = $(value GO111MODULE)"
	go get -u ./...
	go mod tidy

prereq:
	@echo "Installing python prerequisites -- ignore err if already installed:"
	- $(PIP) install -r requirements.txt
	@echo
	@echo "if this fails, you may see errors like this:"
	@echo "    Undefined symbols for architecture x86_64:"
	@echo "    _PyInit__gi, referenced from:..."
	@echo

# Releases are managed by release-please (.github/workflows/release-please.yml):
# merging its release PR to master bumps version.go, tags the commit, and
# publishes the GitHub Release. GoReleaser then builds & uploads binaries to it.

