@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
  echo Go nao foi encontrado no PATH.
  exit /b 1
)

if not exist dist mkdir dist

echo Gerando recursos do Windows com o icone oficial...
go run github.com/tc-hib/go-winres@v0.3.3 simply --arch amd64 --manifest none --icon assets/DualSenseHUB.ico
if errorlevel 1 exit /b 1

go fmt ./...
if errorlevel 1 exit /b 1

go vet ./...
if errorlevel 1 exit /b 1

go test ./...
if errorlevel 1 exit /b 1

go build -buildvcs=false -trimpath -ldflags "-H=windowsgui -s -w" -o "dist\DualSenseHUB.exe" .
if errorlevel 1 exit /b 1

powershell -NoProfile -Command "$h=(Get-FileHash 'dist\DualSenseHUB.exe' -Algorithm SHA256).Hash.ToLowerInvariant(); Set-Content -Encoding ascii -NoNewline 'dist\SHA256SUMS.txt' ($h + '  DualSenseHUB.exe'); Write-Host ('SHA-256: ' + $h)"

echo Build concluido: dist\DualSenseHUB.exe
endlocal
