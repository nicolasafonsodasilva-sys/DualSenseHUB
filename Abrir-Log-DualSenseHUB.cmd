@echo off
set "LOG=%LOCALAPPDATA%\DualSenseHUB\DualSenseHUB-debug.log"
if not exist "%LOG%" (
  echo O log ainda nao foi criado: %LOG%
  pause
  exit /b 1
)
start "" notepad.exe "%LOG%"
