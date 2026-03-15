& {
param(
    [Parameter(Position=0)]
    [ValidatePattern('^(latest|v?\d+\.\d+\.\d+(-[^\s]+)?)$')]
    [string]$Target = "latest"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$ProgressPreference = 'SilentlyContinue'

if (-not [Environment]::Is64BitProcess) {
    Write-Error "portui does not support 32-bit Windows. Please use 64-bit PowerShell."
    exit 1
}

$Repo = "edereagzi/portui"
$BinaryName = "portui.exe"
$InstallDir = if ($env:PORTUI_INSTALL_DIR) { $env:PORTUI_INSTALL_DIR } else { "$env:USERPROFILE\bin" }
$VerifyChecksum = if ($env:PORTUI_VERIFY_CHECKSUM) { $env:PORTUI_VERIFY_CHECKSUM } else { "1" }

if ($VerifyChecksum -notin @("0", "1")) {
    Write-Error "PORTUI_VERIFY_CHECKSUM must be 0 or 1"
    exit 1
}

if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
    $arch = "arm64"
}
else {
    $arch = "amd64"
}

if ($Target -eq "latest") {
    $baseUrl = "https://github.com/$Repo/releases/latest/download"
}
else {
    $normalized = if ($Target.StartsWith("v")) { $Target } else { "v$Target" }
    $baseUrl = "https://github.com/$Repo/releases/download/$normalized"
}

$asset = "portui_windows_$arch.zip"
$assetUrl = "$baseUrl/$asset"
$checksumsUrl = "$baseUrl/checksums.txt"

function Get-ExpectedChecksum([string]$checksumsPath, [string]$assetName) {
    $line = Get-Content -Path $checksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($assetName))$" } | Select-Object -First 1
    if (-not $line) {
        throw "checksum entry not found for $assetName"
    }
    return ($line -split '\s+')[0]
}

function Normalize-PathEntry([string]$value) {
    if ([string]::IsNullOrWhiteSpace($value)) {
        return $null
    }

    $expanded = [Environment]::ExpandEnvironmentVariables($value.Trim().Trim('"'))
    $trimmed = $expanded.TrimEnd('\\')
    if ([string]::IsNullOrWhiteSpace($trimmed)) {
        return $null
    }

    return $trimmed.ToLowerInvariant()
}

function Ensure-PathContains([string]$targetDir) {
    $target = Normalize-PathEntry $targetDir
    if (-not $target) {
        throw "invalid install directory for PATH update"
    }

    $result = @{
        UserPathUpdated = $false
        ProcessPathUpdated = $false
    }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $userEntries = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $userEntries = $userPath -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    }

    $userNormalized = $userEntries | ForEach-Object { Normalize-PathEntry $_ } | Where-Object { $_ }
    if ($userNormalized -notcontains $target) {
        $newUserPath = if ([string]::IsNullOrWhiteSpace($userPath)) {
            $targetDir
        }
        else {
            "$userPath;$targetDir"
        }

        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
        $result.UserPathUpdated = $true
    }

    $processPath = $env:PATH
    $processEntries = @()
    if (-not [string]::IsNullOrWhiteSpace($processPath)) {
        $processEntries = $processPath -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    }

    $processNormalized = $processEntries | ForEach-Object { Normalize-PathEntry $_ } | Where-Object { $_ }
    if ($processNormalized -notcontains $target) {
        $env:PATH = if ([string]::IsNullOrWhiteSpace($processPath)) {
            $targetDir
        }
        else {
            "$targetDir;$processPath"
        }
        $result.ProcessPathUpdated = $true
    }

    return $result
}

$tmpDir = Join-Path $env:TEMP ("portui-install-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $asset
    $checksumsPath = Join-Path $tmpDir "checksums.txt"

    Write-Output "Downloading $asset..."
    Invoke-WebRequest -Uri $assetUrl -OutFile $archivePath -ErrorAction Stop

    if ($VerifyChecksum -eq "1") {
        Write-Output "Verifying checksum..."
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -ErrorAction Stop

        $expected = Get-ExpectedChecksum -checksumsPath $checksumsPath -assetName $asset
        $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($expected.ToLowerInvariant() -ne $actual) {
            throw "checksum mismatch for $asset"
        }
    }

    Write-Output "Extracting archive..."
    Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

    $sourceBin = Join-Path $tmpDir $BinaryName
    if (-not (Test-Path $sourceBin)) {
        throw "$BinaryName not found in archive"
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    }

    $destBin = Join-Path $InstallDir $BinaryName
    Copy-Item -Force -Path $sourceBin -Destination $destBin

    $pathUpdate = Ensure-PathContains -targetDir $InstallDir
    if ($pathUpdate.UserPathUpdated) {
        Write-Output "Added $InstallDir to user PATH."
    }
    if ($pathUpdate.ProcessPathUpdated) {
        Write-Output "Updated current session PATH."
    }

    Write-Output "Installed: $destBin"
    & $destBin -v
}
finally {
    try {
        Start-Sleep -Seconds 1
        if (Test-Path $tmpDir) {
            Remove-Item -Recurse -Force $tmpDir
        }
    }
    catch {
        Write-Warning "Could not remove temporary files: $tmpDir"
    }
}
}
