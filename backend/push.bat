@echo off
setlocal

cd /d "%~dp0"
set "IMAGE=wcs19890321/automatictools-backend:latest"

docker image inspect "%IMAGE%" >nul 2>&1
if errorlevel 1 (
    echo ERROR: Local image not found: %IMAGE%
    echo Run build.bat first.
    exit /b 1
)

echo Pushing %IMAGE% ...
docker push "%IMAGE%"
if errorlevel 1 (
    echo ERROR: Backend image push failed.
    exit /b 1
)

echo Published: %IMAGE%
exit /b 0
