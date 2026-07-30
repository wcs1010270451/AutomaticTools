@echo off
setlocal

cd /d "%~dp0"
set "IMAGE=wcs19890321/automatictools-frontend:latest"

docker buildx version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker Buildx is not available.
    exit /b 1
)

echo Building %IMAGE% ...
docker buildx build --platform linux/amd64 --pull --load --tag "%IMAGE%" .
if errorlevel 1 (
    echo ERROR: Frontend image build failed.
    exit /b 1
)

echo Built locally: %IMAGE%
exit /b 0
