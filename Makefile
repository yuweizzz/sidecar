MAIN := main.go
SIDECAR := sidecar
LD_FLAGS := "-s -w"

CURDIR := $(shell pwd)
OUTPUTDIR := build
SIDECAR_DIR := $(CURDIR)/cmd/$(SIDECAR)

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
	GOARCH="amd64" GOOS="linux" go build -ldflags=$(LD_FLAGS) -o $(OUTPUTDIR)/$(SIDECAR) $(SIDECAR_DIR)/$(MAIN)
build_windows:
	GOARCH="amd64" GOOS="windows" go build -ldflags=$(LD_FLAGS) -o $(OUTPUTDIR)/sidecar.exe $(SIDECAR_DIR)/$(MAIN)
build_mac:
	GOARCH="amd64" GOOS="darwin" go build -ldflags=$(LD_FLAGS) -o $(OUTPUTDIR)/$(SIDECAR) $(SIDECAR_DIR)/$(MAIN)
copy_tpl:
	cp config_toml.tpl $(OUTPUTDIR)/config.toml
copy_linux_scripts:
	cp scripts/linux/* $(OUTPUTDIR)/
copy_windows_scripts:
	cp scripts/windows/* $(OUTPUTDIR)/

.PHONY: clean
clean:
	rm -rf $(CURDIR)/$(OUTPUTDIR)

.PHONY: rebuild
rebuild: clean build
