@echo off
setlocal
cd /d "%~dp0"
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\taller11.ps1" -Paso Todo
set "TALLER11_EXIT=%ERRORLEVEL%"
echo.
if not "%TALLER11_EXIT%"=="0" echo El taller se detuvo. Copia el mensaje de error para revisarlo.
pause
exit /b %TALLER11_EXIT%
