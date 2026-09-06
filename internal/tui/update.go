package tui

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/WaZixwx/RateMate/internal/limiter"
	"github.com/WaZixwx/RateMate/internal/proxy"
)

// Update handles all messages: window resize, ticks, key presses, and proxy
// start/stop results.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Refresh snapshots for the view.
		m.limSnap = m.lim.Snapshot()
		m.recent = m.stats.Recent(40)
		m.bytesIn, m.bytesOut, m.errCount = m.stats.Totals()
		// Status fade.
		if m.status != "" && time.Since(m.statusAt) > 4*time.Second {
			m.status = ""
		}
		return m, tick(500 * time.Millisecond)

	case proxyResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.setStatus("✗ failed to start proxy")
		} else {
			m.err = nil
			if m.proxy.Running() {
				m.setStatus("● proxy started")
			} else {
				m.setStatus("○ proxy stopped")
			}
		}
		return m, nil

	case tea.KeyMsg:
		// While a text input is being edited, route keys there first except
		// for Esc / Enter which commit.
		if m.editing {
			return m.handleEditing(msg)
		}
		return m.handleKey(msg)
	}

	return m, tea.Batch(cmds...)
}

// handleEditing routes keystrokes to the active text input until Enter / Esc.
func (m Model) handleEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.commitEdit()
		m.editing = false
		return m, nil
	case "esc":
		m.editing = false
		// restore the field's value from the live config
		m.refreshInputsFromConfig()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	switch m.focus {
	case focusLimit:
		m.limitInput, cmd = m.limitInput.Update(msg)
	case focusManual:
		m.manualInput, cmd = m.manualInput.Update(msg)
	case focusPort:
		m.portInput, cmd = m.portInput.Update(msg)
	}
	return m, cmd
}

// commitEdit parses the active input and pushes the value into the limiter /
// proxy, then persists.
func (m *Model) commitEdit() {
	switch m.focus {
	case focusLimit:
		if n, err := strconv.Atoi(m.limitInput.Value()); err == nil {
			m.lim.SetLimit(n)
			m.setStatus(fmt.Sprintf("limit → %d", n))
		}
	case focusManual:
		if n, err := strconv.Atoi(m.manualInput.Value()); err == nil {
			m.lim.SetManualDelay(n)
			m.setStatus(fmt.Sprintf("manual delay → %d ms", n))
		}
	case focusPort:
		if m.proxy.Running() {
			m.setStatus("stop proxy to change port")
			return
		}
		p := portValue(m.portInput.Value())
		m.proxy.SetAddr(proxy.ParseAddr(p))
		m.portInput.SetValue(fmt.Sprintf("%d", p))
		m.setStatus(fmt.Sprintf("port → %d", p))
	}
	m.persist()
}

func (m *Model) refreshInputsFromConfig() {
	c := m.lim.Config()
	m.limitInput.SetValue(fmt.Sprintf("%d", c.Limit))
	m.manualInput.SetValue(fmt.Sprintf("%d", c.ManualDelayMs))
	m.portInput.SetValue(fmt.Sprintf("%d", m.cfg.Port))
}

// handleKey handles top-level keys when not editing.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if m.proxy.Running() {
			// graceful: stop proxy then quit
			return m, tea.Sequence(stopProxyCmd(m.proxy), tea.Quit)
		}
		return m, tea.Quit
	case "ctrl+c":
		return m, tea.Quit

	case "s":
		if m.proxy.Running() {
			return m, stopProxyCmd(m.proxy)
		}
		m.err = nil
		return m, startProxyCmd(m.proxy)

	case "p":
		paused := !m.lim.Paused()
		m.lim.SetPaused(paused)
		if paused {
			m.setStatus("⏸ limiter paused (pass-through)")
		} else {
			m.setStatus("▶ limiter resumed")
		}
		return m, nil

	case "m":
		if m.lim.Snapshot().Mode == limiter.ModeAuto {
			m.lim.SetMode(limiter.ModeManual)
			m.setStatus("mode → Manual")
		} else {
			m.lim.SetMode(limiter.ModeAuto)
			m.setStatus("mode → Auto")
		}
		m.persist()
		return m, nil

	case "r":
		m.lim.Reset()
		m.stats.Reset()
		m.setStatus("stats reset")
		return m, nil

	case "c":
		m.stats.Reset()
		m.recent = nil
		m.setStatus("activity log cleared")
		return m, nil

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	// Window cycling (auto mode)
	case "[", "left":
		if m.lim.Snapshot().Mode == limiter.ModeAuto {
			c := m.lim.Config()
			c.Window = (c.Window + 2) % 3 // previous
			m.lim.SetWindow(c.Window)
			m.setStatus(fmt.Sprintf("window → %s", c.Window.String()))
			m.persist()
		}
		return m, nil
	case "]", "right":
		if m.lim.Snapshot().Mode == limiter.ModeAuto {
			c := m.lim.Config()
			c.Window = (c.Window + 1) % 3 // next
			m.lim.SetWindow(c.Window)
			m.setStatus(fmt.Sprintf("window → %s", c.Window.String()))
			m.persist()
		}
		return m, nil

	// Tab cycles focus between editable fields
	case "tab":
		m.focus = nextFocus(m.focus)
		return m, nil
	case "shift+tab":
		m.focus = prevFocus(m.focus)
		return m, nil

	// Enter begins editing the focused field
	case "enter":
		if m.focus == focusNone {
			m.focus = focusLimit
		}
		// only fields with text inputs can be edited
		if m.focus == focusWindow {
			// toggle window inline instead of editing
			c := m.lim.Config()
			c.Window = (c.Window + 1) % 3
			m.lim.SetWindow(c.Window)
			m.setStatus(fmt.Sprintf("window → %s", c.Window.String()))
			m.persist()
			return m, nil
		}
		m.startEdit()
		return m, nil

	// quick numeric tweaks
	case "+", "=":
		m.adjustLimit(+1)
		return m, nil
	case "-":
		m.adjustLimit(-1)
		return m, nil
	case "up":
		m.adjustLimit(+5)
		return m, nil
	case "down":
		m.adjustLimit(-5)
		return m, nil
	}

	return m, nil
}

func (m *Model) adjustLimit(delta int) {
	if m.lim.Snapshot().Mode != limiter.ModeAuto {
		// adjust manual delay instead
		c := m.lim.Config()
		v := c.ManualDelayMs + delta*100
		if v < 0 {
			v = 0
		}
		m.lim.SetManualDelay(v)
		m.manualInput.SetValue(fmt.Sprintf("%d", v))
		m.setStatus(fmt.Sprintf("manual delay → %d ms", v))
	} else {
		c := m.lim.Config()
		v := c.Limit + delta
		if v < 1 {
			v = 1
		}
		m.lim.SetLimit(v)
		m.limitInput.SetValue(fmt.Sprintf("%d", v))
		m.setStatus(fmt.Sprintf("limit → %d", v))
	}
	m.persist()
}

func (m *Model) startEdit() {
	m.editing = true
	switch m.focus {
	case focusLimit:
		m.limitInput.Focus()
	case focusManual:
		m.manualInput.Focus()
	case focusPort:
		m.portInput.Focus()
	}
}

func nextFocus(f focus) focus {
	switch f {
	case focusNone:
		return focusWindow
	case focusWindow:
		return focusLimit
	case focusLimit:
		return focusManual
	case focusManual:
		return focusPort
	default:
		return focusWindow
	}
}

func prevFocus(f focus) focus {
	switch f {
	case focusWindow:
		return focusNone
	case focusLimit:
		return focusWindow
	case focusManual:
		return focusLimit
	case focusPort:
		return focusManual
	default:
		return focusPort
	}
}

var _ = textinput.Model{}
