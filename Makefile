# Armazenda Build Makefile
# Usage:
#   make              Run full pipeline (test + build)
#   make test         Run all tests
#   make test-unit    Run Go unit tests only
#   make test-e2e     Run E2E tests only
#   make build        Build everything (CSS + WASM + Go)
#   make build-go     Build only the main Go binary
#   make clean        Remove build artifacts

# --- Variables ---
TAILWIND ?= tailwind
GO       ?= go
BUN      ?= bun
OUTPUT_DIR ?= ./tmp

# --- Default target ---
.PHONY: all test test-unit test-e2e build build-css build-wasm build-go clean

all: test build

# --- Test targets ---
test: test-unit test-e2e

test-unit:
	$(GO) test ./service/entry_service/test/
	$(GO) test ./pkg/calculator/

test-e2e:
	cd test && $(BUN) run test:e2e

# --- Build targets ---
build: build-css build-wasm build-go

build-css:
	$(TAILWIND) -i assets/static/css/input.css -o assets/static/css/output.css

build-wasm:
	GOOS=js GOARCH=wasm $(GO) build -o ./assets/wasm/calculator.wasm ./pkg/calculator/wasm

build-go:
	$(GO) build -o $(OUTPUT_DIR)/main .

# --- Cleanup ---
clean:
	rm -f $(OUTPUT_DIR)/main ./assets/wasm/calculator.wasm ./assets/static/css/output.css
