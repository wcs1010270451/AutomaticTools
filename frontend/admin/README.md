# AutomaticTools Admin

AutomaticTools 管理端，使用 Vue 3、TypeScript、Vite、Vue Router、Pinia 和 Element Plus。

## 本地开发

```powershell
cd frontend\admin
npm install
npm run dev
```

本地开发默认将 `/api` 代理到 `http://127.0.0.1:8088`。需要连接其他后端时：

```powershell
$env:VITE_API_PROXY_TARGET="http://127.0.0.1:8089"
npm run dev
```

未登录访问管理端会跳转到 `/login`。管理员登录调用
`POST /api/admin/auth/login`，成功后保存管理员令牌并返回管理端首页。
登录后可在 `/admins` 查看、新增、修改和删除管理员账号。
在 `/users` 可分页查看用户，并按用户名、邮箱、手机号或账号状态筛选。

## 检查生产构建

```powershell
npm run build
```

正式发布由 `frontend/build.bat` 统一完成。管理端构建结果与官网一起放入 `automatictools-frontend` 镜像，不单独运行容器。
