package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/WaZixwx/RateMate/internal/limiter"
	"github.com/WaZixwx/RateMate/internal/stats"
)

// TestDumpView is a visual smoke test. Run with:
//
//	go test ./internal/tui/ -run TestDumpView -v
//
// and read the t.Log output to eyeball the layout.
func TestDumpView(t *testing.T) {
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 96, Height: 34})
	m = mm.(Model)

	// seed some activity so the panels look alive
	for i := 0; i < 6; i++ {
		m.stats.Push(stats.Entry{
			Time:     time.Now(),
			Method:   "POST",
			Host:     "api.openai.com",
			Path:     "/v1/chat/completions",
			WaitedMs: 1200,
			Status:   200,
			BytesIn:  4 * 1024,
			BytesOut: 28 * 1024,
		})
	}
	m.stats.Push(stats.Entry{
		Time:   time.Now(),
		Method: "POST", Host: "api.anthropic.com", Path: "/v1/messages",
		WaitedMs: 1200, Status: 429, BytesIn: 128, BytesOut: 512, Err: "rate limited",
	})
	m.stats.Push(stats.Entry{
		Time: time.Now(), Method: "GET", Host: "generativelanguage.googleapis.com",
		Path: "/v1beta/models", WaitedMs: 0, Status: 200, BytesIn: 0, BytesOut: 2048,
	})

	m.limSnap = limiter.Snapshot{
		TotalRequests: 1234, DelayedCount: 1200, PassThrough: 34,
		TotalWaitedMs: 1200 * 1180, LastRequestAt: time.Now(),
		SpacingMs: 1200, Mode: limiter.ModeAuto, Window: limiter.WindowMinute,
		Limit: 50, ManualDelayMs: 1000, Paused: false,
	}
	m.recent = m.stats.Recent(8)
	m.bytesIn, m.bytesOut, m.errCount = m.stats.Totals()

	out := m.View()
	// Print with a visible border so trailing whitespace is obvious.
	fmt.Println("=====BEGIN VIEW=====")
	fmt.Println(out)
	fmt.Println("=====END VIEW=====")
}

func TestDumpHelp(t *testing.T) {
	m := New()
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 40})
	m = mm.(Model)
	m.showHelp = true
	fmt.Println("=====BEGIN HELP=====")
	fmt.Println(m.View())
	fmt.Println("=====END HELP=====")
}
