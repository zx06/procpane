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

Signed & notarized releases would normally skip this step.
