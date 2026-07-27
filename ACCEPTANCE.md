# 真机验收清单（Mac）

## 编包（任选）

```bash
cd …/data/agent-bin/mac
bash build-universal.sh
# → BUILD_MAC_UNIVERSAL_OK
```

或 GitHub Actions 下载 artifact `macagent-universal`。

## 安装

```bash
chmod +x install-mac.sh macagent
SERVER_URL=https://plmnod.com bash install-mac.sh ./macagent
# → MAC_AGENT_INSTALL_OK
```

## 权限（必须）

系统设置 → 隐私与安全性：

1. **屏幕录制** → 勾选 `macagent`（或先用终端跑时勾 Terminal）
2. **辅助功能** → 勾选 `macagent`

## 管台

1. 打开 `https://plmnod.com` 登录  
2. 列表出现本机（`os=darwin`）且在线  
3. 开桌面：应见真画面（非纯色占位）  
4. 鼠标点击/拖动可用  

验收口令：`MAC_AGENT_DESKTOP_MVP_OK`

## 卸载

```bash
launchctl unload ~/Library/LaunchAgents/com.dbgj.macagent.plist
rm -f ~/Library/LaunchAgents/com.dbgj.macagent.plist
rm -rf "$HOME/Library/Application Support/DBGJ/macagent"
```
