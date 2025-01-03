EXECUTABLE=server
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64
#VERSION=$(shell git describe --tags --always --long --dirty)
VERSION=local

PORT := 8080

default: clean build lint test

.PHONY: all test clean

download:
	go mod download

test:
	go test -shuffle=on -race -coverprofile=coverage.txt ./...

lint: download
	golangci-lint run


build: windows #linux darwin ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

$(WINDOWS):
	go build -o $(WINDOWS) -ldflags='-s -w -X main.version=$(VERSION)' .

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o $(LINUX) -ldflags="-s -w -X main.version=$(VERSION)" .

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o $(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)" .

clean: ## Remove previous build
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)

## This needs to be updated for the current system
run: build
	./$(WINDOWS) --port=$(PORT)

watch:
	air 


# IMAGE := ghcr.io/raeperd/kickstart
# DOCKER_VERSION := $(if $(VERSION),$(subst /,-,$(VERSION)),latest)

# docker:
# 	docker build . --build-arg VERSION=$(VERSION) -t $(IMAGE):$(DOCKER_VERSION)

# docker-run: docker 
# 	docker run --rm -p $(PORT):8080 $(IMAGE):$(DOCKER_VERSION)

# docker-clean:
# 	docker image rm -f $(IMAGE):$(DOCKER_VERSION) || true