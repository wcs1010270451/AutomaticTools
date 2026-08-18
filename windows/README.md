# AutomaticTools Windows

Windows 客户端使用 Tkinter 构建界面，通过后端完成注册、登录和授权码兑换。
打包后的 EXE 已包含运行依赖，目标电脑不需要安装 Python。

## 本地运行

先启动本地后端，再双击：

```text
run_automatic_tools.bat
```

直接从源码运行或打包前，先安装开发依赖：

```powershell
python -m pip install -r requirements.txt
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

## 用户主界面

登录后进入统一的工具平台主界面，包含首页、工具中心和账户信息。已经开通的工具
显示在“工具中心”下一级导航中。客户端会同步用户资料、工具列表和开通记录；
未开通的工具可输入一次性授权码兑换，兑换成功后立即加入已开通列表。

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

## 在 macOS 上开发

客户端依赖 Windows DPAPI 和 Win32 点击接口。macOS 可以编辑源码，但不能直接
运行或可靠生成正式 EXE。请保留 Windows 构建机、使用 Windows 虚拟机，或后续
配置 Windows CI。

`windows/dist/` 被 Git 忽略。换电脑时应单独备份正式版，并在构建前端镜像前放到：

```text
windows/dist/AutomaticTools.exe
```

完整迁移步骤见 [macOS 开发环境与迁移指南](../docs/macos-development.md)。
