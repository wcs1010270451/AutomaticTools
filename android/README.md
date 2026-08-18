# AutomaticTools Android

原生 Kotlin Android 客户端，使用无障碍服务执行点击手势，无需 Root。

## 当前功能

- 用户注册、登录和登录状态保存
- 工具列表与授权码兑换
- 可拖动的点击目标点和悬浮控制面板
- 锁定目标坐标、开始和停止自动点击
- 100ms、500ms、1000ms、2 秒间隔
- 点击次数统计和重置

## Android Studio 运行

在 Android Studio 中打开仓库内的 `android` 文件夹，等待 Gradle 同步完成。选择已启用 USB 调试的手机，然后点击 Run 或 Debug。

首次使用悬浮工具时，需要授予：

1. 显示在其他应用上层权限。
2. AutomaticTools 无障碍服务权限。

随后拖动目标点，锁定坐标，选择间隔并开始点击。部分游戏、支付、银行或带反自动化保护的应用可能拦截无障碍手势。

## 命令行构建

macOS：

```sh
cd android
sh ./gradlew assembleDebug
sh ./gradlew assembleRelease
```

Windows PowerShell：

```powershell
cd android
.\gradlew.bat assembleDebug
.\gradlew.bat assembleRelease
```

输出文件：

- Debug：`app/build/outputs/apk/debug/app-debug.apk`
- Release：`app/build/outputs/apk/release/app-release.apk`

## 正式版签名

Release 构建依赖以下本地文件：

- `android/keystore.properties`
- `android/keystore/automatictools-release.jks`

两者都被 Git 忽略，换电脑时必须单独安全迁移并共同备份。丢失密钥后，新 APK 无法覆盖升级已经安装的正式版。

开发版和正式版使用不同签名。首次由 Debug 切换到 Release 时可能需要卸载旧应用。此后必须一直使用同一正式密钥，并在每次公开发布前递增 `app/build.gradle.kts` 中的 `versionCode`。

完整的 macOS 环境配置见 [macOS 开发环境与迁移指南](../docs/macos-development.md)。
