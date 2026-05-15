#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building EffiHub...${NC}"

# 检查 Docker 是否可用
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed or not running.${NC}"
    exit 1
fi

# 检查 Docker daemon 是否运行
if ! docker info &> /dev/null; then
    echo -e "${RED}Error: Docker daemon is not running.${NC}"
    exit 1
fi

# 下载依赖
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod tidy

# 创建输出目录
mkdir -p dist

# 编译 Linux x86_64
echo -e "${YELLOW}Building for Linux x86_64...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/effihub .

# 编译 Windows
echo -e "${YELLOW}Building for Windows...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/effihub.exe .

# 复制静态文件
echo -e "${YELLOW}Copying static files...${NC}"
cp -r static dist/

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo "Output: dist/"

# Docker 构建并推送
# 用法: ./build.sh --push [版本号]
# 示例: ./build.sh --push v1.0.0
if [ "$1" = "--push" ]; then
    VERSION=${2:-latest}

    # 验证版本格式
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]] && [ "$VERSION" != "latest" ]; then
        echo -e "${YELLOW}Warning: Version '$VERSION' doesn't follow semantic versioning. Proceeding anyway...${NC}"
    fi

    echo ""
    echo -e "${GREEN}Building Docker image (jacyli/effihub:${VERSION})...${NC}"

    # 使用 BuildKit 构建，更高效
    export DOCKER_BUILDKIT=1

    docker build -t jacyli/effihub:${VERSION} \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        .

    echo -e "${GREEN}Pushing Docker image...${NC}"
    docker push jacyli/effihub:${VERSION}

    # 非 latest 版本同时更新 latest 标签
    if [ "$VERSION" != "latest" ]; then
        docker tag jacyli/effihub:${VERSION} jacyli/effihub:latest
        docker push jacyli/effihub:latest
    fi

    echo ""
    echo -e "${GREEN}Push complete!${NC}"
    echo "Image: jacyli/effihub:${VERSION}"
fi

# 多架构构建并推送 (需要 docker buildx)
if [ "$1" = "--push-multi" ]; then
    VERSION=${2:-latest}

    echo ""
    echo -e "${GREEN}Setting up Docker buildx for multi-platform build...${NC}"

    # 确保 buildx builder 存在
    docker buildx create --use --name effihub-builder 2>/dev/null || docker buildx use effihub-builder
    docker buildx inspect --bootstrap

    echo -e "${GREEN}Building multi-platform Docker image...${NC}"

    docker buildx build --platform linux/amd64,linux/arm64 \
        -t jacyli/effihub:${VERSION} \
        --push \
        .

    # 非 latest 版本同时更新 latest 标签
    if [ "$VERSION" != "latest" ]; then
        docker buildx build --platform linux/amd64,linux/arm64 \
            -t jacyli/effihub:latest \
            --push \
            .
    fi

    echo ""
    echo -e "${GREEN}Multi-platform push complete!${NC}"
    echo "Images:"
    echo "  - jacyli/effihub:${VERSION}"
    echo "  - linux/amd64"
    echo "  - linux/arm64"
fi
