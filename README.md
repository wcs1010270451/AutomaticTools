# AutomaticTools

AutomaticTools 是一个实用工具平台，目前包含自动点击工具，以及用户、订单、支付宝支付和工具购买记录相关的基础服务。

## 项目目录

- `backend`：Gin + GORM API，按 router、handler、logic、middleware、store 分层
- `frontend/index`：静态官网
- `frontend/admin`：Vue 3 + TypeScript 管理端
- `frontend` 根目录：官网与管理端的统一 Nginx 镜像构建配置
- `android`：Android 客户端
- `windows`：Windows 客户端
- `prod`：Docker Compose、Caddy 和生产环境运维脚本

官网和管理端源码相互独立，但会一起打包到 `automatictools-frontend` 镜像，在同一个 Nginx 容器中运行。

## 发布前后端镜像

先确认 Docker Desktop 已启动并已经登录 Docker Hub：

```powershell
docker login
```

前端先构建，成功后再推送：

```powershell
.\frontend\build.bat
.\frontend\push.bat
```

后端先构建，成功后再推送：

```powershell
.\backend\build.bat
.\backend\push.bat
```

发布的镜像均为 `linux/amd64`：

- `wcs19890321/automatictools-frontend:latest`
- `wcs19890321/automatictools-backend:latest`

构建与推送分开执行。上传失败时可以直接重新运行对应的 `push.bat`，不需要再次构建。

## 服务器首次部署

将完整的 `prod` 目录上传到服务器，必须包含隐藏文件 `prod/.env`。服务器需要安装 Docker Engine 和 Docker Compose v2，并开放 TCP 80、TCP 443；UDP 443 用于 HTTP/3，可选开放。

```sh
cd ~/prod
chmod +x start.sh update.sh stop.sh
./start.sh
```

`start.sh` 会拉取 PostgreSQL、后端、前端和 Caddy 镜像，启动全部容器并等待健康检查通过。服务器不会编译源码或构建镜像。

当前正式域名：

```text
https://autumnwind.top
https://admin.autumnwind.top
```

`prod/.env` 中的域名配置应为：

```env
DOMAIN=autumnwind.top
ADMIN_DOMAIN=admin.autumnwind.top
```

Caddy 会自动申请和续期 HTTPS 证书。

## 后续更新

本地完成前后端镜像构建与推送后，在服务器执行：

```sh
cd ~/prod
./update.sh
```

停止服务：

```sh
cd ~/prod
./stop.sh
```

`stop.sh` 不会删除 PostgreSQL 和 Caddy 数据卷。

## 部署验证

```sh
docker compose ps
curl https://autumnwind.top/health
curl https://autumnwind.top/api/tools
curl -I https://admin.autumnwind.top
```

正常响应示例：

```json
{"ok":true}
```

后端的 `8088` 和 PostgreSQL 的 `5432` 仅在 Docker 内部网络使用，不对服务器公网开放。外部请求统一通过 Caddy 的 80/443 端口访问。

详细部署与排查说明见 [`prod/README.md`](prod/README.md)。
