@echo off
setlocal

cd /d "%~dp0"
set "OUTPUT_DIR=dist"
set "OUTPUT_EXE=%OUTPUT_DIR%\VRCMinimalMuteController.exe"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo Building VRCMinimalMuteController.exe...
go.exe build -ldflags "-H=windowsgui" -o "%OUTPUT_EXE%" .
if errorlevel 1 goto :fail

echo Embedding Icon.ico into executable...
go.exe run ./tools/seticon "%OUTPUT_EXE%" Icon.ico
if errorlevel 1 goto :fail

echo Refreshing Windows icon cache...
if exist "%SystemRoot%\System32\ie4uinit.exe" (
  "%SystemRoot%\System32\ie4uinit.exe" -show >nul 2>&1
)

echo.
echo Build completed:
echo %cd%\%OUTPUT_EXE%
exit /b 0

:fail
echo.
echo Build failed.
exit /b 1
