@echo off
setlocal

cd /d "%~dp0"

set "WINDOWS_SOURCE=..\windows\dist\AutomaticTools.exe"
set "ANDROID_SOURCE=..\android\app\build\outputs\apk\release\app-release.apk"
set "DOWNLOAD_DIR=downloads"

if not exist "%WINDOWS_SOURCE%" (
    echo ERROR: Windows package not found: %WINDOWS_SOURCE%
    echo Build windows\AutomaticTools.spec first.
    exit /b 1
)

if not exist "%ANDROID_SOURCE%" (
    echo ERROR: Android release package not found: %ANDROID_SOURCE%
    echo Run android\gradlew.bat assembleRelease first.
    exit /b 1
)

if not exist "%DOWNLOAD_DIR%" mkdir "%DOWNLOAD_DIR%"
copy /Y "%WINDOWS_SOURCE%" "%DOWNLOAD_DIR%\AutomaticTools-Windows.exe" >nul
copy /Y "%ANDROID_SOURCE%" "%DOWNLOAD_DIR%\AutomaticTools-Android.apk" >nul

echo Download packages prepared in frontend\downloads.
exit /b 0
