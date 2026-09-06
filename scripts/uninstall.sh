#!/usr/bin/env bash
#
# scripts/uninstall.sh — remove every trace of RateMate from a Linux/macOS
# system. Root-and-branch: config, binaries, build artifacts, and (optionally)
# Go module-cache entries. Safe to re-run; nothing happens if a piece is
# already gone.
#
# Usage:
#   ./scripts/uninstall.sh              # standard purge
#   ./scripts/uninstall.sh --purge      # also wipe RateMate's Go module
#                                       # cache + build cache entries
#   ./scripts/uninstall.sh --dry-run    # show what WOULD be removed, change nothing
#
# This script does NOT touch:
#   - your shell's HTTPS_PROXY / HTTP_PROXY (it only tells you how to unset
#     them, since they live in the parent shell and can't be unset from here)
#   - other Go projects' module cache
#   - your Go installation itself
set -uo pipefail

PURGE=0
DRY_RUN=0
for arg in "$@"; do
    case "$arg" in
        --purge)   PURGE=1 ;;
        --dry-run) DRY_RUN=1 ;;
        -h|--help)
            sed -n '2,20p' "$0"
            exit 0
            ;;
        *) echo "unknown flag: $arg (try --help)" >&2; exit 2 ;;
    esac
done

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
info() { printf '  %s\n' "$*"; }
gone() { printf '  \033[32mremoved\033[0m  %s\n' "$*"; }
skip() { printf '  \033[2mabsent\033[0m   %s\n' "$*"; }
warn() { printf '  \033[33mwarn\033[0m    %s\n' "$*"; }

run() {
    if [ "$DRY_RUN" = 1 ]; then
        printf '  \033[2mwould-run\033[0m %s\n' "$*"
    else
        "$@"
    fi
}
rmrf() {
    if [ -e "$1" ]; then
        if [ "$DRY_RUN" = 1 ]; then
            printf '  \033[2mwould-rm\033[0m  %s\n' "$1"
        else
            rm -rf "$1"
            gone "$1"
        fi
    else
        skip "$1"
    fi
}

bold "=== RateMate uninstall (root-and-branch) ==="
[ "$DRY_RUN" = 1 ] && info "(dry-run mode — nothing will actually be deleted)"
echo ""

# ---- 1. config directory: ~/.ratemate -------------------------------------
bold "1. configuration"
CFG_DIR="$HOME/.ratemate"
if [ -d "$CFG_DIR" ]; then
    rmrf "$CFG_DIR"
else
    skip "$CFG_DIR"
fi
echo ""

# ---- 2. installed binary (go install) -------------------------------------
bold "2. installed binary"
GOPATH="${GOPATH:-$HOME/go}"
BIN="$GOPATH/bin/ratemate"
if [ -f "$BIN" ]; then
    rmrf "$BIN"
else
    skip "$BIN"
fi
echo ""

# ---- 3. build artifacts in the repo (bin/, dist/) -------------------------
bold "3. in-repo build artifacts"
for d in bin dist; do
    rmrf "./$d"
done
echo ""

# ---- 4. environment variables: detect + instruct --------------------------
# We cannot unset the parent shell's env vars from a child script, so we
# detect and instruct instead.
bold "4. proxy environment variables"
found_env=0
for v in HTTPS_PROXY HTTP_PROXY; do
    val="${!v:-}"
    if [ -n "$val" ] && printf '%s' "$val" | grep -q '127\.0\.0\.1:8080'; then
        warn "$v=$val is set in the current shell"
        info "  remove it with:  unset $v"
        found_env=1
    fi
done
# Also check ~/.bashrc / ~/.zshrc for a persisted assignment (best-effort).
for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" "$HOME/.bash_profile"; do
    if [ -f "$rc" ] && grep -qE 'HTTPS_PROXY|HTTP_PROXY' "$rc" 2>/dev/null; then
        warn "$rc mentions HTTPS_PROXY/HTTP_PROXY — check it's not RateMate-specific:"
        info "    grep -n PROXY \"$rc\""
        found_env=1
    fi
done
[ "$found_env" = 0 ] && info "no RateMate proxy vars detected"
echo ""

# ---- 5. optional: Go module + build cache ---------------------------------
if [ "$PURGE" = 1 ]; then
    bold "5. Go module/build cache (--purge)"

    if command -v go >/dev/null 2>&1; then
        GOMODCACHE="$(go env GOMODCACHE)"
        GOCACHE="$(go env GOCACHE)"
        # Go encodes uppercase letters in module paths as ! + lowercase.
        # github.com/WaZixwx  ->  github.com/!wa!zixwx
        MOD_DIR="$GOMODCACHE/github.com/!wa!zixwx"
        rmrf "$MOD_DIR"
        # build cache: blow the whole thing (regenerated on next compile,
        # safe — only affects compile speed of other Go projects briefly).
        if [ -d "$GOCACHE" ]; then
            if [ "$DRY_RUN" = 1 ]; then
                printf '  \033[2mwould-run\033[0m go clean -cache\n'
            else
                go clean -cache 2>/dev/null && gone "go build cache ($GOCACHE)"
            fi
        else
            skip "$GOCACHE"
        fi
    else
        warn "go not found on PATH — skipping cache purge"
    fi
    echo ""
fi

bold "=== done. RateMate is fully removed. ==="
[ "$found_env" = 1 ] 2>/dev/null && {
    echo ""
    info "Note: a couple of env vars live in your shell and can only be"
    info "unset from there. Run the 'unset' commands above and you're done."
}
