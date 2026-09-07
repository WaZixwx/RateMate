package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/WaZixwx/RateMate/internal/i18n"
	"github.com/WaZixwx/RateMate/internal/limiter"
	"github.com/WaZixwx/RateMate/internal/stats"
)

// View renders the full screen. Layout (top→bottom):
//
//  1. Title bar with live status badge + prominent language button
//  2. Mode + settings panel (adapts to Auto/Manual)
//  3. Two columns: Live Stats | Recent Activity
//  4. Status line (transient messages)
//  5. Footer with key bindings
//     (help overlay when toggled)
//     (language switcher overlay when toggled)
func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}

	if m.showHelp {
		return m.viewHelp()
	}
	if m.showLang {
		return m.viewLangMenu()
	}
	return m.viewMain()
}

// viewMain renders the ordinary dashboard.
func (m Model) viewMain() string {
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
	sub := subtitleStyle.Render(m.tr.AppSubtitle)

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", sub)

	// live status badge
	badgeText := m.tr.StatusStopped
	badgeCol := colorStop
	if m.proxy.Running() {
		if m.lim.Paused() {
			badgeText, badgeCol = m.tr.StatusPaused, colorWarn
		} else {
			badgeText, badgeCol = m.tr.StatusRunning, colorGo
		}
	}
	addr := m.proxy.Addr()
	badge := lipgloss.NewStyle().
		Foreground(badgeCol).Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("%s  %s", badgeText, addr))

	// prominent language switcher button
	langBtn := m.langButton()

	right := lipgloss.JoinHorizontal(lipgloss.Center, langBtn, "  ", badge)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := w - leftW - rightW
	if pad < 1 {
		pad = 1
	}
	spacer := strings.Repeat(" ", pad)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, spacer, right)
}

// langButton renders the prominent, always-visible language switcher button
// in the title bar. It reads as "clickable" (press <l>) and highlights when the
// switcher menu is open.
func (m Model) langButton() string {
	info := i18n.Info(m.lang)
	label := fmt.Sprintf(m.tr.LangButtonFmt, info.Native) + " ▾"
	if m.showLang {
		return langButtonActiveStyle.Render(label)
	}
	return langButtonStyle.Render(label)
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
	title := innerTitle.Render(m.tr.ConfigTitle)
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
	auto := lipgloss.NewStyle().Foreground(colorGo).Bold(true).Render(autoMark + " " + m.tr.ModeAuto)
	manual := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(manualMark + " " + m.tr.ModeManual)
	hint := hintStyle.Render(m.tr.ModeSwitchHint)
	return fmt.Sprintf("%s    %s    %s", auto, manual, hint)
}

func (m Model) autoSettings() string {
	c := m.lim.Config()
	spacing := m.lim.Snapshot().SpacingMs

	// Window selector with arrows, highlight when focused
	winLabel := m.tr.WindowName(int(c.Window))
	windowLine := fmt.Sprintf("%s  %s  %s  %s",
		labelStyle.Render(m.tr.WindowLabel),
		m.focusedToken(focusWindow, "◀"),
		selectedOrDim(focusWindow == focusWindow, winLabel),
		m.focusedToken(focusWindow, "▶"),
	)

	// Limit input
	limitLine := fmt.Sprintf("%s  %s  %s",
		labelStyle.Render(m.tr.LimitLabel),
		m.renderInput(focusLimit, m.limitInput, 8),
		dimStyle.Render(fmt.Sprintf(m.tr.RequestsPerShortFmt, m.tr.WindowShort(int(c.Window)))),
	)

	computed := fmt.Sprintf(m.tr.ComputedDelayFmt,
		computedStyle.Render(m.fmtMs(spacing)))
	rate := fmt.Sprintf(m.tr.EffectiveThroughputFmt,
		emphasisStyle.Render(m.formatRate(spacing)))

	return strings.Join([]string{
		windowLine,
		limitLine,
		dimStyle.Render(strings.Repeat("─", 38)),
		computed,
		rate,
	}, "\n")
}

func (m Model) manualSettings() string {
	spacing := m.lim.Snapshot().SpacingMs
	delayLine := fmt.Sprintf("%s  %s  %s",
		labelStyle.Render(m.tr.DelayLabel),
		m.renderInput(focusManual, m.manualInput, 8),
		dimStyle.Render(m.tr.MsPerRequest),
	)
	computed := fmt.Sprintf(m.tr.EffectiveRateFmt,
		computedStyle.Render(fmt.Sprintf(m.tr.ReqPerSecMaxFmt, rateFromMs(spacing))))
	return strings.Join([]string{
		delayLine,
		dimStyle.Render(strings.Repeat("─", 38)),
		computed,
		hintStyle.Render(m.tr.NudgeHint),
	}, "\n")
}

// ---- stats panel -----------------------------------------------------------

func (m Model) statsPanel(w int) string {
	s := m.limSnap
	rows := []string{
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.TotalReqs), valueStyle.Render(commas(s.TotalRequests))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.Delayed), valueStyle.Render(commas(s.DelayedCount))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.PassThrough), valueStyle.Render(commas(s.PassThrough))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.AvgWait), valueStyle.Render(m.fmtMs(s.AvgWaitMs()))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.InOut), valueStyle.Render(fmt.Sprintf("%s / %s", humanBytes(m.bytesIn), humanBytes(m.bytesOut)))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.Errors), valueStyle.Render(commas(m.errCount))),
		fmt.Sprintf("%s  %s", labelStyle.Render(m.tr.LastReq), valueStyle.Render(m.timeAgo(s.LastRequestAt))),
	}
	body := strings.Join(rows, "\n")
	title := innerTitle.Render(m.tr.StatsTitle)
	return panel.Width(w - 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// ---- activity panel --------------------------------------------------------

func (m Model) activityPanel(w int) string {
	title := innerTitle.Render(m.tr.ActivityTitle)

	if len(m.recent) == 0 {
		empty := dimStyle.Render(m.tr.NoRequests)
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
				fmt.Sprintf(m.tr.AutoStatusFmt, s.Limit, m.tr.WindowShort(int(s.Window)), s.SpacingMs))
		}
		return hintStyle.Padding(0, 1).Render(
			fmt.Sprintf(m.tr.ManualStatusFmt, s.SpacingMs))
	}
	return lipgloss.NewStyle().Foreground(colorAccent).Padding(0, 1).Render(m.status)
}

func (m Model) footer(w int) string {
	keys := []struct{ k, d string }{
		{"s", m.tr.KeyStartStop},
		{"m", m.tr.KeyMode},
		{"p", m.tr.KeyPause},
		{"l", m.tr.KeyLang},
		{"r", m.tr.KeyReset},
		{"c", m.tr.KeyClear},
		{"+/-", m.tr.KeyTune},
		{"tab", m.tr.KeyFocus},
		{"enter", m.tr.KeyEdit},
		{"?", m.tr.KeyHelp},
		{"q", m.tr.KeyQuit},
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle.Render("<"+k.k+">"), dimStyle.Render(k.d)))
	}
	row := strings.Join(parts, "  ")
	return footerStyle.Width(w).Align(lipgloss.Left).Render(row)
}

// ---- language switcher overlay --------------------------------------------

// viewLangMenu renders the centered language picker. Chinese is listed first,
// then the other five languages. The current language is marked, and the
// cursor highlights the row that <enter> will select.
func (m Model) viewLangMenu() string {
	w := m.width
	h := m.height

	title := langMenuTitleStyle.Render(m.tr.LangMenuTitle)

	// Compute a stable display width for the native-name column so the
	// English names line up regardless of CJK width.
	nativeColW := 0
	for _, info := range i18n.Languages {
		if dw := lipgloss.Width(info.Native); dw > nativeColW {
			nativeColW = dw
		}
	}
	nativeColW += 2 // breathing room

	rows := make([]string, 0, len(i18n.Languages))
	for i, info := range i18n.Languages {
		var b strings.Builder
		// cursor indicator
		if i == m.langCursor {
			b.WriteString(selectedStyle.Render("❯"))
		} else {
			b.WriteString(" ")
		}
		b.WriteString(" ")
		// native name, padded to fixed display width
		native := lipgloss.NewStyle().Width(nativeColW).Render(info.Native)
		b.WriteString(native)
		// english name (dim)
		b.WriteString(dimStyle.Render(info.English))
		// current language marker
		if info.Code == m.lang {
			b.WriteString("   ")
			b.WriteString(lipgloss.NewStyle().Foreground(colorGo).Bold(true).Render("● " + m.tr.LangMenuCurrent))
		}
		rows = append(rows, b.String())
	}
	body := strings.Join(rows, "\n")
	hint := hintStyle.Render(m.tr.LangMenuHint)

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", hint)
	box := langMenuStyle.Render(content)

	// Center within the terminal. Fall back to top-left if the terminal is
	// too small to hold the box.
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// ---- help overlay ---------------------------------------------------------

func (m Model) viewHelp() string {
	body := strings.Join(m.tr.HelpLines, "\n") + "\n\n" + dimStyle.Render(m.tr.HelpReturnHint)
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

// fmtMs renders a millisecond value with the localised unit, e.g. "1200 ms"
// or "1200 毫秒".
func (m Model) fmtMs(ms int64) string {
	return fmt.Sprintf("%d %s", ms, m.tr.MsUnit)
}

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

func (m Model) timeAgo(t time.Time) string {
	if t.IsZero() {
		return dimStyle.Render(m.tr.NeverDash)
	}
	d := time.Since(t).Round(time.Second)
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf(m.tr.AgoFmt, d.String())
}

func (m Model) formatRate(spacingMs int64) string {
	if spacingMs <= 0 {
		return m.tr.Unlimited
	}
	// requests per minute equivalent
	rpm := float64(60_000) / float64(spacingMs)
	if rpm >= 1 {
		return fmt.Sprintf(m.tr.ReqPerMinFmt, rpm)
	}
	rph := float64(3_600_000) / float64(spacingMs)
	return fmt.Sprintf(m.tr.ReqPerHourFmt, rph)
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
