# macOS Agent（自研 · 桌面）

> 立项 §17 · **禁止**整包嵌入第三方远控

## 能力

| 模块 | 实现 |
|------|------|
| register / report / ws | 对齐 plmnod |
| 采屏 | `screencapture`（macOS 15+ 无 CGDisplayCreateImage；须屏幕录制） |
| 键鼠 | `CGEventPost`；键盘兼容前端 Windows VK |
| 安装 | `install-mac.sh` + LaunchAgent |
| 编包 | `build-universal.sh` / `.github/workflows/macagent-universal.yml` |

## 真机步骤

见 [ACCEPTANCE.md](./ACCEPTANCE.md)。

本仓库 Windows 机可 `go build` 协议面（无 CG）；**真采屏必须在 macOS 上 `CGO_ENABLED=1` 编译**。
