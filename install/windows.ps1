# windows.ps1 — Install gpf (Greenfield Port Forwarding) for Windows
# Usage (PowerShell):
#   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/master/install/windows.ps1" -UseBasicParsing | Invoke-Expression
#   .\install\windows.ps1 v0.1.0

param(
    [string]$Version = "latest"
)

$REPO = "turbobit/gpf"
$ProgressPreference = "SilentlyContinue"

# Strip leading 'v' if present
$versionClean = $Version -replace '^v', ''

# Detect architecture from environment variables (no CIM/WMI needed)
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

$binaryName = "gpf_windows_${arch}.exe"
$installDir = Join-Path $env:USERPROFILE ".gpf"
$installPath = Join-Path $installDir "gpf.exe"

try {
    Write-Host "Installing gpf ${Version} for Windows/${arch}..."

    if ($versionClean -eq "latest") {
        $releasesUrl = "https://github.com/${REPO}/releases/latest/download/${binaryName}"
    } else {
        $releasesUrl = "https://github.com/${REPO}/releases/download/v${versionClean}/${binaryName}"
    }

    Write-Host "Downloading from: $releasesUrl"
    Invoke-WebRequest -Uri $releasesUrl -OutFile $installPath -UseBasicParsing

    if (-not (Test-Path $installPath)) {
        Write-Host "Error: download completed but file not found at $installPath"
        exit 1
    }

    $fileSize = (Get-Item $installPath).Length
    Write-Host "Downloaded: $fileSize bytes"

    # Add to user PATH if not already present
    $gpfPath = $installDir
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$gpfPath*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$gpfPath", "User")
        Write-Host "Added $gpfPath to your PATH (user scope)"
        Write-Host "Note: you need to open a new terminal for PATH to take effect."
    } else {
        Write-Host "$gpfPath is already in your PATH."
    }

    # Test the installation by running gpf.exe with --version
    Write-Host ""
    try {
        $verOutput = & $installPath --version 2>&1
        Write-Host "gpf version: $verOutput"
    } catch {
        Write-Host "Warning: could not verify installation. Error: $_"
    }

    Write-Host ""
    Write-Host "==============================="
    Write-Host "Installation complete!"
    Write-Host "Binary: $installPath"
    Write-Host "Version: $versionClean"
    Write-Host ""
    Write-Host "To use gpf in a new terminal:"
    Write-Host "  gpf --help"
    Write-Host ""
    Write-Host "Or run directly now:"
    Write-Host "  $installPath --help"
    Write-Host "==============================="
    Write-Host ""
    Write-Host "Press any key to exit..."
    Start-Sleep -Seconds 10
} catch {
    Write-Host ""
    Write-Host "==============================="
    Write-Host "ERROR: Installation failed!"
    Write-Host "==============================="
    Write-Host ""
    Write-Host "Error: $($_.Exception.Message)"
    Write-Host ""
    Write-Host "At line: $($_.InvocationInfo.ScriptLineNumber)"
    Write-Host "Position: $($_.InvocationInfo.PositionMessage)"
    Write-Host ""
    Write-Host "Press any key to exit..."
    Start-Sleep -Seconds 10
}
