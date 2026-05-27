# windows.ps1 — Install gpf (Greenfield Port Forwarding) for Windows
# Usage (PowerShell):
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/user/port-forwarding/main/install/windows.ps1" -UseBasicParsing | Invoke-Expression
#   .\install\windows.ps1 v0.1.0

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"
$REPO = "user/port-forwarding"

if (-not $Version) {
    Write-Host "Usage: .\windows.ps1 <version>"
    Write-Host "Example: .\windows.ps1 v0.1.0"
    exit 1
}

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

$binaryName = "gpf_windows_${arch}"
$installDir = Join-Path $env:USERPROFILE ".gpf"
$installPath = Join-Path $installDir "gpf.exe"

Write-Host "Installing gpf $Version for Windows/$arch..."

$releasesUrl = "https://github.com/${REPO}/releases/download/v${versionClean}/${binaryName}"

$ProgressPreference = "SilentlyContinue"
Invoke-WebRequest -Uri $releasesUrl -OutFile $installPath -UseBasicParsing

Write-Host "Installed gpf to $installPath"

# Add to user PATH if not already present
$gpfPath = Join-Path $installDir
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$gpfPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$gpfPath", "User")
    Write-Host "Added $gpfPath to your PATH (user scope)"
    Write-Host "You may need to restart your terminal for the PATH change to take effect."
}
