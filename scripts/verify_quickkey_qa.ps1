param()

$ErrorActionPreference = "Stop"

go test ./internal/core ./internal/winapp ./cmd/mofumouse . -run "QuickKey|SettingsRoundTrip|ControlCenterSavesSettings" -count=1
if ($LASTEXITCODE -ne 0) {
  throw "quick key QA failed"
}

[pscustomobject]@{
  Scope = "core/winapp quick key allowlist"
  SendsKeys = $false
  Result = "PASS"
}
