# macOS 开发环境与迁移指南

本文用于把 AutomaticTools 从 Windows 迁移到 macOS，并在 Mac 上继续开发后端、前端和 Android 客户端。Windows 客户端依赖 Win32 API，不能直接在 macOS 上运行或打包。

## 1. 迁移前必须备份

先在旧电脑确认代码已经提交并推送，再将下列被 Git 忽略的文件复制到加密移动硬盘或其他安全存储：

| 文件或目录 | 用途 | 新 Mac 上的位置 |
| --- | --- | --- |
| `android/keystore/automatictools-release.jks` | Android 正式版签名密钥 | 保持原路径 |
| `android/keystore.properties` | Android 签名密码和别名 | 保持原路径 |
| `backend/config.json` | 本地数据库、JWT、SMTP 等配置 | 保持原路径 |
| `prod/.env` | 生产环境配置 | 保持原路径 |
| `prod/secrets/` | 生产密钥文件 | 保持原路径 |
| `windows/dist/AutomaticTools.exe` | 官网发布的 Windows 正式包 | 保持原路径 |

建议同时导出一份 PostgreSQL 备份。上述文件都不得提交到公开 Git 仓库。

## 2. 安装开发工具

先安装 Apple 命令行工具：

```sh
xcode-select --install
```

推荐安装 Homebrew，然后安装 Go 1.24 或更高版本和 Node.js 22：

```sh
brew install go node@22
echo 'export PATH="$(brew --prefix node@22)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

还需要安装：

- Docker Desktop for Mac
- Android Studio，以及 Android SDK Platform 35 和 Platform Tools
- PostgreSQL 17（可选；本地开发也可以直接使用 Docker）

确认环境：

```sh
git --version
go version
node --version
npm --version
docker version
```

## 3. 克隆项目

```sh
mkdir -p ~/Code
cd ~/Code
git clone git@github.com:wcs1010270451/AutomaticTools.git
cd AutomaticTools
git config core.autocrlf input
```

将第 1 节备份的本地配置、密钥和 Windows 正式包恢复到表格中的原路径。不要用 Git 强制添加这些文件。

## 4. 启动本地 PostgreSQL

最省事的方式是使用 Docker：

```sh
docker run --name automatictools-postgres-dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=automatic_tools \
  -p 127.0.0.1:5432:5432 \
  -d postgres:17-alpine
```

对应的本地连接串为：

```text
postgres://postgres:postgres@localhost:5432/automatic_tools?sslmode=disable
```

停止和再次启动数据库：

```sh
docker stop automatictools-postgres-dev
docker start automatictools-postgres-dev
```

也可以执行 `brew install postgresql@17` 使用本机 PostgreSQL。

## 5. 启动后端

首次使用时，从示例创建本地配置并填写真实值：

```sh
test -f backend/config.json || cp backend/config.example.json backend/config.json
cd backend
go mod download
go test ./...
go run ./cmd/api
```

默认监听 `http://127.0.0.1:8088`。启动时会执行幂等初始化 SQL，因此不需要每次手工运行建表脚本。

验证：

```sh
curl http://127.0.0.1:8088/health
curl http://127.0.0.1:8088/api/tools
```

## 6. 启动管理端和官网

管理端：

```sh
cd frontend/admin
npm ci
npm run dev
```

默认将 `/api` 代理到 `http://127.0.0.1:8088`。

官网是静态页面，可以单独预览：

```sh
cd frontend/index
python3 -m http.server 8080
```

单独预览官网时，Caddy 路由和安装包下载链路并不存在。完整联调应使用前端 Docker 镜像。

## 7. 开发和构建 Android

在 Android Studio 中打开项目根目录下的 `android` 文件夹，等待 Gradle 同步完成。连接手机后执行：

```sh
cd android
sh ./gradlew assembleDebug
sh ./gradlew assembleRelease
```

输出位置：

- Debug：`android/app/build/outputs/apk/debug/app-debug.apk`
- Release：`android/app/build/outputs/apk/release/app-release.apk`

Release 构建必须存在：

- `android/keystore.properties`
- `android/keystore/automatictools-release.jks`

每次公开发布前都要递增 `android/app/build.gradle.kts` 中的 `versionCode`。开发版和正式版签名不同，首次从 Debug 切换到 Release 时可能需要卸载旧版；之后必须始终使用同一正式签名密钥。

检查真机连接：

```sh
~/Library/Android/sdk/platform-tools/adb devices
```

## 8. Windows 客户端限制

Windows 客户端使用 Tkinter、Windows DPAPI 和 Win32 点击接口。macOS 可以编辑源码，但不能可靠运行或生成正式 EXE。需要以下任一方案：

1. 保留一台 Windows 构建机。
2. 在 Mac 上使用 Windows 虚拟机。
3. 后续增加 Windows CI 构建流程。

发布前需要将 Windows 构建机生成的 `AutomaticTools.exe` 放入：

```text
windows/dist/AutomaticTools.exe
```

## 9. 构建并发布 Docker 镜像

登录 Docker Hub：

```sh
docker login
```

构建、推送后端：

```sh
sh backend/build.sh
sh backend/push.sh
```

先构建 Android Release APK，并恢复 Windows 正式 EXE，再构建、推送前端：

```sh
cd android
sh ./gradlew assembleRelease
cd ..
sh frontend/build.sh
sh frontend/push.sh
```

脚本统一生成 `linux/amd64` 镜像，以匹配当前生产服务器。Apple Silicon 会通过 Buildx 模拟 amd64，构建速度可能比本机架构慢。

镜像推送完成后，在服务器执行：

```sh
cd ~/prod
./update.sh
```

## 10. 迁移验收清单

- `go test ./...` 在 `backend` 目录通过
- 后端 `/health` 返回 `{"ok":true}`
- `npm ci && npm run build` 在 `frontend/admin` 目录通过
- Android Debug APK 可以安装并运行
- Android Release APK 使用原签名成功构建
- `windows/dist/AutomaticTools.exe` 已单独恢复
- Docker 前后端镜像均能构建
- `prod/.env` 和密钥文件已加密备份，且没有被 Git 跟踪

## 官方安装资料

- [Android Studio for macOS](https://developer.android.com/studio/install.html)
- [Docker Desktop for Mac](https://docs.docker.com/desktop/setup/install/mac-install/)
- [Go 安装说明](https://go.dev/doc/install)
- [Homebrew PostgreSQL 17](https://formulae.brew.sh/formula/postgresql@17)
