@echo off
setlocal enabledelayedexpansion

echo Building EffiHub...

:: 检查 Docker 是否可用
where docker >nul 2>nul
if errorlevel 1 (
    echo [Error] Docker is not installed.
    goto :end
)

:: 检查 Docker daemon 是否运行
docker info >nul 2>nul
if errorlevel 1 (
    echo [Error] Docker daemon is not running.
    goto :end
)

:: 下载依赖
echo [1/5] Downloading dependencies...
go mod tidy

:: 创建输出目录
if not exist dist mkdir dist

:: 编译 Windows
echo [2/5] Building for Windows x64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/effihub.exe .

:: 编译 Linux x86_64
echo [3/5] Building for Linux x64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/effihub .

:: 复制静态文件
echo [4/5] Copying static files...
xcopy /E /I /Y static dist\static >nul 2>nul

echo.
echo [5/5] Build complete!
echo Output: dist/

:: Docker 构建并推送
:: 用法: build.bat --push [版本号]
:: 示例: build.bat --push v1.0.0
if "%1"=="--push" (
    set VERSION=latest
    if not "%2"=="" set VERSION=%2

    echo.
    echo Building Docker image: jacyli/effihub:%VERSION%

    set DOCKER_BUILDKIT=1
    docker build -t jacyli/effihub:%VERSION% .

    echo Pushing image...
    docker push jacyli/effihub:%VERSION%

    if not "%VERSION%"=="latest" (
        docker tag jacyli/effihub:%VERSION% jacyli/effihub:latest
        docker push jacyli/effihub:latest
    )

    echo Push complete!
)

:: 多架构构建并推送
:: 用法: build.bat --push-multi [版本号]
if "%1"=="--push-multi" (
    set VERSION=latest
    if not "%2"=="" set VERSION=%2

    echo.
    echo Setting up Docker buildx...

    docker buildx create --use --name effihub-builder >nul 2>nul || docker buildx use effihub-builder
    docker buildx inspect --bootstrap >nul 2>nul

    echo Building multi-platform image: jacyli/effihub:%VERSION%

    docker buildx build --platform linux/amd64,linux/arm64 -t jacyli/effihub:%VERSION% --push .

    if not "%VERSION%"=="latest" (
        docker buildx build --platform linux/amd64,linux/arm64 -t jacyli/effihub:latest --push .
    )

    echo Multi-platform push complete!
)

:end
echo.
pause
