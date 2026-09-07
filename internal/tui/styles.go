// Package tui contains the bubbletea program that drives RateMate's
// terminal interface.
package tui

import "github.com/charmbracelet/lipgloss"

// ---- Palette ---------------------------------------------------------------
//
// A muted, terminal-friendly palette. We avoid heavy blues/indigos and lean
// on green (running / go), amber (paused / caution), red (stopped / error)
// and a set of grays for chrome. Colors are adaptive so the UI reads well on
// both light and dark terminals.

var (
	colorAccent = lipgloss.AdaptiveColor{Light: "#0f766e", Dark: "#5eead4"} // teal
	colorGo     = lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#86efac"} // green
	colorWarn   = lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#fcd34d"} // amber
	colorStop   = lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#fca5a5"} // red
	colorMuted  = lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#71717a"} // gray
	colorDim    = lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#52525b"} // dimmer gray
	colorFg     = lipgloss.AdaptiveColor{Light: "#1f2937", Dark: "#e4e4e7"} // primary text
	colorHi     = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#fafafa"} // emphasis
	colorPanel  = lipgloss.AdaptiveColor{Light: "#e5e7eb", Dark: "#27272a"} // panel border
	colorBg     = lipgloss.AdaptiveColor{Light: "#fafafa", Dark: "#0c0c0f"} // panel fill
)

// Borders -------------------------------------------------------------------

var rounded = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorPanel)

var panel = rounded.Padding(0, 1)

var innerTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

// Title bar -----------------------------------------------------------------

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorHi).
	Padding(0, 1)

var subtitleStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// Status dot + label --------------------------------------------------------

var statusBase = lipgloss.NewStyle().Bold(true).Padding(0, 1)

// Key / label styles --------------------------------------------------------

var keyStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

var labelStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var valueStyle = lipgloss.NewStyle().
	Foreground(colorFg).
	Bold(true)

var emphasisStyle = lipgloss.NewStyle().
	Foreground(colorHi).
	Bold(true)

var selectedStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

var dimStyle = lipgloss.NewStyle().
	Foreground(colorDim)

var hintStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Italic(true)

// computed-delay callout
var computedStyle = lipgloss.NewStyle().
	Foreground(colorGo).
	Bold(true)

// Footer keybindings --------------------------------------------------------

var footerStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Padding(0, 1)

// Language switcher ---------------------------------------------------------

// langButtonStyle renders the prominent language button in the title bar —
// it sits on a subtly tinted background with a rounded look so it reads as
// "clickable" (press <l>).
var langButtonStyle = lipgloss.NewStyle().
	Foreground(colorHi).
	Background(colorBg).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorAccent).
	Padding(0, 1).
	Bold(true)

// langButtonActiveStyle is the same button when the language menu is open.
var langButtonActiveStyle = langButtonStyle.Copy().
	Foreground(colorAccent).
	BorderForeground(colorGo)

// langMenuStyle is the outer panel of the language switcher overlay.
var langMenuStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorAccent).
	Padding(1, 2).
	Background(colorBg)

// langMenuItemStyle is a single row in the language list.
var langMenuItemStyle = lipgloss.NewStyle().Padding(0, 1)

// langMenuCursorStyle highlights the currently pointed-at language.
var langMenuCursorStyle = lipgloss.NewStyle().
	Foreground(colorHi).
	Background(colorAccent).
	Bold(true).
	Padding(0, 1)

// langMenuCurrentStyle marks the active language row (the one already in use).
var langMenuCurrentStyle = lipgloss.NewStyle().
	Foreground(colorGo).
	Padding(0, 1)

// langMenuTitleStyle is the heading line inside the switcher overlay.
var langMenuTitleStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// helpers -------------------------------------------------------------------

func badge(text string, c lipgloss.AdaptiveColor) string {
	return lipgloss.NewStyle().
		Foreground(c).
		Bold(true).
		Render(text)
}
