EXECUTABLE=server
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
#VERSION=$(shell git describe --tags --always --long --dirty)
VERSION=local

PORT := 8080

default: clean build lint test

download:
	go mod download

test:
	go test -shuffle=on -race -coverprofile=coverage.txt ./...

lint: download
	golangci-lint run

build: clean windows #linux ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

$(WINDOWS):
	go build -o $(WINDOWS) -ldflags='-w -X main.version=$(VERSION)' .

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o $(LINUX) -ldflags="-s -w -X main.version=$(VERSION)" .

clean: ## Remove previous build
	del -f $(WINDOWS)

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