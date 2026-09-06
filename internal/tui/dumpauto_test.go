package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDumpAuto(t *testing.T) {
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 96, Height: 34})
	m = mm.(Model)
	// ensure auto mode
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m = out.(Model)
	if m.lim.Snapshot().Mode.String() == "Manual" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
		m = m2.(Model)
	}
	m.limSnap = m.lim.Snapshot()
	fmt.Println("=====BEGIN AUTO VIEW=====")
	fmt.Println(m.View())
	fmt.Println("=====END AUTO VIEW=====")
}
