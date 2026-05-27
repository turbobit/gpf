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
    Write-Host "Installing gpf ${Version} for Windows/${arch}..."

    $baseDownloadUrl = "https://github.com/${REPO}/releases"
    if ($versionClean -eq "latest") {
        $baseDownloadUrl += "/latest"
    } else {
        $baseDownloadUrl += "/download/v${versionClean}"
    }

    # Try bare binary first, fall back to tar.gz
    $binaryName = "gpf_windows_${arch}.exe"
    $tarName = "gpf_windows_${arch}.tar.gz"
    $binaryUrl = "${baseDownloadUrl}/download/${binaryName}"
    $tarUrl = "${baseDownloadUrl}/download/${tarName}"

    Write-Host "Downloading..."
    $tmpTar = Join-Path $env:TEMP "gpf_install.tar.gz"

    try {
        Invoke-WebRequest -Uri $binaryUrl -OutFile $installPath -UseBasicParsing
        Write-Host "Downloaded binary directly."
    } catch {
        # Bare binary not found, try tar.gz
        try {
            Invoke-WebRequest -Uri $tarUrl -OutFile $tmpTar -UseBasicParsing
            Write-Host "Downloaded tar.gz, extracting..."
            tar -xzf $tmpTar -C (Split-Path $installPath) 2>$null
            # The extracted file may have a different name, find it
            $extracted = Get-ChildItem (Split-Path $installPath) -Filter "*.exe" | Select-Object -First 1
            if ($extracted) {
                if ($extracted.FullName -ne $installPath) {
                    Move-Item $extracted.FullName $installPath -Force
                }
            }
            Remove-Item $tmpTar -ErrorAction SilentlyContinue
            Write-Host "Extracted binary."
        } catch {
            Write-Host "Error: failed to download gpf for Windows/${arch}"
            Write-Host $_.Exception.Message
            exit 1
        }
    }

    if (-not (Test-Path $installPath)) {
        Write-Host "Error: binary not found at $installPath"
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
