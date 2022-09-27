MAIN := main.go
SIDECAR_SERVER := sidecar-server

CURDIR := $(shell pwd)
OUTPUTDIR := build
SIDECAR_SERVER_DIR := $(CURDIR)/cmd/$(SIDECAR_SERVER)

GOFILE := $(shell find . -name "*.go" | xargs)

.PHONY: lint
lint:
	gofmt -w $(GOFILE)

.PHONY: build
build:
	go build -o $(OUTPUTDIR)/$(SIDECAR_SERVER) $(SIDECAR_SERVER_DIR)/$(MAIN)
	cp $(CURDIR)/nginx_conf.tpl $(OUTPUTDIR)/nginx_conf.tpl

.PHONY: clean
clean:
	rm -rf $(CURDIR)/$(OUTPUTDIR)

.PHONY: rebuild
rebuild: clean build
