# RateMate — Makefile
#
# Run scripts (scripts/run.{sh,ps1,cmd}) set GOPROXY=https://goproxy.cn,direct
# for the duration of the `go run` call ONLY, so module downloads work behind
# the GFW without polluting your persistent environment.
#
#   make run       → launch the TUI via the run script (no binary produced)
#   make serve     → run the proxy headless via the run script
#   make build     → produce one optimised binary in ./bin/
#   make build-all → cross-compile Windows/Linux/macOS × amd64/arm64 into ./dist/
#   make uninstall → remove every trace of RateMate from the system
#
# On Windows without `make`, use the scripts directly:
#   .\scripts\run.ps1          or   scripts\run.cmd
#   .\scripts\uninstall.ps1

BINARY   := ratemate
BIN_DIR  := bin
DIST_DIR := dist
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
GOFLAGS  := -trimpath

# Pick the correct run script for the host OS.
ifeq ($(OS),Windows_NT)
	RUN_SCRIPT   := powershell -ExecutionPolicy Bypass -File scripts/run.ps1
	UNINSTALL_CMD := powershell -ExecutionPolicy Bypass -File scripts/uninstall.ps1
else
	RUN_SCRIPT   := ./scripts/run.sh
	UNINSTALL_CMD := ./scripts/uninstall.sh
endif

TARGETS := \
	linux/amd64   linux/arm64   \
	darwin/amd64  darwin/arm64  \
	windows/amd64 windows/arm64

.PHONY: all run serve build build-all test vet fmt clean install uninstall help dist-clean

all: build

# --- first-class: run in place, no binary, no env pollution -----------------

run:        ## Launch the TUI via the run script (auto GOPROXY, no binary)
	@$(RUN_SCRIPT)

serve:      ## Run the proxy headless via the run script
	@$(RUN_SCRIPT) serve

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

# --- quality ---------------------------------------------------------------

test:       ## Run tests
	go test ./...

vet:        ## Run go vet
	go vet ./...

fmt:        ## Format the code
	gofmt -s -w .

install:    ## Install to $$GOPATH/bin
	go install $(GOFLAGS) -ldflags "$(LDFLAGS)" .

# --- cleanup ---------------------------------------------------------------

clean:      ## Remove ./bin and ./dist
	rm -rf $(BIN_DIR) $(DIST_DIR)

uninstall:  ## Remove EVERY trace of RateMate (config, binary, artifacts)
	@$(UNINSTALL_CMD)

uninstall-purge:  ## Same as uninstall, plus Go module/build cache entries
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/uninstall.ps1 -Purge
else
	@./scripts/uninstall.sh --purge
endif

dist-clean: clean  ## Alias: remove build artifacts only
	rm -f $(BINARY) $(BINARY).exe

help:       ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n",$$1,$$2}'
