#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

NAMESPACE=${DOCKERHUB_NAMESPACE:-wcs19890321}
TAG=${IMAGE_TAG:-latest}
IMAGE="$NAMESPACE/automatictools-backend:$TAG"

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    echo "错误：没有找到本地镜像 $IMAGE，请先执行 sh build.sh。" >&2
    exit 1
fi

echo "推送 $IMAGE ..."
docker push "$IMAGE"
echo "镜像已发布：$IMAGE"
