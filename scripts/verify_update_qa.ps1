param()

$ErrorActionPreference = "Stop"

go test ./internal/update ./internal/winapp ./cmd/mofumouse -count=1
if ($LASTEXITCODE -ne 0) {
  throw "update package QA failed"
}

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("mofumouse-update-qa-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempRoot | Out-Null
try {
  $current = Join-Path $tempRoot "mofumouse-x64.exe"
  $asset = Join-Path $tempRoot "downloaded.exe"
  $backupDir = Join-Path $tempRoot "rollback"
  Set-Content -LiteralPath $current -Value "old exe" -NoNewline
  Set-Content -LiteralPath $asset -Value "new exe" -NoNewline

  go run ./cmd/mofumouse --wait-pid 0 --install-update $asset --install-target $current --install-backup-dir $backupDir | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "update CLI install failed"
  }
  $installed = Get-Content -LiteralPath $current -Raw
  if ($installed -ne "new exe") {
    throw "update CLI install wrote unexpected content: $installed"
  }
  $backup = Get-ChildItem -LiteralPath $backupDir -Filter "*.bak" | Select-Object -First 1
  if ($null -eq $backup) {
    throw "update CLI did not create rollback backup"
  }
  $backupContent = Get-Content -LiteralPath $backup.FullName -Raw
  if ($backupContent -ne "old exe") {
    throw "rollback backup has unexpected content: $backupContent"
  }

  go run ./cmd/mofumouse --rollback-update $backup.FullName --install-target $current | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "update CLI rollback failed"
  }
  $rolledBack = Get-Content -LiteralPath $current -Raw
  if ($rolledBack -ne "old exe") {
    throw "update CLI rollback wrote unexpected content: $rolledBack"
  }
} finally {
  $resolvedTemp = (Resolve-Path -LiteralPath $tempRoot).Path
  $systemTemp = ([System.IO.Path]::GetTempPath()).TrimEnd('\')
  if ($resolvedTemp.StartsWith($systemTemp, [System.StringComparison]::OrdinalIgnoreCase)) {
    Remove-Item -LiteralPath $resolvedTemp -Recurse -Force
  }
}

[pscustomobject]@{
  Scope = "internal/update"
  ExternalNetwork = $false
  InstallRollback = $true
  HelperCopy = $true
  Result = "PASS"
}
