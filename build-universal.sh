#!/bin/bash
# 在 macOS 上编 Universal 包（Intel+Apple Silicon）
set -euo pipefail
cd "$(dirname "$0")"
export CGO_ENABLED=1
go mod tidy
GOOS=darwin GOARCH=amd64 go build -o macagent_amd64 .
GOOS=darwin GOARCH=arm64 go build -o macagent_arm64 .
lipo -create -output macagent macagent_amd64 macagent_arm64
rm -f macagent_amd64 macagent_arm64
echo "BUILD_MAC_UNIVERSAL_OK $(ls -la macagent)"
