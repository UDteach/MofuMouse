$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

function Invoke-Native {
  param(
    [Parameter(Mandatory = $true)]
    [string]$FilePath,

    [string[]]$Arguments = @()
  )

  & $FilePath @Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "Command failed with exit code ${LASTEXITCODE}: $FilePath $($Arguments -join ' ')"
  }
}

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

Invoke-Native "go" @("test", "./...")

$env:GOOS = "windows"
$env:GOARCH = "386"
Invoke-Native "go" @("build", "-ldflags=-H=windowsgui", "-o", "dist\mofumouse-x86.exe", "./cmd/mofumouse")

$env:GOARCH = "amd64"
Invoke-Native "go" @("build", "-ldflags=-H=windowsgui", "-o", "dist\mofumouse-x64.exe", "./cmd/mofumouse")

if (Get-Command wails -ErrorAction SilentlyContinue) {
  Invoke-Native "wails" @("build", "-clean", "-platform", "windows/amd64", "-o", "mofumouse-control-x64.exe", "-webview2", "browser")
  Copy-Item -LiteralPath "build\bin\mofumouse-control-x64.exe" -Destination "dist\mofumouse-control-x64.exe" -Force
}
else {
  Write-Warning "wails CLI not found; skipping modern control center build."
}

Invoke-Native "go" @("build", "-ldflags=-H=windowsgui", "-o", "mofumouse.exe", "./cmd/mofumouse")
Invoke-Native ".\mofumouse.exe" @("--smoke")

Get-ChildItem "dist\mofumouse-*.exe" | Select-Object Name, Length, LastWriteTime
