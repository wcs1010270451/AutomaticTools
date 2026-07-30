# AutomaticTools Admin

AutomaticTools 管理端，使用 Vue 3、TypeScript、Vite、Vue Router、Pinia 和 Element Plus。

## 本地开发

```powershell
cd frontend\admin
npm install
npm run dev
```

## 检查生产构建

```powershell
npm run build
```

正式发布由 `frontend/build.bat` 统一完成。管理端构建结果与官网一起放入 `automatictools-frontend` 镜像，不单独运行容器。
