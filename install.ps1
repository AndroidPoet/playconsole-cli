[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repo = if ($env:GPC_REPO) { $env:GPC_REPO } else { "AndroidPoet/playconsole-cli" }
$version = if ($env:GPC_VERSION) { $env:GPC_VERSION } else { "latest" }
$installDir = if ($env:GPC_INSTALL_DIR) {
    $env:GPC_INSTALL_DIR
} elseif ($env:LOCALAPPDATA) {
    Join-Path $env:LOCALAPPDATA "Programs\playconsole-cli"
} else {
    Join-Path $env:USERPROFILE "bin\playconsole-cli"
}

if ($repo -notmatch "^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$") {
    throw "Invalid GPC_REPO value. Expected owner/repository."
}
if ($version -ne "latest" -and $version -notmatch "^[A-Za-z0-9._-]+$") {
    throw "Invalid GPC_VERSION value."
}

$machineArch = if ($env:PROCESSOR_ARCHITEW6432) {
    $env:PROCESSOR_ARCHITEW6432
} else {
    $env:PROCESSOR_ARCHITECTURE
}
$arch = switch ($machineArch.ToUpperInvariant()) {
    "AMD64" { "amd64"; break }
    "ARM64" { "arm64"; break }
    default { throw "Unsupported Windows architecture: $machineArch (supported: amd64, arm64)." }
}

$releaseEndpoint = if ($version -eq "latest") {
    "https://api.github.com/repos/$repo/releases/latest"
} else {
    "https://api.github.com/repos/$repo/releases/tags/$version"
}

$headers = @{ Accept = "application/vnd.github+json" }
if ($env:GITHUB_TOKEN) {
    $headers.Authorization = "Bearer $($env:GITHUB_TOKEN)"
}

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) "playconsole-cli-$([guid]::NewGuid().ToString('N'))"
$archivePath = Join-Path $tempRoot "archive.zip"
$extractDir = Join-Path $tempRoot "extract"

try {
    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
    Write-Host "==> Fetching playconsole-cli release ($repo, windows/$arch)..." -ForegroundColor Green

    $release = Invoke-RestMethod -Uri $releaseEndpoint -Headers $headers
    $asset = @($release.assets | Where-Object {
        $_.name -match "^playconsole-cli_.*_windows_${arch}\.zip$"
    }) | Select-Object -First 1

    if (-not $asset) {
        throw "No pre-built Windows/$arch asset was found for release '$version'. Install Go and build from source with: go install github.com/$repo/cmd/playconsole-cli@latest"
    }

    Write-Host "==> Downloading $($asset.name)..." -ForegroundColor Green
    Invoke-WebRequest -Uri $asset.browser_download_url -Headers $headers -OutFile $archivePath
    Expand-Archive -LiteralPath $archivePath -DestinationPath $extractDir -Force

    $binary = Get-ChildItem -LiteralPath $extractDir -Filter "playconsole-cli.exe" -File -Recurse | Select-Object -First 1
    if (-not $binary) {
        throw "The downloaded archive does not contain playconsole-cli.exe."
    }

    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    $target = Join-Path $installDir "playconsole-cli.exe"
    Copy-Item -LiteralPath $binary.FullName -Destination $target -Force

    # Windows has no portable user-level symlink equivalent. A copy keeps the
    # gpc alias usable without requiring Developer Mode or administrator rights.
    Copy-Item -LiteralPath $target -Destination (Join-Path $installDir "gpc.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @()
    if ($userPath) {
        $pathEntries = @($userPath -split ';' | Where-Object { $_ })
    }
    if (-not ($pathEntries | Where-Object { $_.TrimEnd('\') -ieq $installDir.TrimEnd('\') })) {
        [Environment]::SetEnvironmentVariable("Path", (($pathEntries + $installDir) -join ';'), "User")
        $env:Path = "$installDir;$env:Path"
        Write-Host "==> Added $installDir to the user PATH." -ForegroundColor Green
    }

    Write-Host "==> Installed playconsole-cli to $target" -ForegroundColor Green
    & $target version
    Write-Host "Installation complete. Open a new terminal, then use 'gpc' or 'playconsole-cli'." -ForegroundColor Green
} finally {
    if (Test-Path -LiteralPath $tempRoot) {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}
