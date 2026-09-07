package tui

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/WaZixwx/RateMate/internal/config"
	"github.com/WaZixwx/RateMate/internal/i18n"
	"github.com/WaZixwx/RateMate/internal/limiter"
	"github.com/WaZixwx/RateMate/internal/proxy"
	"github.com/WaZixwx/RateMate/internal/stats"
)

// focus identifies which editable element currently has focus.
type focus int

const (
	focusNone focus = iota
	focusWindow
	focusLimit
	focusManual
	focusPort
)

const activityCapacity = 200

// Model is the bubbletea model for RateMate.
type Model struct {
	// configuration (persisted)
	cfg config.File

	// i18n
	lang i18n.Lang
	tr   i18n.Strings

	// runtime components
	lim   *limiter.Limiter
	proxy *proxy.Server
	stats *stats.Tracker
	log   *log.Logger

	// terminal
	width  int
	height int

	// focus + interaction
	focus    focus
	editing  bool
	showHelp bool

	// language switcher menu
	showLang   bool
	langCursor int

	// inputs
	limitInput  textinput.Model
	manualInput textinput.Model
	portInput   textinput.Model

	// transient status line
	status   string
	err      error
	statusAt time.Time

	// last fetched snapshots (refreshed on tick)
	limSnap  limiter.Snapshot
	recent   []stats.Entry
	bytesIn  int64
	bytesOut int64
	errCount int64

	// activity log filter (for future expansion; currently unused)
	filterHost string
}

// New constructs the initial model. It loads the persisted config and wires up
// the limiter, proxy and stats tracker but does NOT start the proxy — the
// user presses [s] (or the config autostart flag handles it).
func New() Model {
	cfg := config.Load()
	lim := limiter.New(cfg.ToLimiterConfig())
	tr := stats.New(activityCapacity)
	logger := log.New(os.Stderr, "[proxy] ", log.LstdFlags|log.Lmsgprefix)

	m := Model{
		cfg:   cfg,
		lim:   lim,
		stats: tr,
		log:   logger,
	}

	// i18n: resolve the persisted language (default English) and load the
	// matching string set. The cursor starts on the current language so the
	// switcher opens highlighting the active entry.
	m.lang = i18n.Normalise(i18n.Lang(cfg.Language))
	m.tr = i18n.Get(m.lang)
	m.langCursor = i18n.IndexOf(m.lang)

	m.limitInput = newNumInput(fmt.Sprintf("%d", cfg.Limit))
	m.manualInput = newNumInput(fmt.Sprintf("%d", cfg.ManualDelayMs))
	m.portInput = newNumInput(fmt.Sprintf("%d", cfg.Port))

	m.proxy = proxy.New(proxy.ParseAddr(cfg.Port), lim, tr, logger)

	if cfg.AutoStart {
		if err := m.proxy.Start(); err != nil {
			m.err = err
		}
	}
	return m
}

func newNumInput(v string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "0"
	ti.CharLimit = 8
	ti.SetValue(v)
	return ti
}

// Init kicks off the first tick.
func (m Model) Init() tea.Cmd {
	return tick(500 * time.Millisecond)
}

// tickMsg drives periodic refreshes of stats + activity log.
type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// startProxyCmd / stopProxyCmd wrap the side-effectful calls so they run off
// the main update loop.
type proxyResultMsg struct{ err error }

func startProxyCmd(p *proxy.Server) tea.Cmd {
	return func() tea.Msg {
		return proxyResultMsg{err: p.Start()}
	}
}

func stopProxyCmd(p *proxy.Server) tea.Cmd {
	return func() tea.Msg {
		p.Stop()
		return proxyResultMsg{}
	}
}

// setStatus stores a status string with a timestamp so it can fade after a
// while.
func (m *Model) setStatus(s string) {
	m.status = s
	m.statusAt = time.Now()
}

// persist writes the current config to disk.
func (m *Model) persist() {
	s := m.lim.Config()
	m.cfg.FromLimiterConfig(s)
	m.cfg.Port = portValue(m.portInput.Value())
	m.cfg.Language = string(m.lang)
	if err := config.Save(m.cfg); err != nil {
		m.err = err
	}
}

// setLang switches the active language, refreshes the translator, persists
// the choice and shows a transient confirmation. The language menu cursor is
// moved to the newly selected language.
func (m *Model) setLang(l i18n.Lang) {
	m.lang = i18n.Normalise(l)
	m.tr = i18n.Get(m.lang)
	m.langCursor = i18n.IndexOf(m.lang)
	m.cfg.Language = string(m.lang)
	if err := config.Save(m.cfg); err != nil {
		m.err = err
	}
	info := i18n.Info(m.lang)
	m.setStatus(fmt.Sprintf(m.tr.LangChangedFmt, info.Native))
}

func portValue(s string) int {
	var p int
	_, _ = fmt.Sscanf(s, "%d", &p)
	if p <= 0 || p > 65535 {
		return 8080
	}
	return p
}

// unused but kept to silence ctx import when not needed elsewhere
var _ = context.Background
