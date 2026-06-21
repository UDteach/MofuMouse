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
public static class MofuMouseFringeQAMouse {
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
}
"@

Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName System.Windows.Forms

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

$resolvedExe = Resolve-Path -LiteralPath $ExePath
$resolvedConfig = Resolve-Path -LiteralPath $ConfigPath
$outFile = [System.IO.Path]::GetFullPath($OutputPath)
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outFile) | Out-Null

$form = New-Object System.Windows.Forms.Form
$form.Text = "MofuMouse Fringe QA"
$form.StartPosition = [System.Windows.Forms.FormStartPosition]::Manual
$form.Location = [System.Drawing.Point]::new(430, 220)
$form.Size = [System.Drawing.Size]::new(360, 260)
$form.BackColor = [System.Drawing.Color]::White
$form.FormBorderStyle = [System.Windows.Forms.FormBorderStyle]::None
$form.ShowInTaskbar = $false

$process = $null
try {
  $form.Show()
  [System.Windows.Forms.Application]::DoEvents()
  Start-Sleep -Milliseconds 300

  $process = Start-Process -FilePath $resolvedExe -ArgumentList @("-config", $resolvedConfig) -PassThru
  Start-Sleep -Milliseconds 900
  [MofuMouseFringeQAMouse]::SetCursorPos(610, 350) | Out-Null
  Start-Sleep -Milliseconds 1400
  [System.Windows.Forms.Application]::DoEvents()
  Capture-Crop $outFile 430 220 360 260
}
finally {
  if ($process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
  }
  if ($form) {
    $form.Close()
    $form.Dispose()
  }
}
