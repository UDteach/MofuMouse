param([Parameter(Mandatory = $true)][string]$Probe, [Parameter(Mandatory = $true)][string]$Output)
$ErrorActionPreference = 'Stop'
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class MofuNativeQA {
  [StructLayout(LayoutKind.Sequential)] public struct Point { public int X; public int Y; }
  [DllImport("user32.dll")] public static extern bool IsWindow(IntPtr hWnd);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll", EntryPoint="GetWindowLongW")] public static extern int GetWindowLong(IntPtr hWnd, int index);
  [DllImport("user32.dll")] public static extern IntPtr WindowFromPoint(Point p);
  [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint processId);
}
'@
$data = Get-Content -LiteralPath $Probe -Encoding UTF8 -Raw | ConvertFrom-Json
$window = [IntPtr]([long]$data.nativeHandle)
if (-not [MofuNativeQA]::IsWindow($window)) { throw 'The tested window no longer exists.' }
$style = [MofuNativeQA]::GetWindowLong($window, -20)
$point = New-Object MofuNativeQA+Point
$point.X = $data.point.x; $point.Y = $data.point.y
$under = [MofuNativeQA]::WindowFromPoint($point)
$foreground = [MofuNativeQA]::GetForegroundWindow()
[uint32]$underProcess = 0; [uint32]$foregroundProcess = 0; [uint32]$ownerProcess = 0
[void][MofuNativeQA]::GetWindowThreadProcessId($under, [ref]$underProcess)
[void][MofuNativeQA]::GetWindowThreadProcessId($foreground, [ref]$foregroundProcess)
[void][MofuNativeQA]::GetWindowThreadProcessId($window, [ref]$ownerProcess)
$checks = [ordered]@{
  ownedWindow = $ownerProcess -eq $data.processId
  visible = [MofuNativeQA]::IsWindowVisible($window)
  clickThroughStyle = ($style -band 0x20) -ne 0
  noActivateStyle = ($style -band 0x08000000) -ne 0
  underlyingWindowReceivesHit = $under -ne [IntPtr]::Zero -and $underProcess -ne $data.processId
  foregroundBelongsToOtherApp = $foregroundProcess -ne $data.processId
}
$result = [ordered]@{ checkedAt = (Get-Date).ToUniversalTime().ToString('o'); processId = $data.processId; nativeHandle = $data.nativeHandle; extendedStyle = ('0x{0:X8}' -f $style); point = $data.point; checks = $checks; pass = -not ($checks.Values -contains $false) }
$json = $result | ConvertTo-Json -Depth 6
[IO.File]::WriteAllText($Output, $json + "`n", [Text.UTF8Encoding]::new($false))
$json
if (-not $result.pass) { exit 1 }
