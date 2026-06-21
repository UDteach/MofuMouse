param(
  [Parameter(Mandatory = $true)]
  [string]$ExePath,

  [Parameter(Mandatory = $true)]
  [string]$OutPath
)

$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.Drawing

$resolvedExe = Resolve-Path -LiteralPath $ExePath
$outFile = [System.IO.Path]::GetFullPath($OutPath)
$outDir = Split-Path -Parent $outFile
if ($outDir) {
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

$icon = [System.Drawing.Icon]::ExtractAssociatedIcon($resolvedExe)
if ($null -eq $icon) {
  throw "No associated icon found: $resolvedExe"
}

try {
  $bitmap = $icon.ToBitmap()
  try {
    $bitmap.Save($outFile, [System.Drawing.Imaging.ImageFormat]::Png)
  }
  finally {
    $bitmap.Dispose()
  }
}
finally {
  $icon.Dispose()
}

$image = [System.Drawing.Image]::FromFile($outFile)
try {
  if ($image.Width -le 1 -or $image.Height -le 1) {
    throw "Extracted icon is too small: $($image.Width)x$($image.Height)"
  }
  [pscustomobject]@{
    Exe = $resolvedExe.Path
    Output = $outFile
    Width = $image.Width
    Height = $image.Height
  }
}
finally {
  $image.Dispose()
}
