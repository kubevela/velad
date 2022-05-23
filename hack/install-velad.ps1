# Implemented based on Dapr Cli https://github.com/dapr/cli/tree/master/install

param (
    [string]$Version,
    [string]$VelaRoot = "c:\vela"
)

Write-Output ""
$ErrorActionPreference = 'stop'

#Escape space of VelaRoot path
$VelaRoot = $VelaRoot -replace ' ', '` '

# Constants
$VelaDBuildName = "velad"
$VelaDFileName = "velad.exe"
$VelaDFilePath = "${VelaRoot}\${VelaDFileName}"
$RemoteURL = "https://static.kubevela.net/binary/velad"

if ((Get-ExecutionPolicy) -gt 'RemoteSigned' -or (Get-ExecutionPolicy) -eq 'ByPass') {
    Write-Output "PowerShell requires an execution policy of 'RemoteSigned'."
    Write-Output "To make this change please run:"
    Write-Output "'Set-ExecutionPolicy RemoteSigned -scope CurrentUser'"
    break
}

# Change security protocol to support TLS 1.2 / 1.1 / 1.0 - old powershell uses TLS 1.0 as a default protocol
[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"

# Check if VelaD is installed.
if (Test-Path $VelaDFilePath -PathType Leaf) {
    Write-Warning "velad is detected - $VelaDFilePath"
    Invoke-Expression "$VelaDFilePath --version"
    Write-Output "Reinstalling VelaD..."
}
else {
    Write-Output "Installing VelaD..."
}

# Create Vela Directory
Write-Output "Creating $VelaRoot directory"
New-Item -ErrorAction Ignore -Path $VelaRoot -ItemType "directory"
if (!(Test-Path $VelaRoot -PathType Container)) {
    throw "Cannot create $VelaRoot"
}

# Filter windows binary and download archive
$os_arch = "windows-amd64"
$vela_cli_filename = "vela"
if (!$Version) {
    $Version = Invoke-RestMethod -Headers $githubHeader -Uri "${RemoteURL}/latest_version" -Method Get
    $Version = $Version.Trim()
}
if (!$Version.startswith("v")) {
    $Version = "v" + $Version
}

$assetName = "${vela_cli_filename}-${Version}-${os_arch}.zip"
$zipFileUrl = "${RemoteURL}/${Version}/${assetName}"

$zipFilePath = $VelaRoot + "\" + $assetName
Write-Output "Downloading $zipFileUrl ..."

Invoke-WebRequest -Uri $zipFileUrl -OutFile $zipFilePath
if (!(Test-Path $zipFilePath -PathType Leaf)) {
    throw "Failed to download Vela Cli binary - $zipFilePath"
}

# Extract VelaD CLI to $VelaRoot
Write-Output "Extracting $zipFilePath..."
Microsoft.Powershell.Archive\Expand-Archive -Force -Path $zipFilePath -DestinationPath $VelaRoot
$ExtractedVelaDFilePath = "${VelaRoot}\${os_arch}\${VelaDBuildName}"
Copy-Item $ExtractedVelaDFilePath -Destination $VelaDFilePath
if (!(Test-Path $VelaDFilePath -PathType Leaf)) {
    throw "Failed to extract VelaD archive - $zipFilePath"
}

# Check the VelaD version
Invoke-Expression "$VelaDFilePath --version"

# Clean up zipfile
Write-Output "Clean up $zipFilePath..."
Remove-Item $zipFilePath -Force

# Add VelaRoot directory to User Path environment variable
Write-Output "Try to add $VelaRoot to User Path Environment variable..."
$UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPathEnvironmentVar -like '*vela*') {
    Write-Output "Skipping to add $VelaRoot to User Path - $UserPathEnvironmentVar"
}
else {
    [System.Environment]::SetEnvironmentVariable("PATH", $UserPathEnvironmentVar + ";$VelaRoot", "User")
    $UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Output "Added $VelaRoot to User Path - $UserPathEnvironmentVar"
}

Write-Output "`r`VelaD is installed successfully."
Write-Output "To get started with KubeVela and VelaD, please visit https://kubevela.io."
