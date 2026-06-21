param(
  [string]$DumpPath = ".\.codex\qa\assist-dump.exe"
)

$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.Windows.Forms

Add-Type @"
using System;
using System.Runtime.InteropServices;

public static class MofuMouseNotepadAssistQA {
  public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);

  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint processId);

  public static IntPtr FindVisibleWindow(int processId) {
    IntPtr found = IntPtr.Zero;
    EnumWindows(delegate(IntPtr hWnd, IntPtr lParam) {
      uint windowProcessId;
      GetWindowThreadProcessId(hWnd, out windowProcessId);
      if (windowProcessId == (uint)processId && IsWindowVisible(hWnd)) {
        found = hWnd;
        return false;
      }
      return true;
    }, IntPtr.Zero);
    return found;
  }
}
"@

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

function Test-AnyPattern([string]$Value, [string[]]$Patterns) {
  $lower = $Value.ToLowerInvariant()
  foreach ($pattern in $Patterns) {
    if ($lower.Contains($pattern.ToLowerInvariant())) {
      return $true
    }
  }
  return $false
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

$dontSaveCurly = "don" + [char]0x2019 + "t save"
$jaDontSave = -join @([char]0x4FDD, [char]0x5B58, [char]0x3057, [char]0x306A, [char]0x3044)
$jaWithoutSaving = -join @([char]0x4FDD, [char]0x5B58, [char]0x305B, [char]0x305A)
$jaDiscard = -join @([char]0x7834, [char]0x68C4)
$dangerPatterns = @("don't save", "dont save", $dontSaveCurly, "do not save", "discard", $jaDontSave, $jaWithoutSaving, $jaDiscard)

Build-Dump $DumpPath
$dumpExe = (Resolve-Path -LiteralPath $DumpPath).Path

$notepad = $null
$visibleNotepad = $null
$ownedPids = @()
try {
  $beforePids = @(Get-Process -Name Notepad -ErrorAction SilentlyContinue | ForEach-Object { $_.Id })
  $notepad = Start-Process -FilePath "notepad.exe" -PassThru
  $ownedPids += $notepad.Id

  $hwnd = [IntPtr]::Zero
  for ($i = 0; $i -lt 100; $i++) {
    Start-Sleep -Milliseconds 100
    $candidates = @(Get-Process -Name Notepad -ErrorAction SilentlyContinue | Where-Object {
      $beforePids -notcontains $_.Id -and $_.MainWindowHandle -ne 0
    })
    if ($candidates.Count -gt 0) {
      $visibleNotepad = $candidates[0]
      if ($ownedPids -notcontains $visibleNotepad.Id) {
        $ownedPids += $visibleNotepad.Id
      }
      $hwnd = [IntPtr]$visibleNotepad.MainWindowHandle
      break
    }
    $hwnd = [MofuMouseNotepadAssistQA]::FindVisibleWindow($notepad.Id)
    if ($hwnd -ne [IntPtr]::Zero) {
      break
    }
  }
  if ($hwnd -eq [IntPtr]::Zero) {
    throw "Notepad window was not found"
  }

  $typed = $false
  for ($attempt = 0; $attempt -lt 5; $attempt++) {
    [MofuMouseNotepadAssistQA]::ShowWindow($hwnd, 5) | Out-Null
    [MofuMouseNotepadAssistQA]::SetForegroundWindow($hwnd) | Out-Null
    Start-Sleep -Milliseconds 500
    [System.Windows.Forms.SendKeys]::SendWait("MofuMouse QA unsaved dialog")
    Start-Sleep -Milliseconds 700
    if ($visibleNotepad -ne $null) {
      $current = Get-Process -Id $visibleNotepad.Id -ErrorAction SilentlyContinue
      if ($current -and ($current.MainWindowTitle -like "*MofuMouse QA unsaved dialog*" -or $current.MainWindowTitle -like "`**")) {
        $typed = $true
        break
      }
    }
  }
  if (-not $typed) {
    throw "Notepad did not enter an unsaved text state"
  }

  $targets = @()
  $dangerTargets = @()
  for ($closeAttempt = 0; $closeAttempt -lt 4; $closeAttempt++) {
    [MofuMouseNotepadAssistQA]::SetForegroundWindow($hwnd) | Out-Null
    Start-Sleep -Milliseconds 250
    [System.Windows.Forms.SendKeys]::SendWait("%{F4}")
    for ($scanAttempt = 0; $scanAttempt -lt 8; $scanAttempt++) {
      Start-Sleep -Milliseconds 500
      $targets = @(Invoke-DumpTargets $dumpExe)
      $dangerTargets = @($targets | Where-Object { Test-AnyPattern $_.Label $dangerPatterns })
      if ($dangerTargets.Count -gt 0) {
        break
      }
    }
    if ($dangerTargets.Count -gt 0) {
      break
    }
  }

  if ($targets.Count -eq 0) {
    throw "No assist targets were detected in Notepad dialog"
  }
  if ($dangerTargets.Count -eq 0) {
    throw "No discard/dont-save target was found in Notepad dialog. labels=$((@($targets | ForEach-Object { $_.Label })) -join ', ')"
  }
  if (@($dangerTargets | Where-Object { -not $_.Dangerous }).Count -gt 0) {
    throw "A discard/dont-save target was not marked dangerous: $((@($dangerTargets | Where-Object { -not $_.Dangerous } | ForEach-Object { $_.Label })) -join ', ')"
  }

  $safeTargets = @()
  for ($safeAttempt = 0; $safeAttempt -lt 6; $safeAttempt++) {
    [MofuMouseNotepadAssistQA]::SetForegroundWindow($hwnd) | Out-Null
    Start-Sleep -Milliseconds 350
    $safeTargets = @(Invoke-DumpTargets $dumpExe -Safe)
    if ($safeTargets.Count -gt 0) {
      break
    }
  }
  if ($safeTargets.Count -eq 0) {
    throw "No safe targets were detected in Notepad dialog"
  }
  $safeDanger = @($safeTargets | Where-Object { Test-AnyPattern $_.Label $dangerPatterns })
  if ($safeDanger.Count -gt 0) {
    throw "Dangerous discard/dont-save target appeared in safe list: $((@($safeDanger | ForEach-Object { $_.Label })) -join ', ')"
  }

  [pscustomobject]@{
    Window = if ($targets.Count -gt 0) { $targets[0].Window } else { "" }
    App = if ($targets.Count -gt 0) { $targets[0].App } else { "" }
    TargetLabels = (@($targets | ForEach-Object { "$($_.Label):dangerous=$($_.Dangerous):priority=$($_.Priority)" }) -join "; ")
    SafeLabels = (@($safeTargets | ForEach-Object { $_.Label }) -join "; ")
    DangerousDiscardLabels = (@($dangerTargets | ForEach-Object { $_.Label }) -join "; ")
  }
}
finally {
  foreach ($ownedPid in ($ownedPids | Select-Object -Unique)) {
    Stop-Process -Id $ownedPid -Force -ErrorAction SilentlyContinue
  }
}
