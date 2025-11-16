# 🚀 快速开始指南

## 📋 系统要求

- **操作系统**: Windows 10/11 (推荐) / Linux / Mac
- **Go 版本**: >= 1.21
- **WebView2 Runtime**: Windows 平台必需（程序会自动检测并安装）
- **CGO**: SQLite 需要 CGO 支持（Windows 需要 MinGW-w64）
- **MySQL**: server-active 需要 MySQL 8.0+

---

## 🎯 三步快速启动

### 第一步：启动核心后端服务 (server-go)

```bash
cd server-go

# Windows
build.bat
dy-live-monitor.exe

# Linux/Mac
go build -o dy-live-monitor .
./dy-live-monitor
```

**首次启动时会发生什么？**

1. ✅ **自动检查依赖**
   - WebView2 Runtime
   - SQLite 驱动 (CGO)
   - 网络连接
   - 磁盘空间

2. ⚠️ **如果缺少 WebView2**
   ```
   ⚠️  检测到关键依赖缺失
   是否尝试自动安装 WebView2? (y/n): 
   ```
   - 输入 `y` → 自动下载并安装
   - 输入 `n` → 程序退出，需手动安装

3. 🔐 **许可证激活**（首次）
   - 程序会提示输入许可证
   - 从 `server-active` 获取许可证字符串并粘贴
   - （临时跳过：注释掉 `main.go` 中的许可证检查代码）

4. ✅ **启动成功**
   - 系统托盘显示图标
   - WebSocket 服务器监听 `:8090`
   - SQLite 数据库自动初始化

---

### 第二步：安装浏览器插件 (browser-monitor)

#### 方法 A: 手动安装（推荐）

1. 打开 Chrome 或 Edge 浏览器
2. 访问 `chrome://extensions/` （Edge 用户访问 `edge://extensions/`）
3. **启用右上角的「开发者模式」**
4. 点击「加载已解压的扩展程序」
5. 选择 `/workspace/browser-monitor` 目录
6. ✅ 安装完成

#### 方法 B: 通过 server-go 安装

1. 先打包插件：
   ```bash
   cd browser-monitor
   pack.bat  # Windows
   # 或
   ./pack.sh  # Linux/Mac
   ```

2. 在 `server-go` 设置界面点击「安装插件」
3. 程序会自动解压到临时目录
4. 按照提示在浏览器中加载

#### 配置插件

1. 点击插件图标（浏览器工具栏）
2. 设置 WebSocket 地址：`ws://localhost:8090/ws`
3. 保存配置

---

### 第三步：打开抖音直播间测试

1. 访问任意抖音直播间：
   ```
   https://live.douyin.com/123456789
   ```

2. **在 server-go 主界面查看数据**：
   - 左侧：房间列表（自动显示当前直播间）
   - 右侧：
     - 📊 数据概览：礼物总值、消息数
     - 🎁 礼物记录：实时更新
     - 💬 消息记录：聊天、进场、关注
     - 📈 分段记分：创建 PK 时段统计
     - 👤 主播管理：添加主播、绑定礼物

3. **测试分段记分**：
   - 切换到「📈 分段记分」标签
   - 输入分段名称：`PK 第一轮`
   - 点击「创建新分段」
   - 等待一段时间后，点击「结束当前分段」
   - 查看统计结果（礼物总值、消息数）

---

## 🔑 （可选）启动许可证服务 (server-active)

如果需要许可证功能，按以下步骤启动：

### 1. 生成 RSA 密钥对

```bash
cd server-active
mkdir keys

# 生成私钥
openssl genrsa -out keys/private.pem 2048

# 生成公钥
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

### 2. 配置数据库

复制示例配置：
```bash
cp config.example.json config.json
```

编辑 `config.json`：
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": "8080"
  },
  "database": {
    "host": "localhost",
    "port": "3306",
    "user": "root",
    "password": "your_password",
    "database": "dy_license"
  },
  "license": {
    "private_key_path": "./keys/private.pem",
    "public_key_path": "./keys/public.pem"
  }
}
```

### 3. 创建数据库

```bash
mysql -u root -p

CREATE DATABASE dy_license CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXIT;
```

### 4. 启动服务

```bash
# Windows
build.bat
dy-live-license-server.exe

# Linux/Mac
go build -o dy-live-license-server .
./dy-live-license-server
```

### 5. 访问管理后台

打开浏览器访问：
```
http://localhost:8080/admin
```

**管理后台功能**：
- 📝 生成新许可证
- 📋 查看许可证列表
- 👁️ 查看许可证详情
- ❌ 撤销许可证

---

## 🎨 使用场景示例

### 场景 1: 监控单个直播间

1. 启动 `server-go`
2. 打开抖音直播间 `https://live.douyin.com/123456789`
3. 在主界面查看实时数据
4. 礼物和消息会自动记录到数据库

### 场景 2: 多房间同时监控

1. 打开多个浏览器标签页
2. 每个标签页访问不同的直播间
3. 主界面会自动创建多个房间 Tab
4. 点击左侧房间列表切换查看

### 场景 3: PK 时段统计

1. 直播开始 → 创建分段「主播独播」
2. PK 开始 → 创建分段「PK 第一轮」（自动结束「主播独播」）
3. PK 结束 → 点击「结束当前分段」
4. 查看「PK 第一轮」的礼物总值和消息数

### 场景 4: 主播业绩分配

1. 在「👤 主播管理」中添加主播：
   - 主播 ID: `anchor_A`
   - 主播名称: `主播A`
   - 绑定礼物: `玫瑰花,跑车,火箭`

2. 观众送礼物时：
   - 如果礼物是「玫瑰花」→ 自动计入主播A业绩
   - 如果消息内容包含「@主播A」或「送给主播A」→ 自动计入主播A业绩

3. 查看业绩：
   - 数据库 `anchor_performance` 表

---

## 🐛 常见问题

### Q1: 启动时提示「未找到有效许可证」

**A**: 临时跳过许可证检查：

编辑 `server-go/main.go`，注释掉以下代码：
```go
// 3. 许可证校验（强制）
/*
licenseManager := license.NewManager(...)
...
*/
```

### Q2: 启动时提示「CGO 未启用」

**A**: 安装 MinGW-w64（Windows）：

```bash
# 方法 1: Chocolatey
choco install mingw

# 方法 2: 手动下载
# https://sourceforge.net/projects/mingw-w64/
# 安装后添加到 PATH: C:\mingw-w64\bin
```

### Q3: 插件连接不上 server-go

**A**: 检查：
1. `server-go` 是否正在运行（系统托盘有图标）
2. WebSocket 地址是否正确（`ws://localhost:8090/ws`）
3. 浏览器控制台是否有错误（F12）

### Q4: 主界面不显示数据

**A**: 
1. 确保插件已连接（查看插件 popup 界面）
2. 打开抖音直播间（`live.douyin.com/[房间号]`）
3. 等待 5-10 秒，数据会自动更新

### Q5: WebView2 自动安装失败

**A**: 手动下载安装：
```
https://developer.microsoft.com/en-us/microsoft-edge/webview2/
```

下载 "Evergreen Standalone Installer"，运行安装，重启程序。

---

## 📞 技术支持

如有问题，请查看：
- `/workspace/COMPLETION_REPORT.md` - 完整功能报告
- `/workspace/FINAL_STATUS.md` - 最终状态（90% 阶段）
- `/workspace/server-go/README.md` - Go 后端文档
- `/workspace/server-active/README.md` - 许可证服务文档

---

## 🎉 祝你使用愉快！

**项目版本**: v3.1.0  
**完成度**: 🟢 100%  
**最后更新**: 2025-11-15
