# 📦 安装指南

## 🎯 系统要求

### 必需
- **操作系统**: Windows 10/11 (推荐) / Linux / Mac
- **Go 版本**: >= 1.21
- **磁盘空间**: >= 500 MB

### Windows 平台额外要求
- **MinGW-w64**: Fyne 和 SQLite 需要 CGO 支持
- **显卡驱动**: Fyne 需要 OpenGL 支持（通常已安装）

### server-active 额外要求
- **MySQL**: >= 8.0

---

## 🚀 快速安装（首次构建）

### Step 1: 安装依赖

#### Windows

**安装 MinGW-w64（CGO 必需）**:

**方法 1: 使用 Chocolatey（推荐）**
```bash
# 安装 Chocolatey (如果未安装)
# 以管理员身份运行 PowerShell，执行：
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 安装 MinGW
choco install mingw
```

**方法 2: 手动下载**
1. 下载：https://sourceforge.net/projects/mingw-w64/
2. 安装到 `C:\mingw-w64`
3. 添加到 PATH：
   - 打开"系统环境变量"
   - 编辑 `Path`
   - 添加 `C:\mingw-w64\bin`
4. 重启命令行窗口

**验证安装**:
```bash
gcc --version
```

#### Linux

```bash
# Ubuntu/Debian (Fyne 依赖)
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install -y gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel

# Arch
sudo pacman -S base-devel libx11 libxcursor libxrandr libxinerama mesa libxi libxxf86vm
```

#### Mac

```bash
# 安装 Xcode Command Line Tools
xcode-select --install
```

---

### Step 2: 下载依赖包

```bash
cd server-go
go mod download
go mod tidy

cd ../server-active
go mod download
go mod tidy
```

**国内用户加速（可选）**:
```bash
# 设置 GOPROXY
go env -w GOPROXY=https://goproxy.cn,direct
```

---

### Step 3: 构建所有组件

#### Windows

```bash
# Fyne GUI 版本（推荐）
BUILD_WITH_FYNE.bat

# 或使用交互式脚本
QUICK_START.bat
```

**构建顺序**:
1. 打包 browser-monitor → `server-go/assets/browser-monitor.zip`
2. 下载 server-go 依赖
3. 编译 server-go → `dy-live-monitor.exe`
4. 下载 server-active 依赖
5. 编译 server-active → `dy-live-license-server.exe`

#### Linux/Mac

```bash
chmod +x BUILD_ALL.sh
./BUILD_ALL.sh
```

---

## 🐛 常见安装问题

### 问题 1: `gcc: command not found`

**原因**: 未安装 MinGW-w64（Windows）或 GCC（Linux/Mac）

**解决方案**:
- Windows: 安装 MinGW-w64（见 Step 1）
- Linux: `sudo apt-get install gcc`
- Mac: `xcode-select --install`

---

### 问题 2: `missing go.sum entry`

**原因**: 缺少 Go 依赖包

**解决方案**:
```bash
cd server-go
go mod tidy

cd ../server-active
go mod tidy
```

---

### 问题 3: `pattern assets/*: no matching files found`

**原因**: browser-monitor 还未打包

**解决方案**:
```bash
cd browser-monitor
pack.bat  # Windows
# 或
./pack.sh  # Linux/Mac

# 然后重新编译 server-go
cd ../server-go
build.bat
```

---

### 问题 4: `cgo: C compiler "gcc" not found`

**原因**: CGO 找不到 GCC 编译器

**解决方案**:

**Windows**:
1. 确认 MinGW-w64 已安装
2. 确认 `C:\mingw-w64\bin` 已添加到 PATH
3. 重启命令行窗口
4. 验证：`gcc --version`

**临时禁用 CGO（不推荐）**:
```bash
set CGO_ENABLED=0
go build
```

---

### 问题 5: `go: downloading ... connection timed out`

**原因**: 网络问题，无法下载依赖

**解决方案**:

**国内用户**:
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

**使用代理**:
```bash
set HTTP_PROXY=http://127.0.0.1:7890
set HTTPS_PROXY=http://127.0.0.1:7890
```

---

### 问题 6: WebView2 Runtime 未安装

**原因**: Windows 平台缺少 WebView2 Runtime

**解决方案**:

**自动安装（推荐）**:
```bash
# 启动 server-go，程序会提示自动安装
dy-live-monitor.exe
# 输入 y 即可自动下载并安装
```

**手动安装**:
1. 下载：https://developer.microsoft.com/en-us/microsoft-edge/webview2/
2. 选择 "Evergreen Standalone Installer"
3. 运行安装程序

---

### 问题 7: MySQL 连接失败

**原因**: server-active 无法连接 MySQL

**解决方案**:
1. 确认 MySQL 已启动
2. 检查 `config.json` 配置：
   ```json
   {
     "database": {
       "host": "localhost",
       "port": "3306",
       "user": "root",
       "password": "your_password",
       "database": "dy_license"
     }
   }
   ```
3. 创建数据库：
   ```sql
   CREATE DATABASE dy_license CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

---

## 📝 构建验证

### 检查构建产物

```bash
# Windows
dir server-go\dy-live-monitor.exe
dir server-go\assets\browser-monitor.zip
dir server-active\dy-live-license-server.exe

# Linux/Mac
ls -lh server-go/dy-live-monitor
ls -lh server-go/assets/browser-monitor.zip
ls -lh server-active/dy-live-license-server
```

### 验证可执行文件

```bash
# Windows
cd server-go
dy-live-monitor.exe --version  # 如果支持

# Linux/Mac
cd server-go
./dy-live-monitor --version
```

---

## 🔧 开发环境配置

### VSCode 推荐插件

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-typescript-next",
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode"
  ]
}
```

### VSCode 设置

```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.gopath": "${workspaceFolder}",
  "go.goroot": "/usr/local/go",
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint"
}
```

---

## 🚀 生产环境部署

### 1. 准备服务器

**系统要求**:
- Windows Server 2016+ / Ubuntu 20.04+ / CentOS 8+
- 2 Core CPU, 4 GB RAM, 20 GB Disk

### 2. 部署 server-active

```bash
# 1. 上传文件
scp dy-live-license-server user@server:/opt/dy-live/
scp config.json user@server:/opt/dy-live/
scp -r keys/ user@server:/opt/dy-live/

# 2. 配置 systemd (Linux)
sudo cat > /etc/systemd/system/dy-license.service <<EOF
[Unit]
Description=Douyin Live License Service
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/dy-live
ExecStart=/opt/dy-live/dy-live-license-server
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 3. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable dy-license
sudo systemctl start dy-license
sudo systemctl status dy-license
```

### 3. 配置 Nginx (HTTPS)

```nginx
server {
    listen 443 ssl http2;
    server_name license.example.com;

    ssl_certificate /etc/letsencrypt/live/license.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/license.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 📞 技术支持

如遇到其他问题，请查看：
- **QUICK_START.md** - 快速开始指南
- **COMPLETION_REPORT.md** - 完整功能报告
- **server-go/README.md** - 后端服务文档
- **server-active/README.md** - 许可证服务文档

或访问项目 GitHub:
- https://github.com/WanGuChou/dy-live-record

---

**最后更新**: 2025-11-15  
**版本**: v3.1.0
