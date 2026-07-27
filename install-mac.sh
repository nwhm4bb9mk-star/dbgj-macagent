#!/bin/bash
# DBGJ Mac Agent 一键安装（真机执行）
set -euo pipefail
SERVER_URL="${SERVER_URL:-https://plmnod.com}"
TOKEN="${TOKEN:-}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/Library/Application Support/DBGJ/macagent}"
BIN_NAME="macagent"
PLIST="$HOME/Library/LaunchAgents/com.dbgj.macagent.plist"

echo "[1/4] 安装目录 $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

SRC_BIN="${1:-}"
if [ -z "$SRC_BIN" ] || [ ! -f "$SRC_BIN" ]; then
  echo "用法: SERVER_URL=https://plmnod.com $0 /path/to/macagent"
  exit 1
fi

echo "[2/4] 复制二进制"
cp -f "$SRC_BIN" "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"

echo "[3/4] LaunchAgent"
{
  echo '<?xml version="1.0" encoding="UTF-8"?>'
  echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">'
  echo '<plist version="1.0"><dict>'
  echo '  <key>Label</key><string>com.dbgj.macagent</string>'
  echo '  <key>ProgramArguments</key><array>'
  echo "    <string>$INSTALL_DIR/$BIN_NAME</string>"
  echo '    <string>-server</string>'
  echo "    <string>$SERVER_URL</string>"
  if [ -n "$TOKEN" ]; then
    echo '    <string>-token</string>'
    echo "    <string>$TOKEN</string>"
  fi
  echo '  </array>'
  echo '  <key>RunAtLoad</key><true/>'
  echo '  <key>KeepAlive</key><true/>'
  echo "  <key>WorkingDirectory</key><string>$INSTALL_DIR</string>"
  echo "  <key>StandardOutPath</key><string>$INSTALL_DIR/agent.out.log</string>"
  echo "  <key>StandardErrorPath</key><string>$INSTALL_DIR/agent.err.log</string>"
  echo '</dict></plist>'
} > "$PLIST"

launchctl unload "$PLIST" 2>/dev/null || true
launchctl load "$PLIST"
launchctl start com.dbgj.macagent 2>/dev/null || true

echo "[4/4] 请到 系统设置→隐私与安全性 勾选 屏幕录制 + 辅助功能（macagent）"
echo "MAC_AGENT_INSTALL_OK"
echo "日志: tail -f \"$INSTALL_DIR/agent.err.log\""
