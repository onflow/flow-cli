<#
.Synopsis
    Install the Flow CLI on Windows.
.DESCRIPTION
    By default, the latest release will be installed.
    If '-Version' is specified, then the given version is installed.
.Parameter Directory
    The destination path to install to.
.Parameter Version
    The version to install.
.Parameter AddToPath
    Add the absolute destination path to the 'User' scope environment variable 'Path'.
.Parameter GitHubToken
    Optional GitHub token to use to prevent rate limiting.
.EXAMPLE
    Install the current version
    .\install.ps1
.EXAMPLE
    Install version v0.5.2
    .\install.ps1 -Version v0.5.2
.EXAMPLE
    Invoke-Expression "& { $(Invoke-RestMethod 'https://storage.googleapis.com/flow-cli/install.ps1') }"
#>
param (
    [string] $version="",
    [string] $directory = "$env:APPDATA\Flow",
    [bool] $addToPath = $true,
    [string] $githubToken = ""
)

Set-StrictMode -Version 3.0

# Enable support for ANSI escape sequences
Set-ItemProperty HKCU:\Console VirtualTerminalLevel -Type DWORD 1

$ErrorActionPreference = "Stop"

$repo = "onflow/flow-cli"
$versionURL = "https://api.github.com/repos/$repo/releases/latest"
$assetsURL = "https://github.com/$repo/releases/download"

# Function to get the latest version
function Get-Version {
    param (
        [string]$repo,
        [string]$searchTerm,
        [string]$githubTokenHeader
    )

    $page = 1
    $version = null

    while (-not $version) {
        if ($githubTokenHeader) {
            $response = Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases?per_page=100&page=$page" -Headers @{ 'Authorization' = $githubTokenHeader } -UseBasicParsing -ErrorAction SilentlyContinue
        } else {
            $response = Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases?per_page=100&page=$page" -UseBasicParsing -ErrorAction SilentlyContinue
        }

        $status = $response.StatusCode

        if ($status -eq 403 -and $githubTokenHeader) {
            Write-Output "Failed to get latest version from Github API, is your GITHUB_TOKEN valid? Trying without authentication..."
            $githubTokenHeader = ""
            continue
        }

        if ($status -ne 200) {
            Write-Output "Failed to get latest version from Github API, please manually specify a version to install as an argument to this script."
            return $null
        }

        $jsonResponse = $response.Content | ConvertFrom-Json

        foreach ($release in $jsonResponse) {
            if ($release.tag_name -like "*$searchTerm*") {
                $version = $release.tag_name
                break
            }
        }

        $page++
    }

    return $version
}

if (-not $version) {
    $version = Get-Version -repo $repo -searchTerm "cadence-v1.0.0" -githubTokenHeader $githubToken
}

Write-Output("Installing version {0} ..." -f $version)

New-Item -ItemType Directory -Force -Path $directory | Out-Null

$progressPreference = 'silentlyContinue'

Invoke-WebRequest -Uri "$assetsURL/$version/flow-cli-$version-windows-amd64.zip" -UseBasicParsing -OutFile "$directory\flow.zip" @webRequestOptions

Expand-Archive -Path "$directory\flow.zip" -DestinationPath "$directory" -Force

try {
    Stop-Process -Name flow -Force
    Start-Sleep -Seconds 1
}
catch {}

Move-Item -Path "$directory\flow-cli.exe" -Destination "$directory\flow-c1.exe" -Force

# Check if the directory is already in the PATH
$existingPaths = [Environment]::GetEnvironmentVariable("PATH", [System.EnvironmentVariableTarget]::User).Split(';')

if ($addToPath -and $existingPaths -notcontains $directory) {
    Write-Output "Adding to PATH ..."
    $processPath = [System.Environment]::GetEnvironmentVariable('PATH', [System.EnvironmentVariableTarget]::Process) + ";$directory"
    $userPath = [System.Environment]::GetEnvironmentVariable('PATH', [System.EnvironmentVariableTarget]::User) + ";$directory"
    [System.Environment]::SetEnvironmentVariable("PATH", $processPath, [System.EnvironmentVariableTarget]::Process)
    [System.Environment]::SetEnvironmentVariable("PATH", $userPath, [System.EnvironmentVariableTarget]::User)
}

Write-Output "\nSuccessfully installed Flow CLI $version"
Write-Output "PRE-RELEASE: Use the 'flow-c1' command to interact with this Cadence 1.0 CLI pre-release."

Start-Sleep -Seconds 1