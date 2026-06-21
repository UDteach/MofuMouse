param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$ControlPath,

  [Parameter(Mandatory = $true)]
  [string]$ConfigPath,

  [Parameter(Mandatory = $true)]
  [string]$OutputPath
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System.Runtime.InteropServices;
public static class MofuMouseUiaSniffAssistQA {
  public delegate bool EnumWindowsProc(System.IntPtr hWnd, System.IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, System.IntPtr lParam);
  [DllImport("user32.dll")] public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, System.UIntPtr dwExtraInfo);
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(System.IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool ShowWindow(System.IntPtr hWnd, int nCmdShow);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(System.IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(System.IntPtr hWnd, out uint processId);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(System.IntPtr hWnd);
  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }
  public static System.IntPtr FindVisibleWindow(int processId) {
    System.IntPtr found = System.IntPtr.Zero;
    EnumWindows(delegate(System.IntPtr hWnd, System.IntPtr lParam) {
      uint windowProcessId;
      GetWindowThreadProcessId(hWnd, out windowProcessId);
      if (windowProcessId == (uint)processId && IsWindowVisible(hWnd)) {
        found = hWnd;
        return false;
      }
      return true;
    }, System.IntPtr.Zero);
    return found;
  }
  public static void LeftClick() {
    mouse_event(0x0002, 0, 0, 0, System.UIntPtr.Zero);
    mouse_event(0x0004, 0, 0, 0, System.UIntPtr.Zero);
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

$main = Start-Process -FilePath $ExePath -ArgumentList @("-config", $ConfigPath) -PassThru
$control = Start-Process -FilePath $ControlPath -ArgumentList @("-config", $ConfigPath) -PassThru
try {
  $hwnd = [System.IntPtr]::Zero
  for ($i = 0; $i -lt 80; $i++) {
    Start-Sleep -Milliseconds 100
    $hwnd = [MofuMouseUiaSniffAssistQA]::FindVisibleWindow($control.Id)
    if ($hwnd -ne [System.IntPtr]::Zero) {
      break
    }
  }
  if ($hwnd -eq [System.IntPtr]::Zero) {
    throw "MofuMouse Control Center window was not found"
  }

  [MofuMouseUiaSniffAssistQA]::ShowWindow($hwnd, 5) | Out-Null
  [MofuMouseUiaSniffAssistQA]::SetForegroundWindow($hwnd) | Out-Null
  $rect = New-Object MofuMouseUiaSniffAssistQA+RECT
  [MofuMouseUiaSniffAssistQA]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
  [MofuMouseUiaSniffAssistQA]::SetCursorPos(($rect.Left + 90), ($rect.Top + 14)) | Out-Null
  [MofuMouseUiaSniffAssistQA]::LeftClick()
  Start-Sleep -Milliseconds 250

  $cursorX = $rect.Right - 72
  $cursorY = $rect.Top + 86
  [MofuMouseUiaSniffAssistQA]::SetCursorPos($cursorX, $cursorY) | Out-Null
  Start-Sleep -Milliseconds 3500

  Capture-Crop $OutputPath ($rect.Right - 650) ($rect.Top + 30) 640 300
}
finally {
  if ($control -and -not $control.HasExited) {
    Stop-Process -Id $control.Id -Force
  }
  if ($main -and -not $main.HasExited) {
    Stop-Process -Id $main.Id -Force
  }
}
