HELM_HOME ?= $(shell helm home)
BINARY_NAME ?= inj
PLUGIN_NAME ?= helm-inject
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-inject
HAS_GLIDE := $(shell command -v glide;)
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION}"

.PHONY: install
install: bootstrap build
	cp $(BINARY_NAME) $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

.PHONY: hookInstall
hookInstall: bootstrap build

.PHONY: build
build:
	go build -o $(BINARY_NAME) -ldflags $(LDFLAGS) ./main.go

.PHONY: dist
dist:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/$(PLUGIN_NAME)-linux-$(VERSION).tgz $(BINARY_NAME) README.md LICENSE.txt plugin.yaml
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME) -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/$(PLUGIN_NAME)-macos-$(VERSION).tgz $(BINARY_NAME) README.md LICENSE.txt plugin.yaml
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/$(PLUGIN_NAME)-windows-$(VERSION).tgz $(BINARY_NAME).exe README.md LICENSE.txt plugin.yaml

.PHONY: bootstrap
bootstrap:
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
	glide install --strip-vendor
