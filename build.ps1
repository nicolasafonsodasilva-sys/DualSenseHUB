$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

Write-Host "Gerando recursos do Windows com o icone oficial..."
go run github.com/tc-hib/go-winres@v0.3.3 simply --arch amd64 --manifest none --icon assets/DualSenseHUB.ico

Write-Host "Formatando e verificando o codigo..."
go fmt ./...
go vet ./...
go test ./...

go build -buildvcs=false -trimpath -ldflags "-H=windowsgui -s -w" -o "dist\DualSenseHUB.exe" .

$hash = (Get-FileHash "dist\DualSenseHUB.exe" -Algorithm SHA256).Hash.ToLowerInvariant()
"$hash  DualSenseHUB.exe" | Set-Content -Encoding ascii -NoNewline "dist\SHA256SUMS.txt"

Write-Host "Build concluido: dist\DualSenseHUB.exe"
Write-Host "SHA-256: $hash"
