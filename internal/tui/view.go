package tui

import (
        "fmt"
        "strings"
        "time"

        "github.com/charmbracelet/lipgloss"

        "github.com/WaZixwx/RateMate/internal/limiter"
        "github.com/WaZixwx/RateMate/internal/stats"
)

// View renders the full screen. Layout (top→bottom):
//
//   1. Title bar with live status badge
//   2. Mode + settings panel (adapts to Auto/Manual)
//   3. Two columns: Live Stats | Recent Activity
//   4. Status line (transient messages)
//   5. Footer with key bindings
//   (help overlay when toggled)
func (m Model) View() string {
        if m.width == 0 {
                return "loading…"
        }

        if m.showHelp {
                return m.viewHelp()
        }

        w := m.width
        var b strings.Builder

        b.WriteString(m.titleBar(w))
        b.WriteString("\n")

        b.WriteString(m.settingsPanel(w))
        b.WriteString("\n")

        // Stats | Activity side by side
        colW := (w - 3) / 2 // 1 gap between columns + borders
        left := m.statsPanel(colW)
        right := m.activityPanel(colW)
        b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right))
        b.WriteString("\n")

        b.WriteString(m.statusLine(w))
        b.WriteString("\n")

        b.WriteString(m.footer(w))

        return lipgloss.JoinVertical(lipgloss.Left, b.String())
}

// ---- title bar -------------------------------------------------------------

func (m Model) titleBar(w int) string {
        title := titleStyle.Render("⚡ RateMate")
        sub := subtitleStyle.Render("LLM API delay proxy")

        left := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", sub)

        // live status badge
        var badgeText, badgeCol = "● Stopped", colorStop
        if m.proxy.Running() {
                if m.lim.Paused() {
                        badgeText, badgeCol = "⏸ Paused", colorWarn
                } else {
                        badgeText, badgeCol = "● Running", colorGo
                }
        }
        addr := m.proxy.Addr()
        badge := lipgloss.NewStyle().
                Foreground(badgeCol).Bold(true).
                Padding(0, 1).
                Render(fmt.Sprintf("%s  %s", badgeText, addr))

        leftW := lipgloss.Width(left)
        rightW := lipgloss.Width(badge)
        pad := w - leftW - rightW
        if pad < 1 {
                pad = 1
        }
        spacer := strings.Repeat(" ", pad)
        return lipgloss.JoinHorizontal(lipgloss.Center, left, spacer, badge)
}

// ---- settings panel --------------------------------------------------------

func (m Model) settingsPanel(w int) string {
        mode := m.limSnap.Mode
        if mode == 0 && m.lim != nil {
                mode = m.lim.Snapshot().Mode
        }

        body := strings.Builder{}
        body.WriteString(m.modeRow(mode))
        body.WriteString("\n")

        if mode == limiter.ModeAuto {
                body.WriteString(m.autoSettings())
        } else {
                body.WriteString(m.manualSettings())
        }

        inner := strings.TrimSpace(body.String())
        title := innerTitle.Render("⚙ Configuration")
        return panel.Width(w - 2).Render(
                lipgloss.JoinVertical(lipgloss.Left, title, "", inner),
        )
}

func (m Model) modeRow(mode limiter.Mode) string {
        autoMark := "○"
        manualMark := "○"
        if mode == limiter.ModeAuto {
                autoMark = "◉"
        } else {
                manualMark = "◉"
        }
        auto := lipgloss.NewStyle().Foreground(colorGo).Bold(true).Render(autoMark+" Auto")
        manual := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(manualMark+" Manual")
        hint := hintStyle.Render("<m> switch")
        return fmt.Sprintf("%s    %s    %s", auto, manual, hint)
}

func (m Model) autoSettings() string {
        c := m.lim.Config()
        spacing := m.lim.Snapshot().SpacingMs

        // Window selector with arrows, highlight when focused
        winLabel := c.Window.String()
        windowLine := fmt.Sprintf("%s  %s  %s  %s",
                labelStyle.Render("Window:"),
                m.focusedToken(focusWindow, "◀"),
                selectedOrDim(focusWindow == focusWindow, winLabel),
                m.focusedToken(focusWindow, "▶"),
        )

        // Limit input
        limitLine := fmt.Sprintf("%s  %s  %s",
                labelStyle.Render("Limit:"),
                m.renderInput(focusLimit, m.limitInput, 8),
                dimStyle.Render(fmt.Sprintf("requests per %s", c.Window.Short())),
        )

        computed := fmt.Sprintf("→ computed delay: %s between requests",
                computedStyle.Render(fmt.Sprintf("%d ms", spacing)))
        rate := fmt.Sprintf("≈ %s effective throughput",
                emphasisStyle.Render(formatRate(spacing)))

        return strings.Join([]string{
                windowLine,
                limitLine,
                dimStyle.Render(strings.Repeat("─", 38)),
                computed,
                rate,
        }, "\n")
}

func (m Model) manualSettings() string {
        c := m.lim.Config()
        spacing := m.lim.Snapshot().SpacingMs
        delayLine := fmt.Sprintf("%s  %s  %s",
                labelStyle.Render("Delay:"),
                m.renderInput(focusManual, m.manualInput, 8),
                dimStyle.Render("ms per request"),
        )
        computed := fmt.Sprintf("→ effective rate: %s",
                computedStyle.Render(fmt.Sprintf("%d req/s max", rateFromMs(spacing))))
        _ = c
        return strings.Join([]string{
                delayLine,
                dimStyle.Render(strings.Repeat("─", 38)),
                computed,
                hintStyle.Render("[+]/[-] or ↑/↓ nudges by 100ms"),
        }, "\n")
}

// ---- stats panel -----------------------------------------------------------

func (m Model) statsPanel(w int) string {
        s := m.limSnap
        rows := []string{
                fmt.Sprintf("%s  %s", labelStyle.Render("Total reqs"), valueStyle.Render(commas(s.TotalRequests))),
                fmt.Sprintf("%s  %s", labelStyle.Render("Delayed"), valueStyle.Render(commas(s.DelayedCount))),
                fmt.Sprintf("%s  %s", labelStyle.Render("Pass-through"), valueStyle.Render(commas(s.PassThrough))),
                fmt.Sprintf("%s  %s", labelStyle.Render("Avg wait"), valueStyle.Render(fmt.Sprintf("%d ms", s.AvgWaitMs()))),
                fmt.Sprintf("%s  %s", labelStyle.Render("In / Out"), valueStyle.Render(fmt.Sprintf("%s / %s", humanBytes(m.bytesIn), humanBytes(m.bytesOut)))),
                fmt.Sprintf("%s  %s", labelStyle.Render("Errors"), valueStyle.Render(commas(m.errCount))),
                fmt.Sprintf("%s  %s", labelStyle.Render("Last req"), valueStyle.Render(timeAgo(s.LastRequestAt))),
        }
        body := strings.Join(rows, "\n")
        title := innerTitle.Render("📊 Live Stats")
        return panel.Width(w - 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// ---- activity panel --------------------------------------------------------

func (m Model) activityPanel(w int) string {
        title := innerTitle.Render("📡 Recent Activity")

        if len(m.recent) == 0 {
                empty := dimStyle.Render("no requests yet")
                return panel.Width(w - 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", empty))
        }

        lines := make([]string, 0, len(m.recent))
        innerW := w - 6 // padding + borders
        for _, e := range m.recent {
                lines = append(lines, formatActivity(e, innerW))
        }
        body := strings.Join(lines, "\n")
        return panel.Width(w - 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// ---- status + footer ------------------------------------------------------

func (m Model) statusLine(w int) string {
        if m.status == "" {
                if m.err != nil {
                        return lipgloss.NewStyle().Foreground(colorStop).Padding(0, 1).Render("✗ " + m.err.Error())
                }
                // helpful idle hint
                s := m.limSnap
                if s.Mode == limiter.ModeAuto {
                        return hintStyle.Padding(0, 1).Render(
                                fmt.Sprintf("auto: %d req / %s  →  %d ms spacing", s.Limit, s.Window.Short(), s.SpacingMs))
                }
                return hintStyle.Padding(0, 1).Render(
                        fmt.Sprintf("manual: %d ms spacing", s.SpacingMs))
        }
        return lipgloss.NewStyle().Foreground(colorAccent).Padding(0, 1).Render(m.status)
}

func (m Model) footer(w int) string {
        keys := []struct{ k, d string }{
                {"s", "start/stop"},
                {"m", "mode"},
                {"p", "pause"},
                {"r", "reset"},
                {"c", "clear"},
                {"+/-", "tune"},
                {"tab", "focus"},
                {"enter", "edit"},
                {"?", "help"},
                {"q", "quit"},
        }
        parts := make([]string, 0, len(keys))
        for _, k := range keys {
                parts = append(parts, fmt.Sprintf("%s %s", keyStyle.Render("<"+k.k+">"), dimStyle.Render(k.d)))
        }
        row := strings.Join(parts, "  ")
        return footerStyle.Width(w).Align(lipgloss.Left).Render(row)
}

// ---- help overlay ---------------------------------------------------------

func (m Model) viewHelp() string {
        lines := []string{
                "⚡ RateMate — LLM API delay proxy",
                "",
                "RateMate runs a local HTTP/HTTPS proxy that injects a small delay",
                "before each forwarded request so you never blow past an LLM API's",
                "per-second / per-minute / per-hour rate limit.",
                "",
                "Usage:",
                "  1. Press <s> to start the proxy (default :8080).",
                "  2. Point your client at the proxy, e.g.:",
                "       export HTTPS_PROXY=http://127.0.0.1:8080",
                "       export HTTP_PROXY=http://127.0.0.1:8080",
                "  3. Tune the delay with <m> (mode) and the inputs. Tab focuses fields,",
                "     Enter edits, Esc cancels, +/- nudges.",
                "  4. Press <p> to pause (pass-through) without stopping the server.",
                "",
                "Keys:",
                "  s          start / stop the proxy",
                "  m          toggle Auto ↔ Manual mode",
                "  p          pause / resume the limiter",
                "  r          reset counters & log",
                "  c          clear activity log only",
                "  ◀/▶ or </> cycle the time window (Auto)",
                "  +/-         tune limit (Auto) or delay (Manual)",
                "  ↑/↓         tune by ±5 (Auto) or ±100ms (Manual)",
                "  tab/shift+tab cycle focus",
                "  enter       edit focused field",
                "  esc         cancel edit / close help",
                "  ?           toggle this help",
                "  q           quit",
                "",
                "Config is persisted to ~/.ratemate/config.json",
                "",
                dimStyle.Render("press ? or esc to return"),
        }
        body := strings.Join(lines, "\n")
        return panel.Width(min(80, m.width-2)).Render(body)
}

// ---- helpers ---------------------------------------------------------------

// renderInput renders a text input (or a read-only value when not focused /
// not editing) consistently with the surrounding chrome.
func (m Model) renderInput(f focus, ti interface{ View() string }, w int) string {
        if m.focus == f && m.editing {
                s := lipgloss.NewStyle().Width(w).Render(ti.View())
                return s
        }
        // show as a framed value to indicate it is editable
        v := m.inputValue(f)
        return lipgloss.NewStyle().
                Foreground(colorFg).Bold(true).
                Background(colorBg).
                Padding(0, 1).
                Render(v)
}

func (m Model) inputValue(f focus) string {
        switch f {
        case focusLimit:
                return m.limitInput.Value()
        case focusManual:
                return m.manualInput.Value()
        case focusPort:
                return m.portInput.Value()
        }
        return ""
}

func (m Model) focusedToken(f focus, tok string) string {
        if m.focus == f {
                return selectedStyle.Render(tok)
        }
        return dimStyle.Render(tok)
}

func selectedOrDim(selected bool, s string) string {
        if selected {
                return selectedStyle.Render(s)
        }
        return valueStyle.Render(s)
}

func formatActivity(e stats.Entry, w int) string {
        t := e.Time.Format("15:04:05")
        method := padRight(e.Method, 7)
        host := truncate(e.Host, 28)
        wait := ""
        if e.WaitedMs > 0 {
                wait = dimStyle.Render(fmt.Sprintf(" +%dms", e.WaitedMs))
        } else {
                wait = dimStyle.Render(" ·")
        }
        // color by status
        var statusCol lipgloss.AdaptiveColor = colorDim
        if e.Status >= 200 && e.Status < 300 {
                statusCol = colorGo
        } else if e.Status >= 400 || e.Err != "" {
                statusCol = colorStop
        }
        stamp := lipgloss.NewStyle().Foreground(statusCol).Render(method)
        return fmt.Sprintf("%s %s %s%s",
                dimStyle.Render(t), stamp, lipgloss.NewStyle().Foreground(colorFg).Render(host), wait)
}

// ---- formatting ------------------------------------------------------------

func commas(n int64) string {
        if n < 0 {
                return "-" + commas(-n)
        }
        s := fmt.Sprintf("%d", n)
        // insert thousands separators
        pre := len(s) % 3
        if pre == 0 && len(s) > 0 {
                pre = 3
        }
        var b strings.Builder
        if pre < len(s) {
                b.WriteString(s[:pre])
                for i := pre; i < len(s); i += 3 {
                        b.WriteString(",")
                        b.WriteString(s[i : i+3])
                }
        } else {
                b.WriteString(s)
        }
        return b.String()
}

func humanBytes(n int64) string {
        const k = 1024
        if n < k {
                return fmt.Sprintf("%dB", n)
        }
        div, exp := int64(k), 0
        for x := n / k; x >= k; x /= k {
                div *= k
                exp++
        }
        return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func timeAgo(t time.Time) string {
        if t.IsZero() {
                return dimStyle.Render("—")
        }
        d := time.Since(t).Round(time.Second)
        if d < 0 {
                d = 0
        }
        return d.String() + " ago"
}

func formatRate(spacingMs int64) string {
        if spacingMs <= 0 {
                return "unlimited"
        }
        // requests per minute equivalent
        rpm := float64(60_000) / float64(spacingMs)
        if rpm >= 1 {
                return fmt.Sprintf("%.0f req/min", rpm)
        }
        rph := float64(3_600_000) / float64(spacingMs)
        return fmt.Sprintf("%.0f req/h", rph)
}

func rateFromMs(ms int64) int {
        if ms <= 0 {
                return 0
        }
        return int(1000 / ms)
}

func padRight(s string, n int) string {
        if len(s) >= n {
                return s[:n]
        }
        return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
        if len(s) <= n {
                return s
        }
        return s[:n-1] + "…"
}

func min(a, b int) int {
        if a < b {
                return a
        }
        return b
}
