@echo off
REM scripts/run.cmd — launch RateMate via `go run .` on Windows (cmd.exe)
REM
REM Temporarily sets GOPROXY to a China-friendly mirror so module downloads
REM work behind the GFW. `setlocal`/`endlocal` scope the change to this
REM script — it never leaks into the calling shell or persists to the system.
REM
REM Usage:
REM   scripts\run.cmd               (launch the TUI)
REM   scripts\run.cmd serve         (run the proxy headless)
REM   scripts\run.cmd serve -port 9090 -mode manual

setlocal
set "GOPROXY=https://goproxy.cn,direct"
go run . %*
endlocal
