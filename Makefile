MAIN := main.go
SIDECAR_SERVER := sidecar-server
SIDECAR_CTL := sidecar-ctl

CURDIR := $(shell pwd)
SIDECAR_SERVER_DIR := $(CURDIR)/cmd/$(SIDECAR_SERVER)
SIDECAR_CTL_DIR := $(CURDIR)/cmd/$(SIDECAR_CTL)

GOFILE := $(shell find . -name "*.go" | xargs)


.PHONY: lint
lint:
	gofmt -w $(GOFILE)

.PHONY: build
build:
	go build -o $(SIDECAR_SERVER) $(SIDECAR_SERVER_DIR)/$(MAIN)
