#requires -Version 5.1
<#
.SYNOPSIS
    Launch RateMate (TUI by default) via `go run .`
.DESCRIPTION
    Temporarily routes Go module downloads through a China-friendly mirror
    (goproxy.cn) so `go run` works behind the GFW without touching the user's
    persistent environment. The GOPROXY value is scoped to this PowerShell
    process and restored to its previous value (including $null) on exit — it
    never leaks into the parent shell and never gets written to `go env -w`.

    Usage:
      .\scripts\run.ps1                      # launch the TUI
      .\scripts\run.ps1 serve               # run the proxy headless
      .\scripts\run.ps1 serve -port 9090   # pass flags through
      .\scripts\run.ps1 version

    On Windows you may need to allow script execution once:
      powershell -ExecutionPolicy Bypass -File .\scripts\run.ps1
#>
$ErrorActionPreference = 'Stop'

# Resolve the project root regardless of the current working directory.
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
Set-Location $ProjectRoot

# Save and override GOPROXY for this process only.
$prevGoproxy = $env:GOPROXY
$prevHttps   = $env:HTTPS_PROXY
$prevHttp    = $env:HTTP_PROXY
$env:GOPROXY = 'https://goproxy.cn,direct'

try {
    & go run . @args
    $exitCode = $LASTEXITCODE
} finally {
    # Restore the originals — even if go crashed or was interrupted.
    $env:GOPROXY = $prevGoproxy
    $env:HTTPS_PROXY = $prevHttps
    $env:HTTP_PROXY  = $prevHttp
}

exit $exitCode
