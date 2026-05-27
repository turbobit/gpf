# windows.ps1 — Install gpf (Greenfield Port Forwarding) for Windows
# Usage (PowerShell):
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/main/install/windows.ps1" -UseBasicParsing | Invoke-Expression
#   .\install\windows.ps1 v0.1.0

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"
$REPO = "turbobit/gpf"

# Strip leading 'v' if present
$versionClean = $Version -replace '^v', ''

$arch = "amd64"
if ([Environment]::Is64BitOperatingSystem) {
    $osArch = (Get-CimInstance -ClassName Win32_Processor -Property Architecture).Architecture
    switch ($osArch) {
        5 { $arch = "arm64" }   # ARM64
        9 { $arch = "amd64" }   # x64
        default { $arch = "amd64" }
    }
} else {
    Write-Host "Error: 32-bit Windows is not supported"
    exit 1
}

$binaryName = "gpf_windows_${arch}.exe"
$installDir = Join-Path $env:USERPROFILE ".gpf"
$installPath = Join-Path $installDir "gpf.exe"

Write-Host "Installing gpf ${Version} for Windows/${arch}..."

if ($versionClean -eq "latest") {
    $releasesUrl = "https://github.com/${REPO}/releases/latest/download/${binaryName}"
} else {
    $releasesUrl = "https://github.com/${REPO}/releases/download/v${versionClean}/${binaryName}"
}

$ProgressPreference = "SilentlyContinue"
try {
    Invoke-WebRequest -Uri $releasesUrl -OutFile $installPath -UseBasicParsing
} catch {
    Write-Host "Error: failed to download gpf ${Version} for Windows/${arch}"
    Write-Host $_.Exception.Message
    exit 1
}

if (-not (Test-Path $installPath)) {
    Write-Host "Error: download completed but file not found at $installPath"
    exit 1
}

Write-Host "Installed gpf to $installPath"

# Add to user PATH if not already present
$gpfPath = $installDir
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$gpfPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$gpfPath", "User")
    Write-Host "Added $gpfPath to your PATH (user scope)"
    Write-Host "You may need to restart your terminal for the PATH change to take effect."
}
