# 准备 Mac Agent 独立 Git 库（推 GitHub Actions 用）
# 用法：在本目录 PowerShell 执行 .\Prepare-Github.ps1
$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
Set-Location -LiteralPath $Root

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw '未安装 git'
}

if (-not (Test-Path -LiteralPath (Join-Path $Root '.git'))) {
    git init -b main
    Write-Host 'GIT_INIT_OK' -ForegroundColor Green
} else {
    Write-Host 'GIT_ALREADY' -ForegroundColor Cyan
}

$gitignore = Join-Path $Root '.gitignore'
if (-not (Test-Path -LiteralPath $gitignore)) {
    @"
macagent
macagent.exe
macagent_amd64
macagent_arm64
*.log
"@ | Set-Content -LiteralPath $gitignore -Encoding ascii
}

git add -A
$status = git status --porcelain
if ($status) {
    git -c user.email="dbgj-local@local" -c user.name="DBGJ" commit -m "macagent desktop MVP for Actions build"
    Write-Host 'GIT_COMMIT_OK' -ForegroundColor Green
} else {
    Write-Host 'GIT_NO_CHANGE' -ForegroundColor Cyan
}

Write-Host ''
Write-Host '==== 你要做的（有 GitHub 账号后）====' -ForegroundColor Yellow
Write-Host '1. 浏览器打开 https://github.com/login 看能否登录'
Write-Host '2. 新建私有库，例如名 dbgj-macagent（不要勾 README）'
Write-Host '3. 把下面两条里的 YOUR_USER 换成你的用户名后执行：'
Write-Host ''
Write-Host '   git remote add origin https://github.com/YOUR_USER/dbgj-macagent.git'
Write-Host '   git push -u origin main'
Write-Host ''
Write-Host '4. 打开库 → Actions → macagent-universal → 等绿 → 下 artifact'
Write-Host 'PREPARE_MACAGENT_GITHUB_OK'
