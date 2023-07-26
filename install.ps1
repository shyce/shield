param (
    [string]$version = "latest"
)

$username = "shyce"
$repo = "shield"
$binaryName = "shield"

if ($version -eq "latest") {
    try {
        $version = Invoke-RestMethod -Uri "https://api.github.com/repos/$username/$repo/releases/latest" | Select-Object -ExpandProperty tag_name
    }
    catch {
        Write-Host "Unable to fetch the latest version number."
        exit 1
    }
}

$url = "https://github.com/$username/$repo/releases/download/$version/$binaryName-$version-windows-amd64"
$output = "C:\Windows\System32\$binaryName.exe"

Write-Host "Downloading $binaryName..."
try {
    Invoke-WebRequest -Uri $url -OutFile $output
    Write-Host "$binaryName downloaded successfully!"
}
catch {
    Write-Host "Failed to download $binaryName."
    Write-Host $_.Exception.Message
    exit 1
}
