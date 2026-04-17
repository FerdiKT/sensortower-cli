BINARY_NAME ?= sensortower
VERSION ?= dev
DIST_DIR ?= dist

.PHONY: build run tidy build-all checksums dist clean test brew-dist

build:
	go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o bin/$(BINARY_NAME) .

run:
	go run .

tidy:
	go mod tidy

test:
	go test ./...

build-all: clean
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64 .

checksums:
	cd $(DIST_DIR) && shasum -a 256 * > checksums.txt

dist: build-all checksums
	@echo "dist artifacts ready in $(DIST_DIR)"

brew-dist: clean
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/ferdikt/sensortower-cli/internal/buildinfo.Version=$(VERSION)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "brew-dist artifacts ready in $(DIST_DIR)"

clean:
	rm -rf $(DIST_DIR) bin
