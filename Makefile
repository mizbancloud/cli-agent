.PHONY: build clean install

VERSION=0.1.0
BINARY=mizban
BUILD_DIR=bin

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/mizbancloud/cli/pkg/config.Version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY) ./cmd/mizban

clean:
	@rm -rf $(BUILD_DIR)

install: build
	@cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed $(BINARY) to /usr/local/bin"

windows:
	@echo "Building $(BINARY) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/mizbancloud/cli/pkg/config.Version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY).exe ./cmd/mizban
