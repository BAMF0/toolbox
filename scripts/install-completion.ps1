# ToolBox Autocompletion Installation Script for PowerShell

Write-Host "========================================"
Write-Host "ToolBox Autocompletion Installer"
Write-Host "========================================"
Write-Host ""

# Check if tb is installed
if (!(Get-Command tb -ErrorAction SilentlyContinue)) {
    Write-Host "Error: 'tb' command not found in PATH" -ForegroundColor Red
    Write-Host "Please install ToolBox first"
    exit 1
}

Write-Host "Installing PowerShell completion..." -ForegroundColor Green
Write-Host ""

# Generate completion script
$completionScript = & tb completion powershell

# Check if profile exists
if (!(Test-Path $PROFILE)) {
    New-Item -Path $PROFILE -ItemType File -Force | Out-Null
    Write-Host "Created PowerShell profile: $PROFILE"
}

# Check if already installed
$profileContent = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue
if ($profileContent -and $profileContent.Contains("# ToolBox completion")) {
    Write-Host "ToolBox completion already installed in profile" -ForegroundColor Yellow
    Write-Host ""
    
    $response = Read-Host "Do you want to update it? (y/N)"
    if ($response -ne 'y' -and $response -ne 'Y') {
        Write-Host "Installation cancelled"
        exit 0
    }
    
    # Remove old completion
    $lines = Get-Content $PROFILE
    $newLines = @()
    $skip = $false
    
    foreach ($line in $lines) {
        if ($line -match "# ToolBox completion") {
            $skip = $true
        }
        elseif ($line -match "# End ToolBox completion") {
            $skip = $false
            continue
        }
        
        if (!$skip) {
            $newLines += $line
        }
    }
    
    $newLines | Set-Content $PROFILE
}

# Add completion to profile
Add-Content -Path $PROFILE -Value "`n# ToolBox completion"
Add-Content -Path $PROFILE -Value $completionScript
Add-Content -Path $PROFILE -Value "# End ToolBox completion`n"

Write-Host "✓ Installed to: $PROFILE" -ForegroundColor Green
Write-Host ""
Write-Host "========================================"
Write-Host "Installation Complete!"
Write-Host "========================================"
Write-Host ""
Write-Host "Reload your profile to activate:"
Write-Host "  . `$PROFILE" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test completion by typing:"
Write-Host "  tb bui<TAB>" -ForegroundColor Cyan
Write-Host ""
