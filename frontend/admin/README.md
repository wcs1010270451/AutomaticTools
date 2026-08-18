# AutomaticTools Admin

AutomaticTools 管理端，使用 Vue 3、TypeScript、Vite、Vue Router、Pinia 和 Element Plus。

## 本地开发

```sh
cd frontend/admin
npm ci
npm run dev
```

本地开发默认将 `/api` 代理到 `http://127.0.0.1:8088`。需要连接其他后端时：

```powershell
$env:VITE_API_PROXY_TARGET="http://127.0.0.1:8089"
npm run dev
```

macOS/Linux：

```sh
VITE_API_PROXY_TARGET=http://127.0.0.1:8089 npm run dev
```

未登录访问管理端会跳转到 `/login`。管理员登录调用
`POST /api/admin/auth/login`，成功后保存管理员令牌并返回管理端首页。
登录后可在 `/admins` 查看、新增、修改和删除管理员账号。
在 `/users` 可分页查看用户，并按用户名、邮箱、手机号或账号状态筛选。
在 `/license-codes` 可按工具批量生成授权码，查看库存、兑换状态和兑换用户，
并作废尚未使用的授权码。授权码明文只在生成成功时显示一次。

## 检查生产构建

```powershell
npm run build
```

Windows 正式发布由 `frontend/build.bat` 完成，macOS 使用
`sh frontend/build.sh`。管理端构建结果与官网一起放入
`automatictools-frontend` 镜像，不单独运行容器。

完整的 macOS 环境配置见 [macOS 开发环境与迁移指南](../../docs/macos-development.md)。
