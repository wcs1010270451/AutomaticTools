#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

WINDOWS_SOURCE=../windows/dist/AutomaticTools.exe
ANDROID_SOURCE=../android/app/build/outputs/apk/release/app-release.apk
DOWNLOAD_DIR=downloads

if [ ! -f "$WINDOWS_SOURCE" ]; then
    echo "错误：缺少 Windows 安装包：$WINDOWS_SOURCE" >&2
    echo "请从 Windows 构建机复制正式版 AutomaticTools.exe。" >&2
    exit 1
fi

if [ ! -f "$ANDROID_SOURCE" ]; then
    echo "错误：缺少 Android Release APK：$ANDROID_SOURCE" >&2
    echo "请先在 android 目录执行 ./gradlew assembleRelease。" >&2
    exit 1
fi

mkdir -p "$DOWNLOAD_DIR"
cp "$WINDOWS_SOURCE" "$DOWNLOAD_DIR/AutomaticTools-Windows.exe"
cp "$ANDROID_SOURCE" "$DOWNLOAD_DIR/AutomaticTools-Android.apk"

echo "下载包已准备到 frontend/downloads。"
