param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$ConfigPath,

  [switch]$ExpectReturn,

  [switch]$SimulateUserMove
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public static class MofuMouseTargetMoveQA {
  public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool EnumChildWindows(IntPtr hWndParent, EnumWindowsProc lpEnumFunc, IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll")] public static extern bool GetCursorPos(out POINT point);
  [DllImport("user32.dll")] public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, UIntPtr dwExtraInfo);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint processId);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetClassName(IntPtr hWnd, StringBuilder text, int count);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool IsWindowEnabled(IntPtr hWnd);
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
  public static POINT FindButtonCenter(IntPtr parent, string labelPart) {
    POINT result = new POINT { X = Int32.MinValue, Y = Int32.MinValue };
    EnumChildWindows(parent, delegate(IntPtr hWnd, IntPtr lParam) {
      if (!IsWindowVisible(hWnd) || !IsWindowEnabled(hWnd)) {
        return true;
      }
      StringBuilder className = new StringBuilder(256);
      GetClassName(hWnd, className, className.Capacity);
      if (className.ToString() != "Button") {
        return true;
      }
      StringBuilder text = new StringBuilder(256);
      GetWindowText(hWnd, text, text.Capacity);
      if (!text.ToString().Contains(labelPart)) {
        return true;
      }
      RECT rect;
      if (!GetWindowRect(hWnd, out rect)) {
        return true;
      }
      result.X = rect.Left + ((rect.Right - rect.Left) / 2);
      result.Y = rect.Top + ((rect.Bottom - rect.Top) / 2);
      return false;
    }, IntPtr.Zero);
    return result;
  }
  public static IntPtr FindSettingsWindow(int processId) {
    IntPtr found = IntPtr.Zero;
    EnumWindows(delegate(IntPtr hWnd, IntPtr lParam) {
      uint windowProcessId;
      GetWindowThreadProcessId(hWnd, out windowProcessId);
      if (windowProcessId == (uint)processId && IsWindowVisible(hWnd)) {
        StringBuilder className = new StringBuilder(256);
        GetClassName(hWnd, className, className.Capacity);
        if (className.ToString() == "MofuMouseSettingsWindow") {
          found = hWnd;
          return false;
        }
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

$process = Start-Process -FilePath $ExePath -ArgumentList @("--open-settings", "--legacy-settings", "-config", $ConfigPath) -PassThru
try {
  $hwnd = [System.IntPtr]::Zero
  for ($i = 0; $i -lt 50; $i++) {
    Start-Sleep -Milliseconds 100
    $hwnd = [MofuMouseTargetMoveQA]::FindSettingsWindow($process.Id)
    if ($hwnd -ne [System.IntPtr]::Zero) {
      break
    }
  }
  if ($hwnd -eq [System.IntPtr]::Zero) {
    throw "MofuMouse settings window was not found"
  }

  [MofuMouseTargetMoveQA]::ShowWindow($hwnd, 5) | Out-Null
  [MofuMouseTargetMoveQA]::SetForegroundWindow($hwnd) | Out-Null
  $rect = New-Object MofuMouseTargetMoveQA+RECT
  [MofuMouseTargetMoveQA]::GetWindowRect($hwnd, [ref]$rect) | Out-Null

  # SetForegroundWindow can be denied by Windows foreground-lock rules.
  # The QA owns this temporary window, so click its title bar once to make
  # the foreground window deterministic before scanning safe targets.
  [MofuMouseTargetMoveQA]::SetCursorPos(($rect.Left + 40), ($rect.Top + 16)) | Out-Null
  [MofuMouseTargetMoveQA]::LeftClick()
  Start-Sleep -Milliseconds 250

  $origin = New-Object MofuMouseTargetMoveQA+POINT
  $origin.X = $rect.Left + 360
  $origin.Y = $rect.Top + 560
  [MofuMouseTargetMoveQA]::SetCursorPos($origin.X, $origin.Y) | Out-Null

  $userPoint = New-Object MofuMouseTargetMoveQA+POINT
  $userPoint.X = $origin.X + 120
  $userPoint.Y = $origin.Y - 80

  $saveLabelPart = -join ([char]0x4fdd, [char]0x5b58)
  $expectedTarget = [MofuMouseTargetMoveQA]::FindButtonCenter($hwnd, $saveLabelPart)
  if ($expectedTarget.X -eq [int]::MinValue) {
    throw "Expected safe save button was not found"
  }

  if ($SimulateUserMove) {
    $targetReached = $false
    $current = New-Object MofuMouseTargetMoveQA+POINT
    for ($i = 0; $i -lt 60; $i++) {
      Start-Sleep -Milliseconds 100
      [MofuMouseTargetMoveQA]::GetCursorPos([ref]$current) | Out-Null
      if ([MofuMouseTargetMoveQA]::Distance($current, $expectedTarget) -le 36) {
        $targetReached = $true
        break
      }
    }
    if (-not $targetReached) {
      throw "Cursor did not reach target before simulated user movement. current=$($current.X),$($current.Y) target=$($expectedTarget.X),$($expectedTarget.Y)"
    }
    [MofuMouseTargetMoveQA]::SetCursorPos($userPoint.X, $userPoint.Y) | Out-Null
    Start-Sleep -Milliseconds 1000
  }
  else {
    Start-Sleep -Milliseconds 2400
  }

  $final = New-Object MofuMouseTargetMoveQA+POINT
  [MofuMouseTargetMoveQA]::GetCursorPos([ref]$final) | Out-Null

  $distanceToTarget = [MofuMouseTargetMoveQA]::Distance($final, $expectedTarget)
  $distanceToOrigin = [MofuMouseTargetMoveQA]::Distance($final, $origin)
  $distanceToUserPoint = [MofuMouseTargetMoveQA]::Distance($final, $userPoint)

  if ($SimulateUserMove) {
    if ($distanceToUserPoint -gt 18) {
      throw "Cursor return was not canceled by user movement. distanceToUserPoint=$distanceToUserPoint final=$($final.X),$($final.Y) userPoint=$($userPoint.X),$($userPoint.Y)"
    }
  }
  elseif ($ExpectReturn) {
    if ($distanceToOrigin -gt 18) {
      throw "Cursor did not return to origin. distanceToOrigin=$distanceToOrigin final=$($final.X),$($final.Y) origin=$($origin.X),$($origin.Y)"
    }
  }
  else {
    if ($distanceToTarget -gt 36) {
      throw "Cursor did not move to safe target. distanceToTarget=$distanceToTarget final=$($final.X),$($final.Y) target=$($expectedTarget.X),$($expectedTarget.Y)"
    }
  }

  [pscustomobject]@{
    ExpectReturn = [bool]$ExpectReturn
    SimulateUserMove = [bool]$SimulateUserMove
    FinalX = $final.X
    FinalY = $final.Y
    OriginX = $origin.X
    OriginY = $origin.Y
    TargetX = $expectedTarget.X
    TargetY = $expectedTarget.Y
    UserPointX = $userPoint.X
    UserPointY = $userPoint.Y
    DistanceToTarget = [math]::Round($distanceToTarget, 2)
    DistanceToOrigin = [math]::Round($distanceToOrigin, 2)
    DistanceToUserPoint = [math]::Round($distanceToUserPoint, 2)
  }
}
finally {
  if ($process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
  }
}
