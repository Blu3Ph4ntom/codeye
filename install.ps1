# install.ps1 — one-line installer for codeye on Windows
# Usage: iex (irm https://raw.githubusercontent.com/blu3ph4ntom/codeye/main/install.ps1)

$ErrorActionPreference = 'Stop'

$Repo = "blu3ph4ntom/codeye"
$Binary = "codeye.exe"
$InstallDir = "$HOME\.codeye\bin"

function Get-LatestVersion {
    $Uri = "https://api.github.com/repos/$Repo/releases/latest"
    $Release = Invoke-RestMethod -Uri $Uri -UseBasicParsing
    return $Release.tag_name
}

function Install-Codeye {
    $Version = Get-LatestVersion
    $Arch = "amd64" # Default for Windows unless detecting ARM64
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $Arch = "arm64" }
    
    $ZipFile = "codeye_$($Version.TrimStart('v'))_windows_$Arch.zip"
    $Url = "https://github.com/$Repo/releases/download/$Version/$ZipFile"
    $TmpDir = Join-Path $env:TEMP "codeye-install"
    
    if (-not (Test-Path $TmpDir)) { New-Item -ItemType Directory -Path $TmpDir }
    $DestPath = Join-Path $TmpDir $ZipFile
    
    Write-Host "→ Downloading $Url" -ForegroundColor Cyan
    Invoke-WebRequest -Uri $Url -OutFile $DestPath -UseBasicParsing
    
    Write-Host "→ Extracting..." -ForegroundColor Cyan
    Expand-Archive -Path $DestPath -DestinationPath $TmpDir -Force
    
    if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir -Force }
    Copy-Item (Join-Path $TmpDir "codeye.exe") (Join-Path $InstallDir $Binary) -Force
    
    # Add to path if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        Write-Host "✓ Added $InstallDir to user PATH. Please restart your terminal." -ForegroundColor Green
    }
    
    Write-Host "✓ codeye $Version installed successfully to $InstallDir" -ForegroundColor Green
}

Install-Codeye
