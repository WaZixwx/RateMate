// Package i18n provides the internationalisation layer for RateMate's TUI.
//
// RateMate ships with six languages. The default is English; Chinese (中文)
// is intentionally listed first in the language switcher, followed by the
// other four. Each language is defined as a Strings struct; non-English
// languages inherit every field from English and override only the strings
// that differ, so adding a new phrase in English automatically falls back for
// every locale until it is translated.
//
// The package is pure data plus a tiny lookup helper; the TUI owns a Strings
// value and swaps it whenever the user picks a new language.
package i18n

// Lang is a language code (e.g. "en", "zh").
type Lang string

const (
	// ZH — 中文 (简体). Listed FIRST in the switcher.
	ZH Lang = "zh"
	// EN — English. The DEFAULT language.
	EN Lang = "en"
	// JA — 日本語.
	JA Lang = "ja"
	// ES — Español.
	ES Lang = "es"
	// FR — Français.
	FR Lang = "fr"
	// DE — Deutsch.
	DE Lang = "de"
)

// Default is the language used when none is persisted or when the saved
// value is unknown.
const Default Lang = EN

// LangInfo describes a selectable language for the switcher menu.
type LangInfo struct {
	Code    Lang
	Native  string // endonym shown in the menu, e.g. "中文"
	English string // English name, e.g. "Chinese"
	Flag    string // a small glyph used in the title-bar button
}

// Languages is the ORDERED list shown in the language switcher. Chinese is
// deliberately first, then English (the default) and the other four.
var Languages = []LangInfo{
	{Code: ZH, Native: "中文", English: "Chinese", Flag: "🇨🇳"},
	{Code: EN, Native: "English", English: "English", Flag: "🇬🇧"},
	{Code: JA, Native: "日本語", English: "Japanese", Flag: "🇯🇵"},
	{Code: ES, Native: "Español", English: "Spanish", Flag: "🇪🇸"},
	{Code: FR, Native: "Français", English: "French", Flag: "🇫🇷"},
	{Code: DE, Native: "Deutsch", English: "German", Flag: "🇩🇪"},
}

// IndexOf returns the position of a language in the Languages list, or the
// index of the default language when the code is unknown.
func IndexOf(l Lang) int {
	for i, info := range Languages {
		if info.Code == l {
			return i
		}
	}
	return IndexOf(Default)
}

// Info returns the LangInfo for a code, falling back to the default.
func Info(l Lang) LangInfo {
	i := IndexOf(l)
	return Languages[i]
}

// Normalise coerces an arbitrary code into a known language, defaulting to
// English when it is empty or unrecognised.
func Normalise(l Lang) Lang {
	for _, info := range Languages {
		if info.Code == l {
			return l
		}
	}
	return Default
}

// Strings holds every user-facing string the TUI renders. Format fields use
// fmt-style verbs in a FIXED argument order across languages so the call
// sites can pass the same arguments regardless of locale.
type Strings struct {
	// ---- title bar ----
	AppSubtitle string

	// ---- status badges ----
	StatusStopped string
	StatusPaused  string
	StatusRunning string

	// ---- mode row ----
	ModeAuto       string
	ModeManual     string
	ModeSwitchHint string

	// ---- configuration panel ----
	ConfigTitle            string
	WindowLabel            string
	LimitLabel             string
	DelayLabel             string
	RequestsPerShortFmt    string // "requests per %s"  (s = short window)
	ComputedDelayFmt       string // "→ computed delay: %s between requests"  (s = "1200 ms")
	EffectiveThroughputFmt string // "≈ %s effective throughput"
	MsPerRequest           string
	EffectiveRateFmt       string // "→ effective rate: %s"
	ReqPerSecMaxFmt        string // "%d req/s max"
	NudgeHint              string

	// ---- live stats panel ----
	StatsTitle  string
	TotalReqs   string
	Delayed     string
	PassThrough string
	AvgWait     string
	InOut       string
	Errors      string
	LastReq     string

	// ---- activity panel ----
	ActivityTitle string
	NoRequests    string

	// ---- idle status line ----
	AutoStatusFmt   string // "auto: %d req / %s → %d ms spacing"
	ManualStatusFmt string // "manual: %d ms spacing"

	// ---- time / unit formatting ----
	AgoFmt        string // "%s ago"
	NeverDash     string // "—"
	Unlimited     string
	ReqPerMinFmt  string // "%.0f req/min"
	ReqPerHourFmt string // "%.0f req/h"
	MsUnit        string // "ms"

	// ---- footer key descriptions ----
	KeyStartStop string
	KeyMode      string
	KeyPause     string
	KeyReset     string
	KeyClear     string
	KeyTune      string
	KeyFocus     string
	KeyEdit      string
	KeyLang      string
	KeyHelp      string
	KeyQuit      string

	// ---- transient status messages ----
	FailedStart           string
	ProxyStarted          string
	ProxyStopped          string
	PausedMsg             string
	ResumedMsg            string
	ModeManualSet         string
	ModeAutoSet           string
	StatsReset            string
	ActivityCleared       string
	WindowSetFmt          string // "window → %s"
	LimitSetFmt           string // "limit → %d"
	ManualDelaySetFmt     string // "manual delay → %d ms"
	PortSetFmt            string // "port → %d"
	StopProxyToChangePort string
	LangChangedFmt        string // "language → %s"

	// ---- window names (auto mode) ----
	WindowSecond      string
	WindowMinute      string
	WindowHour        string
	WindowShortSecond string // "s"
	WindowShortMinute string // "m"
	WindowShortHour   string // "h"

	// ---- language switcher menu ----
	LangMenuTitle    string
	LangMenuHint     string
	LangMenuCurrent  string
	LangButtonFmt    string // "🌐 %s"  (s = upper code)
	LangMenuCloseKey string

	// ---- help overlay ----
	HelpLines      []string
	HelpReturnHint string
}

// english returns the full English string set, which is the base for every
// other language.
func english() Strings {
	return Strings{
		AppSubtitle: "LLM API delay proxy",

		StatusStopped: "● Stopped",
		StatusPaused:  "⏸ Paused",
		StatusRunning: "● Running",

		ModeAuto:       "Auto",
		ModeManual:     "Manual",
		ModeSwitchHint: "<m> switch",

		ConfigTitle:            "⚙ Configuration",
		WindowLabel:            "Window:",
		LimitLabel:             "Limit:",
		DelayLabel:             "Delay:",
		RequestsPerShortFmt:    "requests per %s",
		ComputedDelayFmt:       "→ computed delay: %s between requests",
		EffectiveThroughputFmt: "≈ %s effective throughput",
		MsPerRequest:           "ms per request",
		EffectiveRateFmt:       "→ effective rate: %s",
		ReqPerSecMaxFmt:        "%d req/s max",
		NudgeHint:              "[+]/[-] or ↑/↓ nudges by 100ms",

		StatsTitle:  "📊 Live Stats",
		TotalReqs:   "Total reqs",
		Delayed:     "Delayed",
		PassThrough: "Pass-through",
		AvgWait:     "Avg wait",
		InOut:       "In / Out",
		Errors:      "Errors",
		LastReq:     "Last req",

		ActivityTitle: "📡 Recent Activity",
		NoRequests:    "no requests yet",

		AutoStatusFmt:   "auto: %d req / %s → %d ms spacing",
		ManualStatusFmt: "manual: %d ms spacing",

		AgoFmt:        "%s ago",
		NeverDash:     "—",
		Unlimited:     "unlimited",
		ReqPerMinFmt:  "%.0f req/min",
		ReqPerHourFmt: "%.0f req/h",
		MsUnit:        "ms",

		KeyStartStop: "start/stop",
		KeyMode:      "mode",
		KeyPause:     "pause",
		KeyReset:     "reset",
		KeyClear:     "clear",
		KeyTune:      "tune",
		KeyFocus:     "focus",
		KeyEdit:      "edit",
		KeyLang:      "lang",
		KeyHelp:      "help",
		KeyQuit:      "quit",

		FailedStart:           "✗ failed to start proxy",
		ProxyStarted:          "● proxy started",
		ProxyStopped:          "○ proxy stopped",
		PausedMsg:             "⏸ limiter paused (pass-through)",
		ResumedMsg:            "▶ limiter resumed",
		ModeManualSet:         "mode → Manual",
		ModeAutoSet:           "mode → Auto",
		StatsReset:            "stats reset",
		ActivityCleared:       "activity log cleared",
		WindowSetFmt:          "window → %s",
		LimitSetFmt:           "limit → %d",
		ManualDelaySetFmt:     "manual delay → %d ms",
		PortSetFmt:            "port → %d",
		StopProxyToChangePort: "stop proxy to change port",
		LangChangedFmt:        "language → %s",

		WindowSecond:      "Second",
		WindowMinute:      "Minute",
		WindowHour:        "Hour",
		WindowShortSecond: "s",
		WindowShortMinute: "m",
		WindowShortHour:   "h",

		LangMenuTitle:    "🌐  Language",
		LangMenuHint:     "↑/↓ navigate   enter select   esc cancel",
		LangMenuCurrent:  "current",
		LangButtonFmt:    "🌐 %s",
		LangMenuCloseKey: "l",

		HelpReturnHint: "press ? or esc to return",
		HelpLines: []string{
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
			"  s              start / stop the proxy",
			"  m              toggle Auto ↔ Manual mode",
			"  p              pause / resume the limiter",
			"  r              reset counters & log",
			"  c              clear activity log only",
			"  ◀/▶ or </>     cycle the time window (Auto)",
			"  +/-            tune limit (Auto) or delay (Manual)",
			"  ↑/↓            tune by ±5 (Auto) or ±100ms (Manual)",
			"  tab/shift+tab  cycle focus",
			"  enter          edit focused field",
			"  esc            cancel edit / close help",
			"  l              switch language",
			"  ?              toggle this help",
			"  q              quit",
			"",
			"Config is persisted to ~/.ratemate/config.json",
		},
	}
}

// chinese (中文 / 简体) — listed first in the switcher.
func chinese() Strings {
	s := english()
	s.AppSubtitle = "LLM API 延时代理"

	s.StatusStopped = "● 已停止"
	s.StatusPaused = "⏸ 已暂停"
	s.StatusRunning = "● 运行中"

	s.ModeAuto = "自动"
	s.ModeManual = "手动"
	s.ModeSwitchHint = "<m> 切换"

	s.ConfigTitle = "⚙ 配置"
	s.WindowLabel = "窗口："
	s.LimitLabel = "限额："
	s.DelayLabel = "延迟："
	s.RequestsPerShortFmt = "每%s 请求数"
	s.ComputedDelayFmt = "→ 计算延迟：请求间隔 %s"
	s.EffectiveThroughputFmt = "≈ %s 有效吞吐"
	s.MsPerRequest = "毫秒 / 每请求"
	s.EffectiveRateFmt = "→ 有效速率：%s"
	s.ReqPerSecMaxFmt = "最高 %d 请求/秒"
	s.NudgeHint = "[+]/[-] 或 ↑/↓ 每次调整 100ms"

	s.StatsTitle = "📊 实时统计"
	s.TotalReqs = "总请求数"
	s.Delayed = "已延迟"
	s.PassThrough = "直通"
	s.AvgWait = "平均等待"
	s.InOut = "入 / 出"
	s.Errors = "错误数"
	s.LastReq = "上次请求"

	s.ActivityTitle = "📡 最近活动"
	s.NoRequests = "暂无请求"

	s.AutoStatusFmt = "自动：%d 请求 / %s → %d 毫秒间隔"
	s.ManualStatusFmt = "手动：%d 毫秒间隔"

	s.AgoFmt = "%s 前"
	s.NeverDash = "—"
	s.Unlimited = "无限制"
	s.ReqPerMinFmt = "%.0f 请求/分"
	s.ReqPerHourFmt = "%.0f 请求/时"
	s.MsUnit = "毫秒"

	s.KeyStartStop = "启动/停止"
	s.KeyMode = "模式"
	s.KeyPause = "暂停"
	s.KeyReset = "重置"
	s.KeyClear = "清除"
	s.KeyTune = "调节"
	s.KeyFocus = "焦点"
	s.KeyEdit = "编辑"
	s.KeyLang = "语言"
	s.KeyHelp = "帮助"
	s.KeyQuit = "退出"

	s.FailedStart = "✗ 启动代理失败"
	s.ProxyStarted = "● 代理已启动"
	s.ProxyStopped = "○ 代理已停止"
	s.PausedMsg = "⏸ 限流器已暂停（直通）"
	s.ResumedMsg = "▶ 限流器已恢复"
	s.ModeManualSet = "模式 → 手动"
	s.ModeAutoSet = "模式 → 自动"
	s.StatsReset = "统计已重置"
	s.ActivityCleared = "活动日志已清空"
	s.WindowSetFmt = "窗口 → %s"
	s.LimitSetFmt = "限额 → %d"
	s.ManualDelaySetFmt = "手动延迟 → %d 毫秒"
	s.PortSetFmt = "端口 → %d"
	s.StopProxyToChangePort = "请先停止代理再修改端口"
	s.LangChangedFmt = "语言 → %s"

	s.WindowSecond = "秒"
	s.WindowMinute = "分"
	s.WindowHour = "时"
	s.WindowShortSecond = "秒"
	s.WindowShortMinute = "分"
	s.WindowShortHour = "时"

	s.LangMenuTitle = "🌐  语言 / Language"
	s.LangMenuHint = "↑/↓ 选择   enter 确认   esc 取消"
	s.LangMenuCurrent = "当前"
	s.LangButtonFmt = "🌐 %s"
	s.LangMenuCloseKey = "l"

	s.HelpReturnHint = "按 ? 或 esc 返回"
	s.HelpLines = []string{
		"⚡ RateMate — LLM API 延时代理",
		"",
		"RateMate 运行一个本地 HTTP/HTTPS 代理，在每个转发请求前注入",
		"一段计算好的延迟，确保你永远不会超过 LLM API 的",
		"每秒 / 每分 / 每时速率限制。",
		"",
		"用法：",
		"  1. 按 <s> 启动代理（默认 :8080）。",
		"  2. 将你的客户端指向代理，例如：",
		"       export HTTPS_PROXY=http://127.0.0.1:8080",
		"       export HTTP_PROXY=http://127.0.0.1:8080",
		"  3. 用 <m>（模式）和输入框调节延迟。Tab 切换焦点，",
		"     Enter 编辑，Esc 取消，+/- 微调。",
		"  4. 按 <p> 暂停（直通）而无需停止服务器。",
		"",
		"按键：",
		"  s              启动 / 停止代理",
		"  m              切换 自动 ↔ 手动 模式",
		"  p              暂停 / 恢复限流器",
		"  r              重置计数器和日志",
		"  c              仅清空活动日志",
		"  ◀/▶ 或 </>     切换时间窗口（自动）",
		"  +/-            调节限额（自动）或延迟（手动）",
		"  ↑/↓            微调 ±5（自动）或 ±100ms（手动）",
		"  tab/shift+tab  切换焦点",
		"  enter          编辑当前字段",
		"  esc            取消编辑 / 关闭帮助",
		"  l              切换语言",
		"  ?              切换此帮助",
		"  q              退出",
		"",
		"配置保存在 ~/.ratemate/config.json",
	}
	return s
}

// japanese (日本語).
func japanese() Strings {
	s := english()
	s.AppSubtitle = "LLM API 遅延プロキシ"

	s.StatusStopped = "● 停止中"
	s.StatusPaused = "⏸ 一時停止"
	s.StatusRunning = "● 実行中"

	s.ModeAuto = "自動"
	s.ModeManual = "手動"
	s.ModeSwitchHint = "<m> 切替"

	s.ConfigTitle = "⚙ 設定"
	s.WindowLabel = "ウィンドウ:"
	s.LimitLabel = "上限:"
	s.DelayLabel = "遅延:"
	s.RequestsPerShortFmt = "%s あたりの要求数"
	s.ComputedDelayFmt = "→ 計算遅延: 要求間 %s"
	s.EffectiveThroughputFmt = "≈ %s の実効スループット"
	s.MsPerRequest = "ms / 要求"
	s.EffectiveRateFmt = "→ 実効レート: %s"
	s.ReqPerSecMaxFmt = "最大 %d req/s"
	s.NudgeHint = "[+]/[-] または ↑/↓ で 100ms 単位"

	s.StatsTitle = "📊 ライブ統計"
	s.TotalReqs = "合計要求"
	s.Delayed = "遅延あり"
	s.PassThrough = "パススルー"
	s.AvgWait = "平均待機"
	s.InOut = "受 / 送"
	s.Errors = "エラー"
	s.LastReq = "最終要求"

	s.ActivityTitle = "📡 最近のアクティビティ"
	s.NoRequests = "まだ要求がありません"

	s.AutoStatusFmt = "自動: %d req / %s → %d ms 間隔"
	s.ManualStatusFmt = "手動: %d ms 間隔"

	s.AgoFmt = "%s 前"
	s.NeverDash = "—"
	s.Unlimited = "無制限"
	s.ReqPerMinFmt = "%.0f req/分"
	s.ReqPerHourFmt = "%.0f req/時"
	s.MsUnit = "ms"

	s.KeyStartStop = "開始/停止"
	s.KeyMode = "モード"
	s.KeyPause = "一時停止"
	s.KeyReset = "リセット"
	s.KeyClear = "クリア"
	s.KeyTune = "調整"
	s.KeyFocus = "フォーカス"
	s.KeyEdit = "編集"
	s.KeyLang = "言語"
	s.KeyHelp = "ヘルプ"
	s.KeyQuit = "終了"

	s.FailedStart = "✗ プロキシ起動失敗"
	s.ProxyStarted = "● プロキシ開始"
	s.ProxyStopped = "○ プロキシ停止"
	s.PausedMsg = "⏸ リミッターを一時停止（パススルー）"
	s.ResumedMsg = "▶ リミッターを再開"
	s.ModeManualSet = "モード → 手動"
	s.ModeAutoSet = "モード → 自動"
	s.StatsReset = "統計をリセット"
	s.ActivityCleared = "アクティビティログをクリア"
	s.WindowSetFmt = "ウィンドウ → %s"
	s.LimitSetFmt = "上限 → %d"
	s.ManualDelaySetFmt = "手動遅延 → %d ms"
	s.PortSetFmt = "ポート → %d"
	s.StopProxyToChangePort = "ポート変更にはプロキシを停止してください"
	s.LangChangedFmt = "言語 → %s"

	s.WindowSecond = "秒"
	s.WindowMinute = "分"
	s.WindowHour = "時"
	s.WindowShortSecond = "秒"
	s.WindowShortMinute = "分"
	s.WindowShortHour = "時"

	s.LangMenuTitle = "🌐  言語 / Language"
	s.LangMenuHint = "↑/↓ 移動   enter 選択   esc キャンセル"
	s.LangMenuCurrent = "現在"
	s.LangButtonFmt = "🌐 %s"
	s.LangMenuCloseKey = "l"

	s.HelpReturnHint = "? または esc で戻る"
	s.HelpLines = []string{
		"⚡ RateMate — LLM API 遅延プロキシ",
		"",
		"RateMate はローカル HTTP/HTTPS プロキシを実行し、転送する",
		"各リクエストの前に小さな遅延を注入して、LLM API の",
		"秒/分/時あたりのレート制限を超えないようにします。",
		"",
		"使い方：",
		"  1. <s> でプロキシを開始（デフォルト :8080）。",
		"  2. クライアントをプロキシに向けます。例：",
		"       export HTTPS_PROXY=http://127.0.0.1:8080",
		"       export HTTP_PROXY=http://127.0.0.1:8080",
		"  3. <m>（モード）と入力と入力欄で遅延を調整。Tab でフォーカス移動、",
		"     Enter で編集、Esc でキャンセル、+/- で微調整。",
		"  4. <p> で一時停止（パススルー）。サーバーは停止しません。",
		"",
		"キー：",
		"  s              プロキシを開始 / 停止",
		"  m               自動 ↔ 手動 モード切替",
		"  p               リミッターの一時停止 / 再開",
		"  r               カウンターとログをリセット",
		"  c               アクティビティログのみクリア",
		"  ◀/▶ または </>  時間ウィンドウ切替（自動）",
		"  +/-              上限（自動）または遅延（手動）を調整",
		"  ↑/↓              ±5（自動）または ±100ms（手動）で微調整",
		"  tab/shift+tab   フォーカス切替",
		"  enter           フォーカス中のフィールドを編集",
		"  esc             編集キャンセル / ヘルプを閉じる",
		"  l               言語切替",
		"  ?               このヘルプの表示切替",
		"  q               終了",
		"",
		"設定は ~/.ratemate/config.json に保存されます",
	}
	return s
}

// spanish (Español).
func spanish() Strings {
	s := english()
	s.AppSubtitle = "proxy de retardo para APIs LLM"

	s.StatusStopped = "● Detenido"
	s.StatusPaused = "⏸ En pausa"
	s.StatusRunning = "● En ejecución"

	s.ModeAuto = "Auto"
	s.ModeManual = "Manual"
	s.ModeSwitchHint = "<m> cambiar"

	s.ConfigTitle = "⚙ Configuración"
	s.WindowLabel = "Ventana:"
	s.LimitLabel = "Límite:"
	s.DelayLabel = "Retardo:"
	s.RequestsPerShortFmt = "peticiones por %s"
	s.ComputedDelayFmt = "→ retardo calculado: %s entre peticiones"
	s.EffectiveThroughputFmt = "≈ %s rendimiento efectivo"
	s.MsPerRequest = "ms por petición"
	s.EffectiveRateFmt = "→ tasa efectiva: %s"
	s.ReqPerSecMaxFmt = "%d req/s máx"
	s.NudgeHint = "[+]/[-] o ↑/↓ ajusta 100ms"

	s.StatsTitle = "📊 Estadísticas en vivo"
	s.TotalReqs = "Peticiones totales"
	s.Delayed = "Retardadas"
	s.PassThrough = "Directas"
	s.AvgWait = "Espera media"
	s.InOut = "Ent / Sal"
	s.Errors = "Errores"
	s.LastReq = "Última petición"

	s.ActivityTitle = "📡 Actividad reciente"
	s.NoRequests = "todavía no hay peticiones"

	s.AutoStatusFmt = "auto: %d req / %s → %d ms de separación"
	s.ManualStatusFmt = "manual: %d ms de separación"

	s.AgoFmt = "hace %s"
	s.NeverDash = "—"
	s.Unlimited = "ilimitado"
	s.ReqPerMinFmt = "%.0f req/min"
	s.ReqPerHourFmt = "%.0f req/h"
	s.MsUnit = "ms"

	s.KeyStartStop = "iniciar/parar"
	s.KeyMode = "modo"
	s.KeyPause = "pausa"
	s.KeyReset = "reiniciar"
	s.KeyClear = "limpiar"
	s.KeyTune = "ajustar"
	s.KeyFocus = "foco"
	s.KeyEdit = "editar"
	s.KeyLang = "idioma"
	s.KeyHelp = "ayuda"
	s.KeyQuit = "salir"

	s.FailedStart = "✗ no se pudo iniciar el proxy"
	s.ProxyStarted = "● proxy iniciado"
	s.ProxyStopped = "○ proxy detenido"
	s.PausedMsg = "⏸ limitador en pausa (paso directo)"
	s.ResumedMsg = "▶ limitador reanudado"
	s.ModeManualSet = "modo → Manual"
	s.ModeAutoSet = "modo → Auto"
	s.StatsReset = "estadísticas reiniciadas"
	s.ActivityCleared = "registro de actividad limpiado"
	s.WindowSetFmt = "ventana → %s"
	s.LimitSetFmt = "límite → %d"
	s.ManualDelaySetFmt = "retardo manual → %d ms"
	s.PortSetFmt = "puerto → %d"
	s.StopProxyToChangePort = "detén el proxy para cambiar el puerto"
	s.LangChangedFmt = "idioma → %s"

	s.WindowSecond = "Segundo"
	s.WindowMinute = "Minuto"
	s.WindowHour = "Hora"
	s.WindowShortSecond = "s"
	s.WindowShortMinute = "m"
	s.WindowShortHour = "h"

	s.LangMenuTitle = "🌐  Idioma / Language"
	s.LangMenuHint = "↑/↓ navegar   enter elegir   esc cancelar"
	s.LangMenuCurrent = "actual"
	s.LangButtonFmt = "🌐 %s"
	s.LangMenuCloseKey = "l"

	s.HelpReturnHint = "pulsa ? o esc para volver"
	s.HelpLines = []string{
		"⚡ RateMate — proxy de retardo para APIs de LLM",
		"",
		"RateMate ejecuta un proxy HTTP/HTTPS local que inyecta un pequeño",
		"retardo antes de cada petición reenviada para no superar el límite",
		"de peticiones por segundo / minuto / hora de tu API de LLM.",
		"",
		"Uso:",
		"  1. Pulsa <s> para iniciar el proxy (por defecto :8080).",
		"  2. Apunta tu cliente al proxy, p. ej.:",
		"       export HTTPS_PROXY=http://127.0.0.1:8080",
		"       export HTTP_PROXY=http://127.0.0.1:8080",
		"  3. Ajusta el retardo con <m> (modo) y los campos. Tab mueve el foco,",
		"     Enter edita, Esc cancela, +/- ajusta.",
		"  4. Pulsa <p> para pausar (paso directo) sin parar el servidor.",
		"",
		"Teclas:",
		"  s              iniciar / parar el proxy",
		"  m               conmutar modo Auto ↔ Manual",
		"  p               pausar / reanudar el limitador",
		"  r               reiniciar contadores y registro",
		"  c               limpiar solo el registro de actividad",
		"  ◀/▶ o </>      cambiar la ventana de tiempo (Auto)",
		"  +/-              ajustar límite (Auto) o retardo (Manual)",
		"  ↑/↓              ajustar ±5 (Auto) o ±100ms (Manual)",
		"  tab/shift+tab   cambiar el foco",
		"  enter           editar el campo con foco",
		"  esc             cancelar edición / cerrar ayuda",
		"  l               cambiar idioma",
		"  ?               mostrar/ocultar esta ayuda",
		"  q               salir",
		"",
		"La configuración se guarda en ~/.ratemate/config.json",
	}
	return s
}

// french (Français).
func french() Strings {
	s := english()
	s.AppSubtitle = "proxy de délai pour APIs LLM"

	s.StatusStopped = "● Arrêté"
	s.StatusPaused = "⏸ En pause"
	s.StatusRunning = "● En cours"

	s.ModeAuto = "Auto"
	s.ModeManual = "Manuel"
	s.ModeSwitchHint = "<m> changer"

	s.ConfigTitle = "⚙ Configuration"
	s.WindowLabel = "Fenêtre :"
	s.LimitLabel = "Limite :"
	s.DelayLabel = "Délai :"
	s.RequestsPerShortFmt = "requêtes par %s"
	s.ComputedDelayFmt = "→ délai calculé : %s entre requêtes"
	s.EffectiveThroughputFmt = "≈ %s débit effectif"
	s.MsPerRequest = "ms par requête"
	s.EffectiveRateFmt = "→ taux effectif : %s"
	s.ReqPerSecMaxFmt = "%d req/s max"
	s.NudgeHint = "[+]/[-] ou ↑/↓ ajuste par 100ms"

	s.StatsTitle = "📊 Stats en direct"
	s.TotalReqs = "Requêtes totales"
	s.Delayed = "Retardées"
	s.PassThrough = "Directes"
	s.AvgWait = "Attente moy."
	s.InOut = "Entr / Sortie"
	s.Errors = "Erreurs"
	s.LastReq = "Dernière req."

	s.ActivityTitle = "📡 Activité récente"
	s.NoRequests = "aucune requête pour le moment"

	s.AutoStatusFmt = "auto : %d req / %s → %d ms d'espacement"
	s.ManualStatusFmt = "manuel : %d ms d'espacement"

	s.AgoFmt = "il y a %s"
	s.NeverDash = "—"
	s.Unlimited = "illimité"
	s.ReqPerMinFmt = "%.0f req/min"
	s.ReqPerHourFmt = "%.0f req/h"
	s.MsUnit = "ms"

	s.KeyStartStop = "démarrer/arrêter"
	s.KeyMode = "mode"
	s.KeyPause = "pause"
	s.KeyReset = "réinit."
	s.KeyClear = "effacer"
	s.KeyTune = "ajuster"
	s.KeyFocus = "focus"
	s.KeyEdit = "éditer"
	s.KeyLang = "langue"
	s.KeyHelp = "aide"
	s.KeyQuit = "quitter"

	s.FailedStart = "✗ échec du démarrage du proxy"
	s.ProxyStarted = "● proxy démarré"
	s.ProxyStopped = "○ proxy arrêté"
	s.PausedMsg = "⏸ limiteur en pause (passant)"
	s.ResumedMsg = "▶ limiteur repris"
	s.ModeManualSet = "mode → Manuel"
	s.ModeAutoSet = "mode → Auto"
	s.StatsReset = "stats réinitialisées"
	s.ActivityCleared = "journal d'activité effacé"
	s.WindowSetFmt = "fenêtre → %s"
	s.LimitSetFmt = "limite → %d"
	s.ManualDelaySetFmt = "délai manuel → %d ms"
	s.PortSetFmt = "port → %d"
	s.StopProxyToChangePort = "arrêtez le proxy pour changer le port"
	s.LangChangedFmt = "langue → %s"

	s.WindowSecond = "Seconde"
	s.WindowMinute = "Minute"
	s.WindowHour = "Heure"
	s.WindowShortSecond = "s"
	s.WindowShortMinute = "m"
	s.WindowShortHour = "h"

	s.LangMenuTitle = "🌐  Langue / Language"
	s.LangMenuHint = "↑/↓ naviguer   enter choisir   esc annuler"
	s.LangMenuCurrent = "actuelle"
	s.LangButtonFmt = "🌐 %s"
	s.LangMenuCloseKey = "l"

	s.HelpReturnHint = "appuie sur ? ou esc pour revenir"
	s.HelpLines = []string{
		"⚡ RateMate — proxy de délai pour APIs LLM",
		"",
		"RateMate lance un proxy HTTP/HTTPS local qui injecte un petit",
		"délai avant chaque requête transmise afin de ne jamais dépasser",
		"la limite de requêtes par seconde / minute / heure de ton API LLM.",
		"",
		"Utilisation :",
		"  1. Appuie sur <s> pour démarrer le proxy (défaut :8080).",
		"  2. Pointe ton client vers le proxy, ex. :",
		"       export HTTPS_PROXY=http://127.0.0.1:8080",
		"       export HTTP_PROXY=http://127.0.0.1:8080",
		"  3. Règle le délai avec <m> (mode) et les champs. Tab déplace le focus,",
		"     Enter édite, Esc annule, +/- ajuste.",
		"  4. Appuie sur <p> pour mettre en pause (passant) sans arrêter le serveur.",
		"",
		"Touches :",
		"  s              démarrer / arrêter le proxy",
		"  m               basculer mode Auto ↔ Manuel",
		"  p               pause / reprise du limiteur",
		"  r               réinitialiser compteurs et journal",
		"  c               effacer uniquement le journal d'activité",
		"  ◀/▶ ou </>      changer la fenêtre temporelle (Auto)",
		"  +/-              ajuster la limite (Auto) ou le délai (Manuel)",
		"  ↑/↓              ajuster de ±5 (Auto) ou ±100ms (Manuel)",
		"  tab/shift+tab   changer le focus",
		"  enter           éditer le champ focus",
		"  esc             annuler l'édition / fermer l'aide",
		"  l               changer de langue",
		"  ?               afficher/masquer cette aide",
		"  q               quitter",
		"",
		"La config est sauvegardée dans ~/.ratemate/config.json",
	}
	return s
}

// german (Deutsch).
func german() Strings {
	s := english()
	s.AppSubtitle = "Verzögerungs-Proxy für LLM-APIs"

	s.StatusStopped = "● Gestoppt"
	s.StatusPaused = "⏸ Pausiert"
	s.StatusRunning = "● Läuft"

	s.ModeAuto = "Auto"
	s.ModeManual = "Manuell"
	s.ModeSwitchHint = "<m> wechseln"

	s.ConfigTitle = "⚙ Konfiguration"
	s.WindowLabel = "Fenster:"
	s.LimitLabel = "Limit:"
	s.DelayLabel = "Verzögerung:"
	s.RequestsPerShortFmt = "Anfragen pro %s"
	s.ComputedDelayFmt = "→ berechnete Verzögerung: %s zwischen Anfragen"
	s.EffectiveThroughputFmt = "≈ %s effektiver Durchsatz"
	s.MsPerRequest = "ms pro Anfrage"
	s.EffectiveRateFmt = "→ effektive Rate: %s"
	s.ReqPerSecMaxFmt = "%d req/s max"
	s.NudgeHint = "[+]/[-] oder ↑/↓ passt um 100ms an"

	s.StatsTitle = "📊 Live-Statistik"
	s.TotalReqs = "Anfragen gesamt"
	s.Delayed = "Verzögert"
	s.PassThrough = "Direkt"
	s.AvgWait = "Ø Wartezeit"
	s.InOut = "Ein / Aus"
	s.Errors = "Fehler"
	s.LastReq = "Letzte Anfrage"

	s.ActivityTitle = "📡 Letzte Aktivität"
	s.NoRequests = "noch keine Anfragen"

	s.AutoStatusFmt = "auto: %d req / %s → %d ms Abstand"
	s.ManualStatusFmt = "manuell: %d ms Abstand"

	s.AgoFmt = "vor %s"
	s.NeverDash = "—"
	s.Unlimited = "unbegrenzt"
	s.ReqPerMinFmt = "%.0f req/min"
	s.ReqPerHourFmt = "%.0f req/h"
	s.MsUnit = "ms"

	s.KeyStartStop = "start/stop"
	s.KeyMode = "Modus"
	s.KeyPause = "Pause"
	s.KeyReset = "Zurücksetzen"
	s.KeyClear = "Löschen"
	s.KeyTune = "Justieren"
	s.KeyFocus = "Fokus"
	s.KeyEdit = "Editieren"
	s.KeyLang = "Sprache"
	s.KeyHelp = "Hilfe"
	s.KeyQuit = "Beenden"

	s.FailedStart = "✗ Proxy-Start fehlgeschlagen"
	s.ProxyStarted = "● Proxy gestartet"
	s.ProxyStopped = "○ Proxy gestoppt"
	s.PausedMsg = "⏸ Limiter pausiert (Durchreichen)"
	s.ResumedMsg = "▶ Limiter fortgesetzt"
	s.ModeManualSet = "Modus → Manuell"
	s.ModeAutoSet = "Modus → Auto"
	s.StatsReset = "Statistik zurückgesetzt"
	s.ActivityCleared = "Aktivitätsprotokoll gelöscht"
	s.WindowSetFmt = "Fenster → %s"
	s.LimitSetFmt = "Limit → %d"
	s.ManualDelaySetFmt = "manuelle Verzögerung → %d ms"
	s.PortSetFmt = "Port → %d"
	s.StopProxyToChangePort = "Stoppe den Proxy um den Port zu ändern"
	s.LangChangedFmt = "Sprache → %s"

	s.WindowSecond = "Sekunde"
	s.WindowMinute = "Minute"
	s.WindowHour = "Stunde"
	s.WindowShortSecond = "s"
	s.WindowShortMinute = "m"
	s.WindowShortHour = "h"

	s.LangMenuTitle = "🌐  Sprache / Language"
	s.LangMenuHint = "↑/↓ navigieren   enter wählen   esc abbrechen"
	s.LangMenuCurrent = "aktuell"
	s.LangButtonFmt = "🌐 %s"
	s.LangMenuCloseKey = "l"

	s.HelpReturnHint = "? oder esc drücken um zurückzukehren"
	s.HelpLines = []string{
		"⚡ RateMate — Verzögerungs-Proxy für LLM-APIs",
		"",
		"RateMate betreibt einen lokalen HTTP/HTTPS-Proxy, der vor jeder",
		"weitergeleiteten Anfrage eine kleine Verzögerung einfügt, damit",
		"du das Sekunden-/Minuten-/Stunden-Limit deiner LLM-API nie überschreitest.",
		"",
		"Benutzung:",
		"  1. Drücke <s> zum Starten des Proxys (Standard :8080).",
		"  2. Richte deinen Client auf den Proxy, z. B.:",
		"       export HTTPS_PROXY=http://127.0.0.1:8080",
		"       export HTTP_PROXY=http://127.0.0.1:8080",
		"  3. Verzögerung mit <m> (Modus) und den Feldern einstellen. Tab wechselt den Fokus,",
		"     Enter editiert, Esc bricht ab, +/- justiert.",
		"  4. Drücke <p> zum Pausieren (Durchreichen) ohne den Server zu stoppen.",
		"",
		"Tasten:",
		"  s              Proxy starten / stoppen",
		"  m               Auto ↔ Manuell-Modus umschalten",
		"  p               Limiter pausieren / fortsetzen",
		"  r               Zähler & Protokoll zurücksetzen",
		"  c               nur Aktivitätsprotokoll löschen",
		"  ◀/▶ oder </>    Zeitfenster wechseln (Auto)",
		"  +/-              Limit (Auto) oder Verzögerung (Manuell) justieren",
		"  ↑/↓              um ±5 (Auto) oder ±100ms (Manuell) justieren",
		"  tab/shift+tab   Fokus wechseln",
		"  enter           fokussiertes Feld editieren",
		"  esc             Bearbeitung abbrechen / Hilfe schließen",
		"  l               Sprache wechseln",
		"  ?               diese Hilfe ein/aus",
		"  q               beenden",
		"",
		"Konfig wird in ~/.ratemate/config.json gespeichert",
	}
	return s
}

// all holds the full translation table.
var all = map[Lang]Strings{
	EN: english(),
	ZH: chinese(),
	JA: japanese(),
	ES: spanish(),
	FR: french(),
	DE: german(),
}

// Get returns the Strings for a language, falling back to the default when
// the code is unknown.
func Get(l Lang) Strings {
	if s, ok := all[Normalise(l)]; ok {
		return s
	}
	return all[Default]
}

// WindowName returns the localised long name of a limiter window index
// (0 = second, 1 = minute, 2 = hour).
func (s Strings) WindowName(w int) string {
	switch w {
	case 0:
		return s.WindowSecond
	case 1:
		return s.WindowMinute
	case 2:
		return s.WindowHour
	}
	return s.WindowMinute
}

// WindowShort returns the localised short label of a limiter window index.
func (s Strings) WindowShort(w int) string {
	switch w {
	case 0:
		return s.WindowShortSecond
	case 1:
		return s.WindowShortMinute
	case 2:
		return s.WindowShortHour
	}
	return s.WindowShortMinute
}
