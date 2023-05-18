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
 
# Add the GitHub token to the web request headers if it was provided
$webRequestOptions = if ($githubToken) {
	@{ 'Headers' = @{ 'Authorization' = "Bearer $githubToken" } }
} else {
    @{}
}
 
if (!$version) {
    $q = (Invoke-WebRequest -Uri "$versionURL" -UseBasicParsing @webRequestOptions) | ConvertFrom-Json
    $version = $q.tag_name
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
 
Move-Item -Path "$directory\flow-cli.exe" -Destination "$directory\flow.exe" -Force
 
if ($addToPath) {
    Write-Output "Adding to PATH ..."
    $newPath = $Env:Path + ";$directory"
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath)
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, [System.EnvironmentVariableTarget]::User)
}
 
Write-Output "Done."
 
Start-Sleep -Seconds 1