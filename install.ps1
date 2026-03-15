$ErrorActionPreference = "Stop"

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$HOME\bin",
    [ValidateSet("0", "1")]
    [string]$VerifyChecksum = "1"
)

$Repo = "edereagzi/portui"
$BinaryName = "portui.exe"

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
    switch ($arch) {
        "x64" { return "amd64" }
        "arm64" { return "arm64" }
        default { throw "unsupported architecture: $arch (supported: amd64, arm64)" }
    }
}

function Resolve-BaseUrl([string]$v) {
    if ($v -eq "latest") {
        return "https://github.com/$Repo/releases/latest/download"
    }

    $normalized = if ($v.StartsWith("v")) { $v } else { "v$v" }
    return "https://github.com/$Repo/releases/download/$normalized"
}

function Get-ExpectedChecksum([string]$checksumsPath, [string]$asset) {
    $line = Get-Content -Path $checksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($asset))$" } | Select-Object -First 1
    if (-not $line) {
        throw "checksum entry not found for $asset"
    }
    return ($line -split '\s+')[0]
}

$arch = Get-Arch
$baseUrl = Resolve-BaseUrl -v $Version
$asset = "portui_windows_$arch.zip"
$assetUrl = "$baseUrl/$asset"
$checksumsUrl = "$baseUrl/checksums.txt"

$tmpDir = Join-Path $env:TEMP ("portui-install-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $asset
    $checksumsPath = Join-Path $tmpDir "checksums.txt"

    Write-Host "Downloading $asset..."
    Invoke-WebRequest -Uri $assetUrl -OutFile $archivePath

    if ($VerifyChecksum -eq "1") {
        Write-Host "Verifying checksum..."
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath
        $expected = Get-ExpectedChecksum -checksumsPath $checksumsPath -asset $asset
        $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
        if ($expected.ToLowerInvariant() -ne $actual) {
            throw "checksum mismatch for $asset"
        }
    }

    Write-Host "Extracting archive..."
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

    Write-Host "Installed: $destBin"
    & $destBin -v

    $pathParts = ($env:PATH -split ';' | ForEach-Object { $_.Trim() })
    if ($pathParts -notcontains $InstallDir) {
        Write-Host "$InstallDir is not in PATH. Add it from Windows Environment Variables."
    }
}
finally {
    if (Test-Path $tmpDir) {
        Remove-Item -Recurse -Force $tmpDir
    }
}
