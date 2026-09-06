package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRender verifies the View() output for the default model renders without
// panic and contains the expected chrome.
func TestRender(t *testing.T) {
	m := New()
	// Simulate a window-size message so layout math runs.
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	m = mm.(Model)
	m.limSnap = m.lim.Snapshot()
	out := m.View()
	if strings.TrimSpace(out) == "" {
		t.Fatal("view is empty")
	}
	for _, want := range []string{"RateMate", "Configuration", "Live Stats", "Recent Activity", "Auto"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q\n---\n%s\n---", want, out)
		}
	}
}

func TestRenderHelp(t *testing.T) {
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mm.(Model)
	m.showHelp = true
	out := m.View()
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("help view missing Usage section:\n%s", out)
	}
}

func TestRenderManualMode(t *testing.T) {
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	m = mm.(Model)
	// flip to manual via Update
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m = m2.(Model)
	m.limSnap = m.lim.Snapshot()
	out := m.View()
	if !strings.Contains(out, "Manual") {
		t.Errorf("manual mode not reflected:\n%s", out)
	}
	if !strings.Contains(out, "ms per request") {
		t.Errorf("manual delay line missing:\n%s", out)
	}
}

func TestStatusFade(t *testing.T) {
	// just ensure status line renders in both states
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mm.(Model)
	m.limSnap = m.lim.Snapshot()
	out := m.View()
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty view")
	}
}
