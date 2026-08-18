# AutomaticTools

AutomaticTools 是一个跨平台实用工具平台，包含官网、管理端、Go 后端、Android 客户端和 Windows 客户端。当前主要工具是自动点击，用户通过邮箱注册登录，并使用授权码开通工具。

## 项目目录

- `backend`：Gin + GORM API，使用 PostgreSQL
- `frontend/index`：静态官网及客户端下载页
- `frontend/admin`：Vue 3 + TypeScript 管理端
- `frontend`：官网和管理端的统一 Nginx 镜像
- `android`：Kotlin Android 客户端
- `windows`：Python + Tkinter Windows 客户端
- `prod`：Docker Compose、Caddy 和生产环境运维脚本
- `docs`：开发及迁移文档

官网与管理端源码相互独立，但会一起构建到 `automatictools-frontend` 镜像，在同一个 Nginx 容器中运行。

## 开发文档

- [macOS 开发环境与迁移指南](docs/macos-development.md)
- [后端接口与配置](backend/README.md)
- [Android 客户端](android/README.md)
- [Windows 客户端](windows/README.md)
- [前端项目与发布](frontend/README.md)
- [生产部署](prod/README.md)

## 构建并发布镜像

先启动 Docker Desktop，并登录 Docker Hub：

```sh
docker login
```

Windows PowerShell：

```powershell
.\backend\build.bat
.\backend\push.bat
.\frontend\build.bat
.\frontend\push.bat
```

macOS 终端：

```sh
sh backend/build.sh
sh backend/push.sh
sh frontend/build.sh
sh frontend/push.sh
```

默认发布以下 `linux/amd64` 镜像：

- `wcs19890321/automatictools-backend:latest`
- `wcs19890321/automatictools-frontend:latest`

前端构建前必须准备好两个正式安装包：

- `windows/dist/AutomaticTools.exe`
- `android/app/build/outputs/apk/release/app-release.apk`

构建脚本会把它们复制到 `frontend/downloads`。这些二进制文件不进入 Git；迁移到新电脑时，需要单独复制 Windows 正式包，并重新构建 Android Release APK。

## 服务器部署

首次部署时，将完整的 `prod` 目录上传到服务器，必须包含隐藏文件 `prod/.env`。服务器只拉取镜像，不编译源码：

```sh
cd ~/prod
chmod +x start.sh update.sh stop.sh
./start.sh
```

后续完成本地镜像构建和推送后，在服务器更新：

```sh
cd ~/prod
./update.sh
```

当前正式地址：

- `https://autumnwind.top`
- `https://admin.autumnwind.top`

验证服务：

```sh
docker compose ps
curl https://autumnwind.top/health
curl https://autumnwind.top/api/tools
curl -I https://admin.autumnwind.top
```

后端 `8088` 和 PostgreSQL `5432` 只在 Docker 内部网络使用。公网请求统一由 Caddy 的 80/443 端口转发。

## 重要备份

以下文件包含密钥、密码或构建产物，已被 `.gitignore` 排除，不能只依赖 Git 迁移：

- `android/keystore/automatictools-release.jks`
- `android/keystore.properties`
- `backend/config.json`
- `prod/.env` 和 `prod/secrets/`
- `windows/dist/AutomaticTools.exe`

请使用加密存储单独备份。丢失 Android 发布密钥后，已安装的正式版将无法通过新 APK 原地升级。
