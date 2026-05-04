# Install DNT-Vault on Windows (PowerShell)

$VERSION = "1.0.0"
$REPO = "dungnt1312/dnt-vault"
$INSTALL_DIR = "$env:USERPROFILE\dnt-vault"

Write-Host "╔═══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                               ║" -ForegroundColor Cyan
Write-Host "║              DNT-Vault SSH Config Sync Installer              ║" -ForegroundColor Cyan
Write-Host "║                                                               ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$ARCH = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

Write-Host "Detected Architecture: $ARCH" -ForegroundColor Cyan
Write-Host ""

# Create install directory
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
}

# Download URLs
$SERVER_URL = "https://github.com/$REPO/releases/download/v$VERSION/dnt-vault-server-windows-$ARCH.exe"
$CLI_URL = "https://github.com/$REPO/releases/download/v$VERSION/dnt-vault-windows-$ARCH.exe"

Write-Host "Downloading binaries..." -ForegroundColor Cyan
Write-Host ""

# Download server
Write-Host "Downloading server..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $SERVER_URL -OutFile "$INSTALL_DIR\dnt-vault-server.exe"
    Write-Host "✓ Server downloaded" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to download server: $_" -ForegroundColor Red
    exit 1
}

# Download CLI
Write-Host "Downloading CLI..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $CLI_URL -OutFile "$INSTALL_DIR\dnt-vault.exe"
    Write-Host "✓ CLI downloaded" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to download CLI: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✓ Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Installed to: $INSTALL_DIR" -ForegroundColor Cyan
Write-Host ""

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$INSTALL_DIR*") {
    Write-Host "Adding to PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$INSTALL_DIR", "User")
    Write-Host "✓ Added to PATH (restart terminal to take effect)" -ForegroundColor Green
    Write-Host ""
    Write-Host "For current session, run:" -ForegroundColor Yellow
    Write-Host "  `$env:Path += `";$INSTALL_DIR`"" -ForegroundColor White
} else {
    Write-Host "Already in PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  dnt-vault-server.exe    # Start server" -ForegroundColor White
Write-Host "  dnt-vault.exe init       # Initialize client" -ForegroundColor White
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Cyan
Write-Host "  1. Start server: dnt-vault-server.exe" -ForegroundColor White
Write-Host "  2. Init client:  dnt-vault.exe init" -ForegroundColor White
Write-Host "  3. Login:        dnt-vault.exe login" -ForegroundColor White
Write-Host "  4. Push config:  dnt-vault.exe push" -ForegroundColor White
Write-Host ""
Write-Host "Documentation: https://github.com/$REPO" -ForegroundColor Cyan
