param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$ConfigPath,

  [Parameter(Mandatory = $true)]
  [string]$OutputDir
)

$ErrorActionPreference = "Stop"

Add-Type @"
using System.Runtime.InteropServices;
public static class MofuMouseQAMouse {
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
}
"@

Add-Type -AssemblyName System.Drawing

function Capture-Crop([string]$Path, [int]$X, [int]$Y, [int]$W, [int]$H) {
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

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$process = Start-Process -FilePath $ExePath -ArgumentList @("-config", $ConfigPath) -PassThru
try {
  Start-Sleep -Milliseconds 900
  [MofuMouseQAMouse]::SetCursorPos(500, 300) | Out-Null
  Start-Sleep -Milliseconds 1400
  Capture-Crop (Join-Path $OutputDir "overlay-small-near-idle.png") 480 260 120 100

  [MofuMouseQAMouse]::SetCursorPos(540, 310) | Out-Null
  Start-Sleep -Milliseconds 80
  [MofuMouseQAMouse]::SetCursorPos(585, 325) | Out-Null
  Start-Sleep -Milliseconds 80
  [MofuMouseQAMouse]::SetCursorPos(630, 340) | Out-Null
  Start-Sleep -Milliseconds 80
  Capture-Crop (Join-Path $OutputDir "overlay-small-near-moving.png") 525 270 190 130
}
finally {
  if ($process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
  }
}
