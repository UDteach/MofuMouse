param(
  [Parameter(Mandatory = $true)]
  [string]$AppPath,

  [int]$MinimumDpiAwareness = 2
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;

public static class MofuMouseDisplayQA {
  public delegate bool MonitorEnumProc(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT lprcMonitor, IntPtr dwData);

  [DllImport("user32.dll")] public static extern int GetSystemMetrics(int nIndex);
  [DllImport("user32.dll")] public static extern bool EnumDisplayMonitors(IntPtr hdc, IntPtr clip, MonitorEnumProc callback, IntPtr data);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern bool GetMonitorInfo(IntPtr hMonitor, ref MONITORINFOEX info);
  [DllImport("kernel32.dll")] public static extern IntPtr OpenProcess(UInt32 desiredAccess, bool inheritHandle, UInt32 processId);
  [DllImport("kernel32.dll")] public static extern bool CloseHandle(IntPtr handle);
  [DllImport("shcore.dll")] public static extern int GetProcessDpiAwareness(IntPtr processHandle, out int awareness);
  [DllImport("shcore.dll")] public static extern int GetDpiForMonitor(IntPtr hMonitor, int dpiType, out UInt32 dpiX, out UInt32 dpiY);

  public const int SM_XVIRTUALSCREEN = 76;
  public const int SM_YVIRTUALSCREEN = 77;
  public const int SM_CXVIRTUALSCREEN = 78;
  public const int SM_CYVIRTUALSCREEN = 79;
  public const UInt32 PROCESS_QUERY_LIMITED_INFORMATION = 0x1000;

  [StructLayout(LayoutKind.Sequential)]
  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }

  [StructLayout(LayoutKind.Sequential, CharSet=CharSet.Unicode)]
  public struct MONITORINFOEX {
    public int cbSize;
    public RECT rcMonitor;
    public RECT rcWork;
    public uint dwFlags;
    [MarshalAs(UnmanagedType.ByValTStr, SizeConst=32)]
    public string szDevice;
  }

  public static List<MONITORINFOEX> Monitors() {
    List<MONITORINFOEX> monitors = new List<MONITORINFOEX>();
    EnumDisplayMonitors(IntPtr.Zero, IntPtr.Zero, delegate(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT rect, IntPtr data) {
      MONITORINFOEX info = new MONITORINFOEX();
      info.cbSize = Marshal.SizeOf(typeof(MONITORINFOEX));
      if (GetMonitorInfo(hMonitor, ref info)) {
        monitors.Add(info);
      }
      return true;
    }, IntPtr.Zero);
    return monitors;
  }

  public static List<int> MonitorDpis() {
    List<int> dpis = new List<int>();
    EnumDisplayMonitors(IntPtr.Zero, IntPtr.Zero, delegate(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT rect, IntPtr data) {
      UInt32 dpiX;
      UInt32 dpiY;
      int hr = GetDpiForMonitor(hMonitor, 0, out dpiX, out dpiY);
      dpis.Add(hr == 0 ? (int)dpiX : 0);
      return true;
    }, IntPtr.Zero);
    return dpis;
  }
}
"@

function Resolve-RequiredPath([string]$Path) {
  $resolved = Resolve-Path -LiteralPath $Path -ErrorAction Stop
  return $resolved.Path
}

$exe = Resolve-RequiredPath $AppPath

$monitors = [MofuMouseDisplayQA]::Monitors()
if ($monitors.Count -eq 0) {
  throw "No display monitors were reported by EnumDisplayMonitors"
}
$monitorDpis = [MofuMouseDisplayQA]::MonitorDpis()

$lefts = @()
$tops = @()
$rights = @()
$bottoms = @()
foreach ($monitor in $monitors) {
  $lefts += $monitor.rcMonitor.Left
  $tops += $monitor.rcMonitor.Top
  $rights += $monitor.rcMonitor.Right
  $bottoms += $monitor.rcMonitor.Bottom
}

$unionLeft = [int](($lefts | Measure-Object -Minimum).Minimum)
$unionTop = [int](($tops | Measure-Object -Minimum).Minimum)
$unionRight = [int](($rights | Measure-Object -Maximum).Maximum)
$unionBottom = [int](($bottoms | Measure-Object -Maximum).Maximum)
$unionWidth = $unionRight - $unionLeft
$unionHeight = $unionBottom - $unionTop
$maxMonitorDpi = [int](($monitorDpis | Measure-Object -Maximum).Maximum)

$virtualLeft = [MofuMouseDisplayQA]::GetSystemMetrics([MofuMouseDisplayQA]::SM_XVIRTUALSCREEN)
$virtualTop = [MofuMouseDisplayQA]::GetSystemMetrics([MofuMouseDisplayQA]::SM_YVIRTUALSCREEN)
$virtualWidth = [MofuMouseDisplayQA]::GetSystemMetrics([MofuMouseDisplayQA]::SM_CXVIRTUALSCREEN)
$virtualHeight = [MofuMouseDisplayQA]::GetSystemMetrics([MofuMouseDisplayQA]::SM_CYVIRTUALSCREEN)

if ($virtualLeft -ne $unionLeft -or $virtualTop -ne $unionTop -or $virtualWidth -ne $unionWidth -or $virtualHeight -ne $unionHeight) {
  throw "Virtual screen mismatch. metrics=$virtualLeft,$virtualTop,$virtualWidth,$virtualHeight union=$unionLeft,$unionTop,$unionWidth,$unionHeight"
}

$app = $null
$handle = [IntPtr]::Zero
try {
  $app = Start-Process -FilePath $exe -ArgumentList @("--smoke") -PassThru
  Start-Sleep -Milliseconds 700
  if ($app.HasExited) {
    throw "App exited before DPI awareness could be inspected"
  }

  $handle = [MofuMouseDisplayQA]::OpenProcess([MofuMouseDisplayQA]::PROCESS_QUERY_LIMITED_INFORMATION, $false, [uint32]$app.Id)
  if ($handle -eq [IntPtr]::Zero) {
    throw "OpenProcess failed for PID $($app.Id)"
  }

  $awareness = 0
  $hr = [MofuMouseDisplayQA]::GetProcessDpiAwareness($handle, [ref]$awareness)
  if ($hr -ne 0) {
    throw ("GetProcessDpiAwareness failed: 0x{0:X8}" -f $hr)
  }
  if ($awareness -lt $MinimumDpiAwareness) {
    throw "DPI awareness $awareness is below required $MinimumDpiAwareness"
  }

  [pscustomobject]@{
    App = $exe
    DpiAwareness = $awareness
    MinimumDpiAwareness = $MinimumDpiAwareness
    MonitorCount = $monitors.Count
    MonitorDpiX = ($monitorDpis -join ",")
    MaxMonitorDpi = $maxMonitorDpi
    VirtualScreen = "$virtualLeft,$virtualTop,$virtualWidth,$virtualHeight"
    MonitorUnion = "$unionLeft,$unionTop,$unionWidth,$unionHeight"
    PhysicalHighDpiCovered = ($maxMonitorDpi -gt 96)
    PhysicalMultiMonitorCovered = ($monitors.Count -gt 1)
  }
}
finally {
  if ($handle -ne [IntPtr]::Zero) {
    [MofuMouseDisplayQA]::CloseHandle($handle) | Out-Null
  }
  if ($app -and -not $app.HasExited) {
    Stop-Process -Id $app.Id -Force
  }
}
