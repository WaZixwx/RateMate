#requires -Version 5.1
<#
.SYNOPSIS
    Remove every trace of RateMate from a Windows system (root-and-branch).
.DESCRIPTION
    Cleans: the ~/.ratemate config directory, the `go install`-ed binary,
    in-repo build artifacts (bin/, dist/), and optionally RateMate's entries
    in the Go module/build cache. Also detects and offers to remove
    user-scoped HTTPS_PROXY / HTTP_PROXY environment variables that point at
    the local proxy.

    Safe to re-run — each piece is checked for existence first.

    Usage:
      .\scripts\uninstall.ps1              # standard purge
      .\scripts\uninstall.ps1 -Purge       # also wipe Go module/build cache
      .\scripts\uninstall.ps1 -DryRun       # show what would be removed

    The script does NOT touch:
      - other Go projects' module cache
      - your Go installation itself
#>
[CmdletBinding()]
param(
    [switch]$Purge,
    [switch]$DryRun
)
$ErrorActionPreference = 'Stop'

function Write-Header($t) { Write-Host "`n=== $t ===" -ForegroundColor Cyan }
function Write-Gone($t)  { Write-Host "  removed   $t" -ForegroundColor Green }
function Write-Absent($t){ Write-Host "  absent    $t" -ForegroundColor DarkGray }
function Write-Warn2($t) { Write-Host "  warn      $t" -ForegroundColor Yellow }
function Write-Info2($t) { Write-Host "  $t" }

function Remove-Path($p) {
    if (Test-Path -LiteralPath $p) {
        if ($DryRun) {
            Write-Host "  would-rm  $p" -ForegroundColor DarkGray
        } else {
            Remove-Item -Recurse -Force -LiteralPath $p
            Write-Gone $p
        }
    } else {
        Write-Absent $p
    }
}

Write-Host "=== RateMate uninstall (root-and-branch) ===" -ForegroundColor Cyan
if ($DryRun) { Write-Info2 "(dry-run mode — nothing will actually be deleted)" }

# ---- 1. config directory: %USERPROFILE%\.ratemate -------------------------
Write-Header "1. configuration"
$cfgDir = Join-Path $env:USERPROFILE '.ratemate'
Remove-Path $cfgDir

# ---- 2. installed binary (go install) -------------------------------------
Write-Header "2. installed binary"
$gopath = if ($env:GOPATH) { $env:GOPATH } else { Join-Path $env:USERPROFILE 'go' }
$binPath = Join-Path $gopath 'bin\ratemate.exe'
Remove-Path $binPath

# ---- 3. in-repo build artifacts -------------------------------------------
Write-Header "3. in-repo build artifacts"
foreach ($d in 'bin','dist') {
    $p = Join-Path $PSScriptRoot "..\$d"
    Remove-Path $p
}

# ---- 4. proxy environment variables ---------------------------------------
Write-Header "4. proxy environment variables"
$foundEnv = $false
foreach ($v in 'HTTPS_PROXY','HTTP_PROXY') {
    # Check the persistent (User) scope — the one `setx` / PowerShell env writes to.
    $userVal = [Environment]::GetEnvironmentVariable($v, 'User')
    if ($userVal -and $userVal -match '127\.0\.0\.1:8080') {
        Write-Warn2 "$v=$userVal (User-scope, persistent)"
        Write-Info2 "  PowerShell:  Remove-Item Env:$v ; [Environment]::SetEnvironmentVariable('$v', `$null, 'User')"
        $foundEnv = $true
        if (-not $DryRun) {
            $ans = Read-Host "  remove it now? [y/N]"
            if ($ans -eq 'y' -or $ans -eq 'Y') {
                [Environment]::SetEnvironmentVariable($v, $null, 'User')
                Write-Gone "user-scope $v"
            }
        }
    }
    # Also check the process scope (current session).
    $procVal = [Environment]::GetEnvironmentVariable($v, 'Process')
    if ($procVal -and $procVal -match '127\.0\.0\.1:8080') {
        Write-Warn2 "$v=$procVal (current session only — close the terminal to clear)"
        $foundEnv = $true
    }
}
if (-not $foundEnv) { Write-Info2 "no RateMate proxy vars detected" }

# ---- 5. optional: Go module/build cache -----------------------------------
if ($Purge) {
    Write-Header "5. Go module/build cache (--Purge)"
    $goCmd = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goCmd) {
        Write-Warn2 "go not found on PATH — skipping cache purge"
    } else {
        $gomodcache = (go env GOMODCACHE)
        $gocache    = (go env GOCACHE)
        # Go encodes uppercase as ! + lowercase: github.com/WaZixwx -> github.com/!wa!zixwx
        $modDir = Join-Path $gomodcache 'github.com\!wa!zixwx'
        Remove-Path $modDir
        if (Test-Path $gocache) {
            if ($DryRun) {
                Write-Host "  would-run  go clean -cache" -ForegroundColor DarkGray
            } else {
                go clean -cache 2>$null
                Write-Gone "go build cache ($gocache)"
            }
        } else {
            Write-Absent $gocache
        }
    }
}

Write-Host ""
Write-Host "=== done. RateMate is fully removed. ===" -ForegroundColor Green
if ($foundEnv) {
    Write-Host ""
    Write-Info2 "Note: a couple of env vars live in your shell/registry."
    Write-Info2 "Run the commands above (or answer 'y' to the prompts) to finish."
}
