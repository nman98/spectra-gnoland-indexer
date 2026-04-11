.PHONY: build 
		install 
		clean 
		build-experimental 
		install-experimental 
		build-api 
		integration-test 
		test 
		vulnerability-scan 
		snyk 
		semgrep 
		code-quality

########################################################
# Build and install the indexer
########################################################


# Get git information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
VERSION := $(if $(GIT_TAG),$(GIT_TAG),$(GIT_BRANCH)-$(GIT_COMMIT))

build-indexer:
	mkdir -p build
	go build -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION) -w -s" -o build/indexer indexer/indexer.go

build-api:
	mkdir -p build
	go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION) -w -s" -o build/api ./api

clean:
	rm -rf build

########################################################
# Test the indexer
########################################################

test:
	go test -v ./...

integration-test:
	cd integration && go test -v -tags=integration -timeout=20m ./...

########################################################
# Vulnerability scanning
########################################################

vulnerability-scan:
	govulncheck ./...

snyk:
	snyk test 

semgrep:
	semgrep ci

########################################################
# Code quality
########################################################

lint:
	golangci-lint run

vulncheck:
	govulncheck ./...

########################################################
# Train the zstd dictionary
########################################################

.PHONY: train-zstd

train-zstd:
	@echo "Training the zstd dictionary"
	@read -p "Enter the amount of events to collect (default: 10000): " amount; \
	amount=$${amount:-10000}; \
	go run compression/cmd/main.go --config training-config.yml --amount $$amount --chain-name gnoland --dict-path ./pkgs/dict_loader/events.zstd.bin
