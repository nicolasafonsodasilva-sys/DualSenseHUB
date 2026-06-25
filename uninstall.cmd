@echo off
setlocal

echo Encerrando DualSenseHUB...
taskkill /F /IM DualSenseHUB.exe >nul 2>nul

reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "DualSenseHUB" /f >nul 2>nul
reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "DualSensePower" /f >nul 2>nul

timeout /t 1 /nobreak >nul

if exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\DualSenseHUB.lnk" (
  del /f /q "%APPDATA%\Microsoft\Windows\Start Menu\Programs\DualSenseHUB.lnk"
)

if exist "%LOCALAPPDATA%\DualSenseHUB" (
  rmdir /s /q "%LOCALAPPDATA%\DualSenseHUB"
)

echo DualSenseHUB foi removido deste usuario.
pause
endlocal
