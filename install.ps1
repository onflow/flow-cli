
param ($version="", $directory = "$env:APPDATA\Flow")

$baseURL = "https://storage.googleapis.com/flow-cli"

if (!$version) {
    $version = (Invoke-WebRequest -Uri "$baseURL/version.txt" -UseBasicParsing).Content.Trim()
}

Write-Output("Installing version {0} ..." -f $version)

New-Item -ItemType Directory -Force -Path $directory | Out-Null

$progressPreference = 'silentlyContinue'

Invoke-WebRequest -Uri "$baseURL/flow-x86_64-windows-$version" -UseBasicParsing -OutFile "$directory\flow.exe"

$newPath = $Env:Path + ";$directory"
[System.Environment]::SetEnvironmentVariable("PATH", $newPath)
[System.Environment]::SetEnvironmentVariable("PATH", $newPath, [System.EnvironmentVariableTarget]::User)

Write-Output "Done."

Start-Sleep -Seconds 1
