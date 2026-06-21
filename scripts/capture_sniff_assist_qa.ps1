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
using System.Runtime.InteropServices;
using System.Text;
public static class MofuMouseSniffAssistQA {
  public delegate bool EnumWindowsProc(System.IntPtr hWnd, System.IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, System.IntPtr lParam);
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern System.IntPtr FindWindow(string className, string windowName);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(System.IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool ShowWindow(System.IntPtr hWnd, int nCmdShow);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(System.IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(System.IntPtr hWnd, out uint processId);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetClassName(System.IntPtr hWnd, StringBuilder text, int count);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(System.IntPtr hWnd);
  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }
  public static System.IntPtr FindSettingsWindow(int processId) {
    System.IntPtr found = System.IntPtr.Zero;
    EnumWindows(delegate(System.IntPtr hWnd, System.IntPtr lParam) {
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
    }, System.IntPtr.Zero);
    return found;
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
    $hwnd = [MofuMouseSniffAssistQA]::FindSettingsWindow($process.Id)
    if ($hwnd -ne [System.IntPtr]::Zero) {
      break
    }
  }
  if ($hwnd -eq [System.IntPtr]::Zero) {
    throw "MofuMouse settings window was not found"
  }
  [MofuMouseSniffAssistQA]::ShowWindow($hwnd, 5) | Out-Null
  [MofuMouseSniffAssistQA]::SetForegroundWindow($hwnd) | Out-Null
  $rect = New-Object MofuMouseSniffAssistQA+RECT
  [MofuMouseSniffAssistQA]::GetWindowRect($hwnd, [ref]$rect) | Out-Null

  [MofuMouseSniffAssistQA]::SetCursorPos(($rect.Left + 340), ($rect.Top + 710)) | Out-Null
  Start-Sleep -Milliseconds 1600
  Capture-Crop $OutputPath ($rect.Left) ($rect.Top + 560) 520 220
}
finally {
  if ($process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
  }
}
