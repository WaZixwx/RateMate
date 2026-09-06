# RateMate — Makefile
#
# Default workflow favours `go run` (no binary pollution) as requested:
#   make run       → launch the TUI immediately via `go run .`
#   make serve     → run the proxy headless via `go run . serve`
#   make build     → produce one optimised binary in ./bin/
#   make build-all → cross-compile Windows/Linux/macOS × amd64/arm64 into ./dist/
#
# On Windows without `make`, the equivalent commands are just:
#   go run .
#   go run . serve
#   go build -trimpath -ldflags "-s -w" -o ratemate.exe .

BINARY   := ratemate
BIN_DIR  := bin
DIST_DIR := dist
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
GOFLAGS  := -trimpath

TARGETS := \
	linux/amd64   linux/arm64   \
	darwin/amd64  darwin/arm64  \
	windows/amd64 windows/arm64

.PHONY: all run serve build build-all test vet fmt clean install help dist-clean

all: build

# --- first-class: run in place, no binary ----------------------------------

run:        ## Launch the TUI via `go run .` (no binary produced)
	go run .

serve:      ## Run the proxy headless via `go run . serve`
	go run . serve

# --- building --------------------------------------------------------------

build:      ## Build an optimised binary into ./bin/
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) .

build-all:  ## Cross-compile all 6 platform/arch combos into ./dist/
	@mkdir -p $(DIST_DIR)
	@for target in $(TARGETS); do \
		os=$${target%/*}; arch=$${target#*/}; \
		ext=""; [ $$os = windows ] && ext=".exe"; \
		echo "  → $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
			-o $(DIST_DIR)/$(BINARY)-$$os-$$arch$$ext . || exit 1; \
	done
	@ls -lh $(DIST_DIR)

# --- quality --------------------------------------------------------------

test:       ## Run tests
	go test ./...

vet:        ## Run go vet
	go vet ./...

fmt:        ## Format the code
	gofmt -s -w .

install:    ## Install to $$GOPATH/bin
	go install $(GOFLAGS) -ldflags "$(LDFLAGS)" .

# --- cleanup --------------------------------------------------------------

clean:      ## Remove ./bin and ./dist
	rm -rf $(BIN_DIR) $(DIST_DIR)

dist-clean: clean  ## Also drop the local config and any stray binaries
	rm -f $(BINARY) $(BINARY).exe
	rm -rf ~/.ratemate

help:       ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n",$$1,$$2}'
