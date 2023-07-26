# Create directories if they do not exist
New-Item -ItemType Directory -Force -Path secrets, vendors, test

# Generate random content
$randomContent = -join ((65..90) + (97..122) | Get-Random -Count 180 | % { [char]$_ })

# Create secret files
Set-Content -Path "my.secret" -Value "This file should be encrypted: $randomContent"
Set-Content -Path "secrets/info.txt" -Value "This file should be encrypted: $randomContent"
Set-Content -Path "secrets/key.pem" -Value "This file should be encrypted: $randomContent"
Set-Content -Path "temp.secret" -Value "This file should not be encrypted: $randomContent"

# Create non-secret files
Set-Content -Path "vendors/info.txt" -Value "This file should not be encrypted: $randomContent"
Set-Content -Path "test/info.txt" -Value "This file should not be encrypted: $randomContent"
