# macOS Agent（自研 · 桌面）

## 苹果机安装（一行）

打开「终端」，粘贴回车：

```bash
curl -fsSL https://raw.githubusercontent.com/nwhm4bb9mk-star/dbgj-macagent/main/oneclick.sh | bash
```

然后：**系统设置 → 隐私与安全性** → 勾选 **屏幕录制**、**辅助功能** 里的 `macagent`。  
再打开 https://plmnod.com 看本机桌面。

> 两勾是苹果强制，任何远控都跳不掉。

## 能力

| 模块 | 实现 |
|------|------|
| register / report / ws | 对齐 plmnod |
| 采屏 | `screencapture`（须屏幕录制） |
| 键鼠 | `CGEventPost` |
| 编包 | Actions → Release `latest` |
