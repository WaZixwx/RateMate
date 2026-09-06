#!/usr/bin/env bash
#
# scripts/run.sh — launch RateMate (TUI by default) via `go run .`
#
# Temporarily routes Go module downloads through a China-friendly mirror
# (goproxy.cn) so `go run` works behind the GFW without touching the user's
# persistent environment. The GOPROXY value is scoped to the `go run` process
# ONLY — it never leaks into the calling shell, never gets written to
# `go env`, and disappears the moment `go run` exits.
#
# Usage:
#   ./scripts/run.sh              # launch the TUI
#   ./scripts/run.sh serve       # run the proxy headless
#   ./scripts/run.sh serve -port 9090 -mode manual
#   ./scripts/run.sh version
#
# Works on Linux and macOS. For Windows use run.ps1 or run.cmd instead.
set -euo pipefail

# Resolve the project root (parent of this script's directory) so the script
# works no matter where it's invoked from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Hand off to `go run` with a process-local GOPROXY. `exec` replaces the
# shell, so there's no leftover state to clean up — the GOPROXY dies with the
# go process.
exec env GOPROXY=https://goproxy.cn,direct go run . "$@"
