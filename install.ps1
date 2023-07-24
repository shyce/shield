param (
    [string]$version = "latest"
)

$username = "shyce"
$repo = "shield"
$binaryName = "shield"

if ($version -eq "latest") {
    $version = Invoke-RestMethod -Uri "https://api.github.com/repos/$username/$repo/releases/latest" | Select-Object -ExpandProperty tag_name
}

$url = "https://github.com/$username/$repo/releases/download/$version/$binaryName-$version-windows-amd64.exe"
$output = "$HOME/$binaryName.exe"

Write-Host "Downloading $binaryName..."
Invoke-WebRequest -Uri $url -OutFile $output

Write-Host "$binaryName downloaded successfully!"
