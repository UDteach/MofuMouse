param()

$ErrorActionPreference = "Stop"

go test ./internal/core ./internal/winapp ./cmd/mofumouse . -run "QuickScroll|SettingsRoundTrip|ControlCenterSavesSettings" -count=1
if ($LASTEXITCODE -ne 0) {
  throw "quick scroll QA failed"
}

[pscustomobject]@{
  Scope = "core/winapp quick scroll allowlist"
  SendsWheel = $false
  Result = "PASS"
}
