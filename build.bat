@echo off
setlocal

:: Default output directory is "dist" if no argument is passed
set "OUT_DIR=%~1"
if "%OUT_DIR%"=="" set "OUT_DIR=dist"

echo Target folder: "%OUT_DIR%"

:: Create target directory if it doesn't exist
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

:: Remove existing executable to ensure a fresh replacement
if exist "%OUT_DIR%\mediaserver.exe" del /f /q "%OUT_DIR%\mediaserver.exe"

echo Building mediaserver.exe...
go build -ldflags="-s -w" -o "%OUT_DIR%\mediaserver.exe" .\cmd\server

if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] Build failed! Check the errors above.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo [SUCCESS] Build succeeded! Executable saved to: %OUT_DIR%\mediaserver.exe
endlocal
