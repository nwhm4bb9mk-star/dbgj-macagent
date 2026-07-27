#!/bin/bash
# 苹果机：终端粘贴一行即可
# curl -fsSL https://raw.githubusercontent.com/nwhm4bb9mk-star/dbgj-macagent/main/oneclick.sh | bash
set -euo pipefail
SERVER_URL="${SERVER_URL:-https://plmnod.com}"
REPO="nwhm4bb9mk-star/dbgj-macagent"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"
echo "[1/3] 下载 macagent"
curl -fsSL -o macagent "https://github.com/${REPO}/releases/latest/download/macagent"
curl -fsSL -o install-mac.sh "https://raw.githubusercontent.com/${REPO}/main/install-mac.sh"
chmod +x macagent install-mac.sh
echo "[2/3] 安装"
SERVER_URL="$SERVER_URL" bash install-mac.sh ./macagent
echo "[3/3] 系统设置→隐私与安全性：勾选「屏幕录制」「辅助功能」里的 macagent，然后打开 https://plmnod.com"
