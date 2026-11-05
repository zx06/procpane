procpane — macOS 安装说明
========================

若首次打开时提示「已损坏，无法打开」或无法运行，应用本身通常并未损坏。
这是 macOS Gatekeeper 对未签名、未公证的下载附加的隔离（quarantine）标记。

安装步骤
--------

1. 将 procpane.app 拖入「应用程序」(Applications) 文件夹。
2. 打开「终端」(Terminal)，执行（路径按实际安装位置修改）：

   xattr -cr /Applications/procpane.app

3. 从「应用程序」或 Launchpad 正常打开 procpane。
4. 若仍被拦截：在 procpane.app 上按住 Control 键并点按 → 选择「打开」→ 再点「打开」一次。

本地网络（SSH / Docker 访问局域网主机）
----------------------------------------

若菜单栏 GUI 连接局域网主机（如 NAS 10.0.0.x）失败，提示「no route to host」或无法连接，
而同一台 Mac 上终端里的 ssh 或 CLI `procpane ssh ls …` 正常，请检查：

  系统设置 → 隐私与安全性 → 本地网络 → 打开 procpane 的开关。

（macOS Sequoia 及更新版本对 GUI 应用访问局域网有单独权限；安装新版 procpane 后首次连 LAN 时也可能弹出授权提示。）

说明
----

当前发布版本未使用 Apple Developer ID 签名与公证（notarize）。
未来若提供已签名、已公证的构建，一般无需上述 xattr 步骤。

---
English (brief)

If macOS says the app is “damaged”, it is usually Gatekeeper quarantine on an unsigned download—not file corruption.

Install: drag procpane.app to Applications → run:

  xattr -cr /Applications/procpane.app

Then open the app. If blocked, Control-click the app → Open once.

Local Network: if the tray GUI cannot reach LAN hosts (e.g. NAS) with “no route to host” while Terminal/CLI works, enable procpane under System Settings → Privacy & Security → Local Network.

Signed & notarized releases would normally skip the xattr step.

登录时自动启动（设置 → 登录时启动）
------------------------------------

在 procpane 设置中打开「登录时启动」后，系统会在
~/Library/LaunchAgents/com.zx06.procpane.plist 注册启动项，登录项中应显示 procpane（而非其他进程名）。

若您曾使用 v0.2.2 并看到错误的「swift-frontend」登录项：请先在
系统设置 → 通用 → 登录项 中删除「swift-frontend」，再在 procpane 中关闭并重新打开「登录时启动」。

---
Autostart (Settings → launch at login)

Enabling autostart installs ~/Library/LaunchAgents/com.zx06.procpane.plist pointing at your
procpane.app binary; Login Items should list procpane.

If v0.2.2 left a stray "swift-frontend" login item: remove it under System Settings → General →
Login Items, then toggle autostart off and on again in procpane.
