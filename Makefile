MAIN := main.go
SIDECAR_SERVER := sidecar-server
SIDECAR_CTL := sidecar-ctl

CURDIR := $(shell pwd)
OUTPUTDIR := build
SIDECAR_SERVER_DIR := $(CURDIR)/cmd/$(SIDECAR_SERVER)
SIDECAR_CTL_DIR := $(CURDIR)/cmd/$(SIDECAR_CTL)

GOFILE := $(shell find . -name "*.go" | xargs)


.PHONY: lint
lint:
	gofmt -w $(GOFILE)

.PHONY: build
build:
	go build -o $(OUTPUTDIR)/$(SIDECAR_SERVER) $(SIDECAR_SERVER_DIR)/$(MAIN)
	go build -o $(OUTPUTDIR)/$(SIDECAR_CTL) $(SIDECAR_CTL_DIR)/$(MAIN)

.PHONY: clean
clean:
	rm -rf $(CURDIR)/$(OUTPUTDIR)

.PHONY: rebuild
rebuild: clean build
