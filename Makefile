MAIN := main.go
SIDECAR_SERVER := sidecar-server

CURDIR := $(shell pwd)
OUTPUTDIR := build
SIDECAR_SERVER_DIR := $(CURDIR)/cmd/$(SIDECAR_SERVER)

GOFILE := $(shell find . -name "*.go" | xargs)

.PHONY: lint
lint:
	gofmt -w $(GOFILE)

.PHONY: linux windows mac
linux: clean build_linux copy_tpl copy_linux_scripts
windows: clean build_windows copy_tpl copy_windows_scripts
mac: clean build_mac copy_tpl

.PHONY: build_linux build_windows build_mac copy_tpl
build_linux:
	GOARCH="amd64" GOOS="linux" go build -o $(OUTPUTDIR)/$(SIDECAR_SERVER) $(SIDECAR_SERVER_DIR)/$(MAIN)
build_windows:
	GOARCH="amd64" GOOS="windows" go build -ldflags="-H windowsgui" -o $(OUTPUTDIR)/sidecar-server.exe $(SIDECAR_SERVER_DIR)/$(MAIN)
build_mac:
	GOARCH="amd64" GOOS="darwin" go build -o $(OUTPUTDIR)/$(SIDECAR_SERVER) $(SIDECAR_SERVER_DIR)/$(MAIN)
copy_tpl:
	cp nginx_conf.tpl $(OUTPUTDIR)/nginx_conf.tpl
	cp conf_toml.tpl $(OUTPUTDIR)/conf.toml
copy_linux_scripts:
	cp scripts/linux/* $(OUTPUTDIR)/
copy_windows_scripts:
	cp scripts/windows/* $(OUTPUTDIR)/

.PHONY: clean
clean:
	rm -rf $(CURDIR)/$(OUTPUTDIR)

.PHONY: rebuild
rebuild: clean build
