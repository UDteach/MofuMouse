param()

$ErrorActionPreference = "Stop"

go test ./internal/core ./internal/winapp -run "LaunchAtLogin|SettingsRoundTrip|DefaultSettingsAreSafe|StateSnapshotIncludesRuntimeSettings" -count=1
if ($LASTEXITCODE -ne 0) {
  throw "autostart QA failed"
}

[pscustomobject]@{
  RegistryScope = "HKCU:\Software\MofuMouse\QA\AutoStart\*"
  RealRunKeyTouched = $false
  Result = "PASS"
}
