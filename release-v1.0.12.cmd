@echo off
setlocal
cd /d "%~dp0"

git tag -a v1.0.12 -m "DualSenseHUB v1.0.12"
if errorlevel 1 (
  echo A tag v1.0.12 ja existe ou nao pode ser criada.
  pause
  exit /b 1
)

git push origin v1.0.12
if errorlevel 1 (
  echo Nao foi possivel enviar a tag.
  pause
  exit /b 1
)

echo A GitHub Action vai compilar e publicar o executavel em Releases.
echo Acompanhe a aba Actions do repositorio.
pause
endlocal
