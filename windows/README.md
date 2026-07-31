# AutomaticTools Windows

Windows 客户端使用 Python 标准库和 Tkinter，不需要额外运行依赖。

## 本地运行

先启动本地后端，再双击：

```text
run_automatic_tools.bat
```

该脚本会将 API 地址设置为 `http://127.0.0.1:8088`。直接运行
`automatic_tools_gui.py` 或打包后的 EXE 时，默认连接：

```text
https://autumnwind.top
```

也可以通过环境变量覆盖 API 地址：

```powershell
$env:AUTOMATIC_TOOLS_API_BASE_URL="https://example.com"
python .\automatic_tools_gui.py
```

## 登录状态

勾选“记住登录状态”后，JWT 使用 Windows DPAPI 加密并保存在：

```text
%APPDATA%\AutomaticTools\session.dat
```

客户端首次启动时生成独立的设备 ID，保存在同目录的 `client.json`。
退出登录会清除登录令牌，但保留设备 ID。

## 测试

```powershell
cd windows
python -m unittest -v test_auth_client.py
python -m py_compile automatic_tools_gui.py auth_client.py
```

## 打包

安装 PyInstaller 后执行：

```powershell
cd windows
python -m PyInstaller --noconfirm --clean AutomaticTools.spec
```

生成文件位于 `windows\dist\AutomaticTools.exe`。

本地联调版本默认连接 `http://127.0.0.1:8088`：

```powershell
python -m PyInstaller --noconfirm --clean AutomaticToolsLocal.spec
```

生成文件位于 `windows\dist\AutomaticToolsLocal.exe`。本地版本仅用于开发测试，
不要作为正式安装包发布。
