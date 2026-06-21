param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$ConfigPath,

  [Parameter(Mandatory = $true)]
  [string]$ExpectedButton,

  [string]$ForbiddenButton = "Delete",

  [string]$DisabledButton = "Apply",

  [string]$FixturePath = ".\.codex\qa\assist-fixture.exe"
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;

public static class MofuMouseAssistFixtureQA {
  public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
  public delegate bool EnumChildProc(IntPtr hWnd, IntPtr lParam);

  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool EnumChildWindows(IntPtr hWnd, EnumChildProc lpEnumFunc, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool GetCursorPos(out POINT point);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")] public static extern int GetWindowTextLengthW(IntPtr hWnd);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowTextW(IntPtr hWnd, StringBuilder text, int count);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint processId);
  [DllImport("user32.dll")] public static extern bool IsWindowEnabled(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, UIntPtr dwExtraInfo);
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);

  public struct POINT {
    public int X;
    public int Y;
  }

  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }

  public static void LeftClick() {
    mouse_event(0x0002, 0, 0, 0, UIntPtr.Zero);
    mouse_event(0x0004, 0, 0, 0, UIntPtr.Zero);
  }

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

  public static string WindowText(IntPtr hWnd) {
    int length = GetWindowTextLengthW(hWnd);
    if (length <= 0) {
      return "";
    }
    StringBuilder text = new StringBuilder(length + 1);
    GetWindowTextW(hWnd, text, text.Capacity);
    return text.ToString();
  }

  public static IntPtr FindChildByText(IntPtr parent, string label) {
    IntPtr found = IntPtr.Zero;
    EnumChildWindows(parent, delegate(IntPtr hWnd, IntPtr lParam) {
      if (WindowText(hWnd) == label) {
        found = hWnd;
        return false;
      }
      return true;
    }, IntPtr.Zero);
    return found;
  }

  public static double Distance(POINT a, POINT b) {
    int dx = a.X - b.X;
    int dy = a.Y - b.Y;
    return Math.Sqrt(dx * dx + dy * dy);
  }
}
"@

function Resolve-RequiredPath([string]$Path) {
  $resolved = Resolve-Path -LiteralPath $Path -ErrorAction Stop
  return $resolved.ProviderPath
}

function Build-Fixture([string]$OutPath) {
  $dir = Split-Path -Parent $OutPath
  if ($dir) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
  }
  go build -ldflags="-H=windowsgui" -o $OutPath .\cmd\assistfixture
}

function Get-ButtonHandle([IntPtr]$Parent, [string]$Label) {
  $button = [MofuMouseAssistFixtureQA]::FindChildByText($Parent, $Label)
  if ($button -eq [IntPtr]::Zero) {
    throw "Button '$Label' was not found"
  }
  return $button
}

function Get-ButtonCenter([IntPtr]$Parent, [string]$Label, [switch]$AllowDisabled) {
  $button = Get-ButtonHandle $Parent $Label
  if (-not [MofuMouseAssistFixtureQA]::IsWindowEnabled($button)) {
    if ($AllowDisabled) {
      $rect = New-Object MofuMouseAssistFixtureQA+RECT
      [MofuMouseAssistFixtureQA]::GetWindowRect($button, [ref]$rect) | Out-Null
      $point = New-Object MofuMouseAssistFixtureQA+POINT
      $point.X = [int](($rect.Left + $rect.Right) / 2)
      $point.Y = [int](($rect.Top + $rect.Bottom) / 2)
      return $point
    }
    throw "Button '$Label' is disabled"
  }
  $rect = New-Object MofuMouseAssistFixtureQA+RECT
  [MofuMouseAssistFixtureQA]::GetWindowRect($button, [ref]$rect) | Out-Null
  $point = New-Object MofuMouseAssistFixtureQA+POINT
  $point.X = [int](($rect.Left + $rect.Right) / 2)
  $point.Y = [int](($rect.Top + $rect.Bottom) / 2)
  return $point
}

$exe = Resolve-RequiredPath $ExePath
$config = Resolve-RequiredPath $ConfigPath
Build-Fixture $FixturePath
$fixtureExe = Resolve-RequiredPath $FixturePath

$fixture = $null
$app = $null
try {
  $fixture = Start-Process -FilePath $fixtureExe -PassThru
  $fixtureHwnd = [IntPtr]::Zero
  for ($i = 0; $i -lt 80; $i++) {
    Start-Sleep -Milliseconds 100
    $fixtureHwnd = [MofuMouseAssistFixtureQA]::FindVisibleWindow($fixture.Id)
    if ($fixtureHwnd -ne [IntPtr]::Zero) {
      break
    }
  }
  if ($fixtureHwnd -eq [IntPtr]::Zero) {
    throw "Assist fixture window was not found"
  }

  [MofuMouseAssistFixtureQA]::ShowWindow($fixtureHwnd, 5) | Out-Null
  [MofuMouseAssistFixtureQA]::SetForegroundWindow($fixtureHwnd) | Out-Null
  $fixtureRect = New-Object MofuMouseAssistFixtureQA+RECT
  [MofuMouseAssistFixtureQA]::GetWindowRect($fixtureHwnd, [ref]$fixtureRect) | Out-Null

  $expectedPoint = Get-ButtonCenter $fixtureHwnd $ExpectedButton
  $forbiddenPoint = $null
  if ($ForbiddenButton) {
    $forbiddenPoint = Get-ButtonCenter $fixtureHwnd $ForbiddenButton
  }
  $disabledPoint = $null
  if ($DisabledButton) {
    $disabledHandle = Get-ButtonHandle $fixtureHwnd $DisabledButton
    if ([MofuMouseAssistFixtureQA]::IsWindowEnabled($disabledHandle)) {
      throw "Button '$DisabledButton' was expected to be disabled"
    }
    $disabledPoint = Get-ButtonCenter $fixtureHwnd $DisabledButton -AllowDisabled
  }

  $app = Start-Process -FilePath $exe -ArgumentList @("-config", $config) -PassThru
  Start-Sleep -Milliseconds 1200
  for ($i = 0; $i -lt 6; $i++) {
    [MofuMouseAssistFixtureQA]::ShowWindow($fixtureHwnd, 5) | Out-Null
    [MofuMouseAssistFixtureQA]::SetForegroundWindow($fixtureHwnd) | Out-Null
    Start-Sleep -Milliseconds 100
  }
  [MofuMouseAssistFixtureQA]::SetCursorPos(($fixtureRect.Left + 40), ($fixtureRect.Top + 16)) | Out-Null
  [MofuMouseAssistFixtureQA]::LeftClick()
  Start-Sleep -Milliseconds 250

  $origin = New-Object MofuMouseAssistFixtureQA+POINT
  $origin.X = $fixtureRect.Left + 35
  $origin.Y = $fixtureRect.Bottom - 35
  [MofuMouseAssistFixtureQA]::SetCursorPos($origin.X, $origin.Y) | Out-Null
  Start-Sleep -Milliseconds 100
  [MofuMouseAssistFixtureQA]::SetForegroundWindow($fixtureHwnd) | Out-Null

  Start-Sleep -Milliseconds 4200

  $final = New-Object MofuMouseAssistFixtureQA+POINT
  [MofuMouseAssistFixtureQA]::GetCursorPos([ref]$final) | Out-Null

  $distanceToExpected = [MofuMouseAssistFixtureQA]::Distance($final, $expectedPoint)
  if ($distanceToExpected -gt 42) {
    throw "Cursor did not move to expected '$ExpectedButton'. distance=$distanceToExpected final=$($final.X),$($final.Y) expected=$($expectedPoint.X),$($expectedPoint.Y)"
  }

  $distanceToForbidden = $null
  if ($forbiddenPoint -ne $null) {
    $distanceToForbidden = [MofuMouseAssistFixtureQA]::Distance($final, $forbiddenPoint)
    if ($ExpectedButton -ne $ForbiddenButton -and $distanceToForbidden -lt 42) {
      throw "Cursor moved near forbidden '$ForbiddenButton'. distance=$distanceToForbidden final=$($final.X),$($final.Y) forbidden=$($forbiddenPoint.X),$($forbiddenPoint.Y)"
    }
  }

  $distanceToDisabled = $null
  if ($disabledPoint -ne $null) {
    $distanceToDisabled = [MofuMouseAssistFixtureQA]::Distance($final, $disabledPoint)
    if ($ExpectedButton -ne $DisabledButton -and $distanceToDisabled -lt 42) {
      throw "Cursor moved near disabled '$DisabledButton'. distance=$distanceToDisabled final=$($final.X),$($final.Y) disabled=$($disabledPoint.X),$($disabledPoint.Y)"
    }
  }

  [pscustomobject]@{
    Config = $config
    ExpectedButton = $ExpectedButton
    ForbiddenButton = $ForbiddenButton
    DisabledButton = $DisabledButton
    FinalX = $final.X
    FinalY = $final.Y
    ExpectedX = $expectedPoint.X
    ExpectedY = $expectedPoint.Y
    ForbiddenDistance = if ($distanceToForbidden -eq $null) { $null } else { [math]::Round($distanceToForbidden, 2) }
    DisabledDistance = if ($distanceToDisabled -eq $null) { $null } else { [math]::Round($distanceToDisabled, 2) }
    ExpectedDistance = [math]::Round($distanceToExpected, 2)
  }
}
finally {
  if ($app -and -not $app.HasExited) {
    Stop-Process -Id $app.Id -Force
  }
  if ($fixture -and -not $fixture.HasExited) {
    Stop-Process -Id $fixture.Id -Force
  }
}
