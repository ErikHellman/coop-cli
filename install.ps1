$ErrorActionPreference = "Stop"

$Repo = "ErikHellman/coop-cli"
$Binary = "coop-cli.exe"
$InstallDir = "$env:LOCALAPPDATA\coop-cli"

$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

if ($args.Count -gt 0) {
    $Version = $args[0]
} else {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $Release.tag_name
}

$VersionNum = $Version.TrimStart("v")
$Archive = "coop-cli_${VersionNum}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"

$TmpDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "coop-cli-install-$(Get-Random)")

try {
    Write-Host "Downloading coop-cli $Version for windows/$Arch..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $TmpDir $Archive)

    Write-Host "Extracting..."
    Expand-Archive -Path (Join-Path $TmpDir $Archive) -DestinationPath $TmpDir -Force

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    Move-Item -Path (Join-Path $TmpDir $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your PATH."
    }

    Write-Host "Installed coop-cli $Version to $InstallDir\$Binary"
} finally {
    Remove-Item -Recurse -Force $TmpDir
}
