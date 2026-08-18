#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

NAMESPACE=${DOCKERHUB_NAMESPACE:-wcs19890321}
TAG=${IMAGE_TAG:-latest}
IMAGE="$NAMESPACE/automatictools-frontend:$TAG"

sh ./prepare-downloads.sh

if ! docker buildx version >/dev/null 2>&1; then
    echo "错误：Docker Buildx 不可用，请先启动 Docker Desktop。" >&2
    exit 1
fi

echo "构建 $IMAGE ..."
docker buildx build --platform linux/amd64 --pull --load --tag "$IMAGE" .
echo "本地镜像已生成：$IMAGE"
