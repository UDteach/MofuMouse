param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$ConfigPath,

  [Parameter(Mandatory = $true)]
  [string]$OutputPath
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public static class MofuMouseReturnPoseQA {
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
      RECT buttonRect;
      if (!GetWindowRect(hWnd, out buttonRect)) {
        return true;
      }
      result.X = buttonRect.Left + ((buttonRect.Right - buttonRect.Left) / 2);
      result.Y = buttonRect.Top + ((buttonRect.Bottom - buttonRect.Top) / 2);
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

Add-Type -AssemblyName System.Drawing

function Capture-Crop([string]$Path, [int]$X, [int]$Y, [int]$W, [int]$H) {
  $dir = Split-Path -Parent $Path
  if ($dir) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
  }
  $bitmap = New-Object System.Drawing.Bitmap($W, $H)
  $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
  try {
    $graphics.CopyFromScreen($X, $Y, 0, 0, [System.Drawing.Size]::new($W, $H))
    $bitmap.Save($Path, [System.Drawing.Imaging.ImageFormat]::Png)
  }
  finally {
    $graphics.Dispose()
    $bitmap.Dispose()
  }
}

$process = Start-Process -FilePath $ExePath -ArgumentList @("--open-settings", "--legacy-settings", "-config", $ConfigPath) -PassThru
try {
  $hwnd = [System.IntPtr]::Zero
  for ($i = 0; $i -lt 50; $i++) {
    Start-Sleep -Milliseconds 100
    $hwnd = [MofuMouseReturnPoseQA]::FindSettingsWindow($process.Id)
    if ($hwnd -ne [System.IntPtr]::Zero) {
      break
    }
  }
  if ($hwnd -eq [System.IntPtr]::Zero) {
    throw "MofuMouse settings window was not found"
  }

  [MofuMouseReturnPoseQA]::ShowWindow($hwnd, 5) | Out-Null
  [MofuMouseReturnPoseQA]::SetForegroundWindow($hwnd) | Out-Null
  $rect = New-Object MofuMouseReturnPoseQA+RECT
  [MofuMouseReturnPoseQA]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
  [MofuMouseReturnPoseQA]::SetCursorPos(($rect.Left + 80), ($rect.Top + 14)) | Out-Null
  [MofuMouseReturnPoseQA]::LeftClick()
  Start-Sleep -Milliseconds 250

  $origin = New-Object MofuMouseReturnPoseQA+POINT
  $origin.X = $rect.Left + 360
  $origin.Y = $rect.Top + 560
  [MofuMouseReturnPoseQA]::SetCursorPos($origin.X, $origin.Y) | Out-Null

  $expectedTarget = New-Object MofuMouseReturnPoseQA+POINT
  $saveLabel = [string]([char]0x4FDD) + [string]([char]0x5B58)
  $expectedTarget = [MofuMouseReturnPoseQA]::FindButtonCenter($hwnd, $saveLabel)
  if ($expectedTarget.X -eq [int]::MinValue) {
    throw 'Save button was not found'
  }

  $current = New-Object MofuMouseReturnPoseQA+POINT
  $targetReached = $false
  for ($i = 0; $i -lt 60; $i++) {
    Start-Sleep -Milliseconds 100
    [MofuMouseReturnPoseQA]::GetCursorPos([ref]$current) | Out-Null
    if ([MofuMouseReturnPoseQA]::Distance($current, $expectedTarget) -le 36) {
      $targetReached = $true
      break
    }
  }
  if (-not $targetReached) {
    throw "Cursor did not reach target before return capture. current=$($current.X),$($current.Y) target=$($expectedTarget.X),$($expectedTarget.Y)"
  }

  $returned = $false
  for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Milliseconds 60
    [MofuMouseReturnPoseQA]::GetCursorPos([ref]$current) | Out-Null
    if ([MofuMouseReturnPoseQA]::Distance($current, $origin) -le 18) {
      $returned = $true
      break
    }
  }
  if (-not $returned) {
    throw "Cursor did not return before capture. current=$($current.X),$($current.Y) origin=$($origin.X),$($origin.Y)"
  }

  Start-Sleep -Milliseconds 260
  Capture-Crop $OutputPath ($rect.Left - 160) ($rect.Top - 120) 1040 920

  [pscustomobject]@{
    OutputPath = (Resolve-Path -LiteralPath $OutputPath).Path
    OriginX = $origin.X
    OriginY = $origin.Y
    TargetX = $expectedTarget.X
    TargetY = $expectedTarget.Y
  }
}
finally {
  if ($process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
  }
}
