@echo off

:: ============================================
:: Go PowerControl Build Script
:: ============================================
:: Usage: build-windows.bat [-BuildInstaller]
:: ============================================

echo Building Go PowerControl for Windows...
echo.

:: ============================================
:: Parse Command Line Arguments
:: ============================================

set BUILD_INSTALLER=0
if /I "%~1"=="-BuildInstaller" set BUILD_INSTALLER=1

:: ============================================
:: Read .env file using PowerShell (handles special chars correctly)
:: ============================================

if exist "..\\.env" (
    echo Loading configuration from .env file...
    
    :: Use PowerShell script to parse .env and generate batch commands
    powershell -NoProfile -ExecutionPolicy Bypass -File "parse-env.ps1" > "%TEMP%\\gopowercontrol_env.bat"
    
    :: Execute the generated batch file to set variables
    call "%TEMP%\\gopowercontrol_env.bat"
    del "%TEMP%\\gopowercontrol_env.bat"
    
    echo.
) else (
    echo Warning: .env file not found. Continuing without signing configuration.
    echo.
)

:: Set defaults if not in .env
if "%VERSION%"=="" (
    echo Reading version from wails.json...
    for /f "tokens=2 delims=:, " %%a in ('findstr /C:"productVersion" ..\\wails.json') do (
        set VERSION=%%~a
    )
)
if "%COMPANY_NAME%"=="" set COMPANY_NAME=Unknown
if "%INSTALL_PATH%"=="" set INSTALL_PATH=C:\Program Files\Go PowerControl

echo Configuration:
echo   Version: %VERSION%
echo   Company: %COMPANY_NAME%
echo   Install Path: %INSTALL_PATH%
echo.

:: ============================================
:: Build with Wails
:: ============================================

echo Running Wails build...
cd ..
call wails build -platform windows/amd64
if %errorlevel% neq 0 (
    echo ERROR: Wails build failed!
    cd build
    pause
    exit /b 1
)
cd build
echo Wails build complete.
echo.

:: ============================================
:: Sign the Executable (Optional)
:: ============================================

set EXE_PATH=bin\go-powercontrol.exe
set SIGN_AVAILABLE=0

:: Check if signing is configured
if not "%SIGNTOOL%"=="" (
    if not "%CERTIFICATE_PATH%"=="" (
        if exist "%SIGNTOOL%" (
            if exist "%CERTIFICATE_PATH%" (
                set SIGN_AVAILABLE=1
            )
        )
    )
)

if %SIGN_AVAILABLE%==1 (
    echo Signing executable with certificate...
    "%SIGNTOOL%" sign /f "%CERTIFICATE_PATH%" /p "%CERTIFICATE_PASSWORD%" /tr http://timestamp.digicert.com /td sha256 /fd sha256 "%EXE_PATH%"
    if %errorlevel% equ 0 (
        echo Executable signed successfully.
    ) else (
        echo Warning: Failed to sign executable. Continuing anyway.
    )
    echo.
) else (
    echo Skipping code signing ^(signtool or certificate not found^).
    echo.
)

:: ============================================
:: Build Installer (if flag is set)
:: ============================================

if %BUILD_INSTALLER%==0 goto :SkipInstaller

echo Building installer with NSIS...
echo.

:: Check if NSIS is available
set "NSIS_PATH="

if exist "C:\Program Files (x86)\NSIS\makensis.exe" goto :FoundNSIS_x86
if exist "C:\Program Files\NSIS\makensis.exe" goto :FoundNSIS_64

where makensis.exe >nul 2>&1
if not errorlevel 1 goto :FoundNSIS_PATH

goto :NSISNotFound

:FoundNSIS_x86
set "NSIS_PATH=C:\Program Files (x86)\NSIS\makensis.exe"
goto :NSISFound

:FoundNSIS_64
set "NSIS_PATH=C:\Program Files\NSIS\makensis.exe"
goto :NSISFound

:FoundNSIS_PATH
set "NSIS_PATH=makensis.exe"
goto :NSISFound

:NSISNotFound
echo ERROR: NSIS not found. Please install NSIS from https://nsis.sourceforge.io/
pause
exit /b 1

:NSISFound

echo Found NSIS at: %NSIS_PATH%
echo.

:: Build the installer
"%NSIS_PATH%" installer.nsi
if %errorlevel% neq 0 (
    echo ERROR: Installer build failed!
    pause
    exit /b 1
)

echo Installer built successfully.
echo.

:: Sign the installer (if signing is available)
if %SIGN_AVAILABLE%==1 (
    echo Signing installer...
    
    :: Find and sign the installer file
    for %%F in (bin\GoPowerControl-Installer-*.exe) do call :SignInstallerFile "%%F"
    
    echo.
)

:SkipInstaller

:: ============================================
:: Build Complete
:: ============================================

echo ============================================
echo Build Complete!
echo ============================================
echo.
echo Output:
echo   Executable: build\bin\go-powercontrol.exe
if %BUILD_INSTALLER%==1 (
    echo   Installer: build\bin\GoPowerControl-Installer-%VERSION%.exe
)
echo.
pause
exit /b 0

:: ============================================
:: Subroutine: Sign Installer File
:: ============================================
:SignInstallerFile
set "INSTALLER_FILE=%~1"
echo Signing: %INSTALLER_FILE%
"%SIGNTOOL%" sign /f "%CERTIFICATE_PATH%" /p "%CERTIFICATE_PASSWORD%" /tr http://timestamp.digicert.com /td sha256 /fd sha256 "%INSTALLER_FILE%"
if %errorlevel% equ 0 (
    echo Installer signed successfully.
) else (
    echo Warning: Failed to sign installer. Error code: %errorlevel%
)
goto :eof
