# AutomaticTools Frontend

前端由两个独立项目组成，但统一构建为一个 Nginx 镜像：

- `index`：静态官网和客户端下载页
- `admin`：Vue 3 管理端
- `downloads`：构建镜像时使用的 Windows 和 Android 安装包

## 管理端开发

```sh
cd frontend/admin
npm ci
npm run dev
```

生产构建检查：

```sh
npm run build
```

## 准备客户端下载包

构建前端镜像前必须存在：

- `windows/dist/AutomaticTools.exe`
- `android/app/build/outputs/apk/release/app-release.apk`

Windows：

```powershell
.\frontend\prepare-downloads.bat
```

macOS：

```sh
sh frontend/prepare-downloads.sh
```

脚本会复制并重命名安装包到 `frontend/downloads`。安装包被 Git 忽略，不会随源码仓库迁移。

## 构建和推送镜像

Windows：

```powershell
.\frontend\build.bat
.\frontend\push.bat
```

macOS：

```sh
sh frontend/build.sh
sh frontend/push.sh
```

默认镜像为 `wcs19890321/automatictools-frontend:latest`，目标平台为生产服务器使用的 `linux/amd64`。
