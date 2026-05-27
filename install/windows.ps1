# windows.ps1 — Install gpf (Greenfield Port Forwarding) for Windows
# Usage (PowerShell):
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/master/install/windows.ps1" -UseBasicParsing | Invoke-Expression
#   .\install\windows.ps1 v0.1.0

param(
    [string]$Version = "latest"
)

$REPO = "turbobit/gpf"
$ProgressPreference = "SilentlyContinue"

# Detect architecture from environment variables
$arch = "amd64"
$procArch = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE", "Process")
$procArchW64 = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITEW6432", "Process")
if ($procArch -eq "ARM64" -or $procArchW64 -eq "ARM64") {
    $arch = "arm64"
} elseif ($procArch -eq "AMD64" -or $procArchW64 -eq "AMD64") {
    $arch = "amd64"
} elseif ($procArch -eq "x86") {
    if ($procArchW64) { $arch = "amd64" } else { Write-Host "Error: 32-bit Windows is not supported"; exit 1 }
}

$installDir = Join-Path $env:USERPROFILE ".gpf"
$installPath = Join-Path $installDir "gpf.exe"

try {
    $binaryName = "gpf_windows_${arch}.exe"

    $downloadUrl = "https://github.com/${REPO}/releases/latest/download/${binaryName}"
    if ($Version -ne "latest") {
        $versionClean = $Version -replace '^v', ''
        $downloadUrl = "https://github.com/${REPO}/releases/download/v${versionClean}/${binaryName}"
    }

    Write-Host "Installing gpf ${Version} for Windows/${arch}..."
    Write-Host "Downloading from: $downloadUrl"

    Invoke-WebRequest -Uri $downloadUrl -OutFile $installPath -UseBasicParsing

    if (-not (Test-Path $installPath)) {
        Write-Host "Error: download completed but file not found at $installPath"
        exit 1
    }

    $fileSize = (Get-Item $installPath).Length
    Write-Host "Installed: $installPath ($fileSize bytes)"

    # Add to user PATH if not already present
    $gpfPath = $installDir
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$gpfPath*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$gpfPath", "User")
        Write-Host "Added $gpfPath to your PATH (user scope)"
        Write-Host "Note: open a new terminal for PATH to take effect."
    } else {
        Write-Host "$gpfPath is already in your PATH."
    }

    # Verify installation
    Write-Host ""
    try {
        $verOutput = & $installPath --version 2>&1
        Write-Host "gpf version: $verOutput"
    } catch {
        Write-Host "Warning: could not verify installation."
    }

    Write-Host ""
    Write-Host "==============================="
    Write-Host "Installation complete!"
    Write-Host "Binary: $installPath"
    Write-Host "==============================="
    Write-Host ""
    Write-Host "Open a new terminal and run:"
    Write-Host "  gpf --help"
    Write-Host ""
    Start-Sleep -Seconds 10
} catch {
    Write-Host ""
    Write-Host "==============================="
    Write-Host "ERROR: Installation failed!"
    Write-Host "==============================="
    Write-Host ""
    Write-Host "Error: $($_.Exception.Message)"
    Write-Host "Line: $($_.InvocationInfo.ScriptLineNumber)"
    Write-Host ""
    Start-Sleep -Seconds 10
}
