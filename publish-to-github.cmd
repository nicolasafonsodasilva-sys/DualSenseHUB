@echo off
setlocal
cd /d "%~dp0"

where git >nul 2>nul
if errorlevel 1 (
  echo Git nao foi encontrado. Instale Git for Windows ou use GitHub Desktop.
  pause
  exit /b 1
)

echo Crie primeiro um repositorio PUBLICO e VAZIO chamado DualSenseHUB no GitHub.
echo Nao marque README, .gitignore ou LICENSE na tela de criacao.
echo.
set /p REPO_URL=Cole a URL HTTPS do repositorio: 
if "%REPO_URL%"=="" exit /b 1

if not exist .git git init
git add .
git commit -m "Publica DualSenseHUB v1.0.12 como codigo aberto"
if errorlevel 1 (
  echo Nao foi possivel criar o commit. Confira seu nome e email do Git.
  pause
  exit /b 1
)

git branch -M main
git remote remove origin >nul 2>nul
git remote add origin "%REPO_URL%"
git push -u origin main
if errorlevel 1 (
  echo O envio falhou. Confira a URL e conclua o login do GitHub no navegador.
  pause
  exit /b 1
)

echo.
echo Codigo publicado. Para criar a primeira versao para download, execute:
echo release-v1.0.12.cmd
pause
endlocal
