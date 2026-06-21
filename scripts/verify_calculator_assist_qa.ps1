param(
  [string]$DumpPath = ".\.codex\qa\assist-dump.exe"
)

$ErrorActionPreference = "Stop"

function Build-Dump([string]$OutPath) {
  $dir = Split-Path -Parent $OutPath
  if ($dir) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
  }
  go build -o $OutPath .\cmd\assistdump
  if ($LASTEXITCODE -ne 0) {
    throw "assistdump build failed"
  }
}

function Parse-DumpLine([string]$Line) {
  $match = [regex]::Match($Line, 'label="(?<label>.*?)"\s+priority=(?<priority>\d+)\s+dangerous=(?<dangerous>true|false)\s+app="(?<app>.*?)"\s+window="(?<window>.*?)"')
  if (-not $match.Success) {
    return $null
  }
  [pscustomobject]@{
    Label = $match.Groups["label"].Value
    Priority = [int]$match.Groups["priority"].Value
    Dangerous = [bool]::Parse($match.Groups["dangerous"].Value)
    App = $match.Groups["app"].Value
    Window = $match.Groups["window"].Value
  }
}

function Invoke-DumpTargets([string]$DumpExe, [switch]$Safe) {
  if ($Safe) {
    $lines = & $DumpExe -safe -delay-ms 250
  }
  else {
    $lines = & $DumpExe -delay-ms 250
  }
  if ($LASTEXITCODE -ne 0) {
    throw "assistdump failed"
  }
  @($lines | ForEach-Object { Parse-DumpLine $_ } | Where-Object { $_ -ne $null })
}

$calculatorTitle = -join @([char]0x96FB, [char]0x5353)

Build-Dump $DumpPath
$dumpExe = (Resolve-Path -LiteralPath $DumpPath).Path

$beforeCalcPids = @(Get-Process | Where-Object { $_.ProcessName -match "Calculator|calc" } | ForEach-Object { $_.Id })
$started = Start-Process -FilePath "calc.exe" -PassThru

try {
  $targets = @()
  for ($attempt = 0; $attempt -lt 10; $attempt++) {
    Start-Sleep -Milliseconds 600
    $targets = @(Invoke-DumpTargets $dumpExe)
    if (@($targets | Where-Object { $_.App -eq "ApplicationFrameHost.exe" -and ($_.Window -eq $calculatorTitle -or $_.Window -like "*Calculator*") }).Count -gt 0) {
      break
    }
  }
  $calculatorTargets = @($targets | Where-Object { $_.App -eq "ApplicationFrameHost.exe" -and ($_.Window -eq $calculatorTitle -or $_.Window -like "*Calculator*") })
  if ($calculatorTargets.Count -eq 0) {
    throw "No Calculator assist targets were detected. labels=$((@($targets | ForEach-Object { $_.App + ':' + $_.Window + ':' + $_.Label })) -join ', ')"
  }

  $safeTargets = @(Invoke-DumpTargets $dumpExe -Safe | Where-Object { $_.App -eq "ApplicationFrameHost.exe" -and ($_.Window -eq $calculatorTitle -or $_.Window -like "*Calculator*") })
  if ($safeTargets.Count -eq 0) {
    throw "Calculator produced targets but no safe targets"
  }
  if (@($safeTargets | Where-Object { $_.Dangerous }).Count -gt 0) {
    throw "Calculator safe list included a dangerous target"
  }

  [pscustomobject]@{
    Window = $calculatorTargets[0].Window
    App = $calculatorTargets[0].App
    TargetCount = $calculatorTargets.Count
    SafeLabels = (@($safeTargets | ForEach-Object { $_.Label }) -join "; ")
    FirstTargets = (@($calculatorTargets | Select-Object -First 8 | ForEach-Object { "$($_.Label):priority=$($_.Priority):dangerous=$($_.Dangerous)" }) -join "; ")
  }
}
finally {
  $after = @(Get-Process | Where-Object {
    $_.ProcessName -match "Calculator|calc" -and $beforeCalcPids -notcontains $_.Id
  })
  foreach ($process in $after) {
    Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
  }
  if ($started -and -not $started.HasExited) {
    Stop-Process -Id $started.Id -Force -ErrorAction SilentlyContinue
  }
}
