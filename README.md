# ⚡ RateMate

> A tiny, beautiful terminal rate-limiter / delay proxy for LLM APIs.
> One codebase. Three platforms. No binary required.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-windows%20%7C%20linux%20%7C%20macOS-blue)](#cross-platform)

RateMate runs a local HTTP/HTTPS forward proxy that injects a small, computed
delay before every forwarded request so you never blow past an LLM API's
per-second / per-minute / per-hour rate limit. Configure it once, point your
client at it, and forget.

- **Two modes.** *Auto*: pick a window (second / minute / hour) + a request
  limit, RateMate computes the even-spacing delay for you. *Manual*: type the
  delay in ms directly.
- **System-wide.** Anything that honours `HTTPS_PROXY` / `HTTP_PROXY` is
  covered — CLI tools, SDKs, whole shells, even your IDE.
- **Cross-platform.** Identical code builds & runs on **Windows, Linux, macOS**
  for both **amd64** and **arm64**.
- **No binary required.** `go run .` launches the full TUI with zero build
  artifacts polluting your tree. (A `make build` is there if you want one.)
- **Beautiful TUI.** Live stats, recent-activity feed, computed-throughput
  readout, inline editing, help overlay.
- **One binary, ~7 MB.** No runtime deps, no daemon scaffolding.
- **Headless mode.** `ratemate serve` runs the proxy without a TUI — great for
  CI, containers, or background use.
- **MIT licensed.**

---

## Quick start

### Option A — run it now, no build (simplest)

```bash
# requires Go 1.22+ installed
git clone https://github.com/WaZixwx/RateMate.git
cd RateMate
go run .            # launches the TUI
```

Inside the TUI, press **`s`** to start the proxy (default `:8080`), then point
your client at it:

```bash
# Linux / macOS
export HTTPS_PROXY=http://127.0.0.1:8080
export HTTP_PROXY=http://127.0.0.1:8080

# Windows (PowerShell)
$env:HTTPS_PROXY = "http://127.0.0.1:8080"
$env:HTTP_PROXY  = "http://127.0.0.1:8080"

# Windows (cmd)
set HTTPS_PROXY=http://127.0.0.1:8080
set HTTP_PROXY=http://127.0.0.1:8080
```

Any HTTPS request is now spaced by your configured delay:

```bash
curl https://api.openai.com/v1/models
```

### Option B — build a binary

```bash
make build          # → ./bin/ratemate  (optimised, ~7 MB)
# or, without make:
go build -trimpath -ldflags "-s -w" -o ratemate .
./ratemate
```

### Option C — install to `$GOPATH/bin`

```bash
make install        # or: go install ./...
ratemate            # now on your PATH
```

---

## Cross-platform

RateMate is pure Go using only cross-platform standard library packages
(`net`, `net/http`, `os/signal`) plus [bubbletea](https://github.com/charmbracelet/bubbletea)
which itself ships Windows console input support. Nothing in the codebase is
platform-specific.

Build for every supported target from any host:

```bash
make build-all      # → ./dist/ratemate-{os}-{arch}[.exe]
```

Produces all six binaries:

| OS      | amd64 | arm64 |
|---------|-------|-------|
| linux   | ✅ 7.3 MB | ✅ 6.9 MB |
| darwin  | ✅ 7.4 MB | ✅ 7.0 MB |
| windows | ✅ 7.5 MB | ✅ 6.9 MB |

Or build a single target manually:

```bash
# from Linux/macOS
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o ratemate.exe .

# from Windows (cmd)
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o ratemate .
```

---

## The two modes

### Auto — "I have a rate limit, you do the math"

Pick a time window and a request budget. RateMate computes
`delay = window / limit` and enforces even spacing between successive
forwarded requests. This is the safest way to stay under a hard RPM/RPH cap.

```
Window:  ◀  Minute  ▶
Limit:   50   requests per m
→ computed delay: 1200 ms between requests
≈ 50 req/min effective throughput
```

### Manual — "I know exactly how long to wait"

Type a millisecond delay. Every request waits that long before being forwarded.

```
Delay:   1000   ms per request
→ effective rate: 1 req/s max
```

---

## Keybindings

| Key          | Action                                        |
|--------------|-----------------------------------------------|
| `s`          | start / stop the proxy                         |
| `m`          | toggle Auto ↔ Manual mode                     |
| `p`          | pause / resume the limiter (pass-through)     |
| `◀` `▶`      | cycle the time window (Auto)                  |
| `+` `-`      | tune limit (Auto) or delay (Manual)           |
| `↑` `↓`      | tune by ±5 (Auto) or ±100 ms (Manual)         |
| `tab`        | cycle focus between editable fields           |
| `enter`      | edit the focused field                        |
| `r`          | reset counters & activity log                  |
| `c`          | clear the activity log only                    |
| `?`          | toggle the help overlay                        |
| `q` / `esc`  | quit (stops the proxy first if running)        |

---

## Headless / daemon mode

```bash
# uses the persisted config
go run . serve

# override on the command line
go run . serve -port 9090 -mode manual
```

Logs one line per proxied activity to stdout, blocks until `Ctrl+C` / SIGTERM.

---

## Configuration

Stored at `~/.ratemate/config.json` (Linux/macOS) or
`%USERPROFILE%\.ratemate\config.json` (Windows). The TUI writes it on every
change; the headless mode reads it on start.

```json
{
  "mode": "auto",
  "window": "minute",
  "limit": 50,
  "manual_delay_ms": 1000,
  "port": 8080,
  "autostart": false,
  "log_body_bytes": false
}
```

---

## How it works

RateMate is a standard HTTP forward proxy. For HTTPS targets it speaks the
`CONNECT` method, dials the target, replies `200 Connection Established`, and
tunnels bytes in both directions — applying the limiter's delay *before* the
dial. For plain HTTP it rewrites and forwards the request.

The limiter enforces **even spacing**: each request reserves a release slot
`delay` ms after the previous one, so concurrent callers queue behind each
other rather than bursting. The first request after a quiet period goes
through immediately; subsequent ones wait the full spacing.

---

## Makefile targets

```
make run        Launch the TUI via `go run .`         (no binary produced)
make serve      Run the proxy headless via `go run . serve`
make build      Build an optimised binary into ./bin/
make build-all  Cross-compile all 6 platform/arch combos into ./dist/
make test       Run tests
make vet        Run go vet
make fmt        Format the code
make install     Install to $GOPATH/bin
make clean      Remove ./bin and ./dist
make help       Show all targets
```

No `make` on Windows? Every target maps directly to a `go` command — see the
header comments in [`Makefile`](Makefile).

---

## Project layout

```
RateMate/
├── main.go                  # entry: TUI + `serve` subcommand + help
├── go.mod / go.sum
├── Makefile                 # run / build / build-all / test / vet
├── LICENSE                  # MIT
└── internal/
    ├── config/              # JSON persistence (cross-platform home dir)
    ├── limiter/             # rate-limit / delay core (auto + manual)
    ├── proxy/               # HTTP/HTTPS forward proxy with delay injection
    ├── stats/               # bounded activity ring buffer + aggregates
    └── tui/                 # bubbletea TUI (model / update / view / styles)
```

---

## Contributing

PRs welcome. Keep it small and focused — RateMate deliberately avoids feature
bloat. Run `make vet test` before submitting.

---

## License

[MIT](LICENSE) — © WaZixwx
