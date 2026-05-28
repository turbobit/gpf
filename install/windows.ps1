#!/usr/bin/env pwsh

# windows.ps1 — Install gpf (Greenfield Port Forwarding) for Windows
# Inspired by ggh (https://github.com/byawitz/ggh)
# Usage (PowerShell):
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/master/install/windows.ps1" -UseBasicParsing | Invoke-Expression
#   .\install\windows.ps1 v0.1.0

$GPFInstallDir = "$env:USERPROFILE\.gpf"
$GPFCliName = "gpf.exe"
$GPFCliPath = "${GPFInstallDir}\${GPFCliName}"

[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"

$REPO = "turbobit/gpf"
$ProgressPreference = "SilentlyContinue"

# Detect architecture
$arch = "amd64"
$procArch = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE", "Process")
$procArchW64 = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITEW6432", "Process")
if ($procArch -eq "ARM64" -or $procArchW64 -eq "ARM64") {
    $arch = "arm64"
} elseif ($procArch -eq "AMD64" -or $procArchW64 -eq "AMD64") {
    $arch = "amd64"
} elseif ($procArch -eq "x86") {
    if ($procArchW64) { $arch = "amd64" } else {
        Write-Host "Error: 32-bit Windows is not supported"
        return 1
    }
}

# Version from argument or default to latest
$Version = if ($args[0]) { $args[0] } else { "latest" }
$versionClean = $Version -replace '^v', ''

$binaryName = "gpf_windows_${arch}.exe"
$downloadUrl = "https://github.com/${REPO}/releases/latest/download/${binaryName}"
if ($Version -ne "latest") {
    $downloadUrl = "https://github.com/${REPO}/releases/download/v${versionClean}/${binaryName}"
}

Write-Host "Installing gpf ${Version} for Windows/${arch}" -ForegroundColor DarkCyan
Write-Host "Creating the directory in $GPFInstallDir" -ForegroundColor Green
New-Item -ErrorAction Ignore -Path $GPFInstallDir -ItemType "directory"
if (!(Test-Path $GPFInstallDir -PathType Container)) {
    Write-Host "Error: Could not create $GPFInstallDir" -ForegroundColor Red
    return 1
}

Write-Host "Downloading $binaryName ..." -ForegroundColor DarkCyan
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $GPFCliPath -UseBasicParsing
} catch {
    Write-Host "Error: Failed to download gpf" -ForegroundColor Red
    Write-Host $_.Exception.Message
    return 1
}

if (!(Test-Path $GPFCliPath -PathType Leaf)) {
    Write-Host "Error: Failed to download gpf" -ForegroundColor Red
    return 1
}

$fileSize = (Get-Item $GPFCliPath).Length
Write-Host "Downloaded: $fileSize bytes" -ForegroundColor Green

# Verify installation
Write-Host ""
try {
    $verOutput = & $GPFCliPath --version 2>&1
    Write-Host "gpf version: $verOutput" -ForegroundColor Green
} catch {
    Write-Host "Warning: could not verify installation." -ForegroundColor Yellow
}

# Add to PATH
Write-Host "Attempting to add $GPFInstallDir to User Path Environment variable..."
$UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPathEnvironmentVar -like "*$GPFInstallDir*") {
    Write-Host "gpf already in the path, skipping..." -ForegroundColor Cyan
} else {
    [System.Environment]::SetEnvironmentVariable("PATH", $UserPathEnvironmentVar + ";$GPFInstallDir", "User")
    $UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Host "Added $GPFInstallDir to User Path" -ForegroundColor Green
}

Write-Host ""
Write-Host "gpf was installed successfully to $GPFInstallDir" -ForegroundColor Green
Write-Host ""
Write-Host "Restart the terminal and run:"
Write-Host "  gpf --help"
Write-Host ""
Start-Sleep -Seconds 10
