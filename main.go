// Command ratemate is a tiny terminal rate-limiter / delay proxy for LLM APIs.
//
// It is fully cross-platform — one codebase builds for Windows, Linux and
// macOS on amd64 and arm64. Run with no arguments to launch the TUI; the
// proxy listens on the port saved in ~/.ratemate/config.json (default :8080)
// and can be started/stopped from inside the UI.
//
// `ratemate serve` runs the proxy headless (useful as a daemon or in CI),
// reading its configuration from the persisted config file.
//
// Prefer not to build at all? `go run .` works on every platform.
package main

import (
        "context"
        "flag"
        "fmt"
        "log"
        "os"
        "os/signal"
        "runtime"
        "syscall"

        tea "github.com/charmbracelet/bubbletea"

        "github.com/WaZixwx/RateMate/internal/config"
        "github.com/WaZixwx/RateMate/internal/limiter"
        "github.com/WaZixwx/RateMate/internal/proxy"
        "github.com/WaZixwx/RateMate/internal/stats"
        "github.com/WaZixwx/RateMate/internal/tui"
)

// version is injected at build time via -ldflags "-X main.version=...".
// When running via `go run .` it stays at the dev sentinel.
var version = "dev"

func main() {
        if len(os.Args) > 1 {
                switch os.Args[1] {
                case "-h", "--help", "help":
                        os.Stdout.WriteString(usage)
                        return
                case "serve":
                        runServe(os.Args[2:])
                        return
                case "version", "-v", "--version":
                        fmt.Printf("ratemate %s (%s/%s)\n", version, osGOOS(), arch())
                        return
                }
        }

        p := tea.NewProgram(tui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
        if _, err := p.Run(); err != nil {
                fmt.Fprintf(os.Stderr, "ratemate: %v\n", err)
                os.Exit(1)
        }
}

func osGOOS() string { return runtime.GOOS }
func arch() string   { return runtime.GOARCH }

// runServe runs the proxy headless. It loads the persisted config, starts the
// proxy and blocks until SIGINT/SIGTERM.
func runServe(args []string) {
        fs := flag.NewFlagSet("serve", flag.ExitOnError)
        port := fs.Int("port", 0, "override listen port (0 = use config)")
        mode := fs.String("mode", "", "override mode: auto|manual")
        _ = fs.Parse(args)

        cfg := config.Load()
        if *port > 0 {
                cfg.Port = *port
        }
        lc := cfg.ToLimiterConfig()
        if *mode == "auto" || *mode == "manual" {
                cfg.Mode = *mode
                lc = cfg.ToLimiterConfig()
        }

        lim := limiter.New(lc)
        tr := stats.New(200)
        logger := log.New(os.Stdout, "", log.LstdFlags)
        addr := proxy.ParseAddr(cfg.Port)
        srv := proxy.New(addr, lim, tr, logger)

        if err := srv.Start(); err != nil {
                fmt.Fprintf(os.Stderr, "start: %v\n", err)
                os.Exit(1)
        }
        logger.Printf("ratemate serve on %s  mode=%s  limit=%d  manual=%dms  spacing=%dms",
                addr, lc.Mode, lc.Limit, lc.ManualDelayMs, lim.Snapshot().SpacingMs)

        // Wait for signal.
        ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
        defer cancel()
        <-ctx.Done()
        logger.Println("shutting down…")
        srv.Stop()
}

const usage = `ratemate — a terminal rate-limiter / delay proxy for LLM APIs.
Cross-platform: one codebase for Windows, Linux and macOS (amd64 + arm64).

Usage:
  ratemate                    Launch the TUI (default).
  ratemate serve [flags]       Run the proxy headless.
  ratemate -h                  Show this help.
  ratemate version             Print version + platform.

Run without building at all:
  go run .                     Launch the TUI.
  go run . serve               Run headless.

serve flags:
  -port int    override listen port (0 = use config)
  -mode str    override mode: auto|manual

Inside the TUI:
  s          start / stop the proxy
  m          toggle Auto ↔ Manual mode
  p          pause / resume the limiter (pass-through)
  l          open the language switcher (zh, en, ja, es, fr, de)
  <-/->      cycle time window (Auto)
  +/-        tune limit (Auto) or delay (Manual)
  up/dn      tune by +-5 (Auto) or +-100ms (Manual)
  tab        cycle focus between editable fields
  enter      edit focused field
  r          reset counters
  c          clear activity log
  ?          toggle help
  q          quit

Point your client at the proxy (pick the one for your shell):

  Linux / macOS:
    export HTTPS_PROXY=http://127.0.0.1:8080
    export HTTP_PROXY=http://127.0.0.1:8080

  Windows (cmd):
    set HTTPS_PROXY=http://127.0.0.1:8080
    set HTTP_PROXY=http://127.0.0.1:8080

  Windows (PowerShell):
    $env:HTTPS_PROXY = "http://127.0.0.1:8080"
    $env:HTTP_PROXY  = "http://127.0.0.1:8080"

Config is stored at:
  Linux/macOS: ~/.ratemate/config.json
  Windows:    %USERPROFILE%\.ratemate\config.json
`
