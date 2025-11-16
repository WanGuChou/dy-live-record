# 启动脚本使用说明

## 📝 脚本说明

本目录包含3个Windows启动脚本，用于隐藏"正在调试此浏览器"横幅。

### 文件列表

| 脚本 | 说明 | 推荐度 |
|------|------|--------|
| `START_CHROME.bat` | 自动检测Chrome路径并启动 | ⭐⭐⭐⭐⭐ |
| `START_EDGE.bat` | 自动检测Edge路径并启动 | ⭐⭐⭐⭐⭐ |
| `START_CHROME_MANUAL.bat` | 手动指定Chrome路径 | ⭐⭐⭐ |
| `start-chrome.sh` | macOS/Linux启动脚本 | ⭐⭐⭐⭐⭐ |

---

## 🚀 使用方法

### Windows用户

#### 方法1: 自动检测（推荐）

**启动Chrome:**
```
双击运行: START_CHROME.bat
```

**启动Edge:**
```
双击运行: START_EDGE.bat
```

**脚本会自动：**
1. ✅ 检测Chrome/Edge安装路径
2. ✅ 关闭现有浏览器实例
3. ✅ 使用隐藏调试横幅的参数启动
4. ✅ 显示详细的进度信息

#### 方法2: 手动指定路径

如果自动检测失败：

```
双击运行: START_CHROME_MANUAL.bat
```

按提示输入Chrome完整路径，例如：
```
C:\Program Files\Google\Chrome\Application\chrome.exe
```

### macOS/Linux用户

```bash
cd dy-live-record/brower-monitor
chmod +x start-chrome.sh  # 首次运行需要
./start-chrome.sh
```

---

## 🔧 常见路径

### Chrome

**Windows:**
- `C:\Program Files\Google\Chrome\Application\chrome.exe` (64位系统默认)
- `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe` (32位或旧版)

**macOS:**
- `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`

**Linux:**
- `/usr/bin/google-chrome`
- `/usr/bin/google-chrome-stable`
- `/usr/bin/chromium-browser`

### Edge

**Windows:**
- `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe` (常见)
- `C:\Program Files\Microsoft\Edge\Application\msedge.exe`

---

## ❓ 常见问题

### Q1: 中文显示乱码

**现象：**
```
CDP Monitor - Chrome 鍚姩鑴氭湰
```

**原因：** CMD编码问题

**解决：** 
- 最新版本脚本已自动修复（使用 `chcp 65001`）
- 如果仍有问题，右键CMD窗口 → 属性 → 选项 → 当前代码页 → 选择"UTF-8"

### Q2: 提示"找不到Chrome"

**解决方法：**

1. **检查Chrome是否已安装**
   ```
   在开始菜单搜索"Chrome"
   ```

2. **使用手动路径脚本**
   ```
   运行: START_CHROME_MANUAL.bat
   输入完整路径
   ```

3. **查找Chrome实际路径**
   ```
   按Win+R，输入: %ProgramFiles%
   或: %ProgramFiles(x86)%
   查找: Google\Chrome\Application\chrome.exe
   ```

### Q3: 提示"此时不应有..."

**原因：** 路径引号问题

**解决：** 
- 使用最新版本脚本（已修复）
- 或使用 `START_CHROME_MANUAL.bat` 手动输入路径

### Q4: Chrome启动后仍显示调试横幅

**原因：** Chrome在后台已经运行

**解决：**
1. 完全关闭Chrome（包括后台进程）
2. 打开任务管理器（Ctrl+Shift+Esc）
3. 结束所有 `chrome.exe` 进程
4. 重新运行启动脚本

**验证方法：**
```powershell
# 在PowerShell中执行
Get-Process chrome -ErrorAction SilentlyContinue | Stop-Process -Force
```

### Q5: 脚本闪退

**原因：** 可能遇到错误但窗口立即关闭

**解决：**
1. 右键脚本 → 编辑
2. 在最后一行前添加：`pause`
3. 保存并重新运行
4. 查看错误信息

或者在CMD中手动运行：
```cmd
cd C:\path\to\brower-monitor
START_CHROME.bat
```

---

## 🔍 调试技巧

### 验证Chrome是否使用了正确参数

**Windows (PowerShell):**
```powershell
Get-Process chrome | Select-Object -ExpandProperty CommandLine
```

**输出应该包含：**
```
--silent-debugger-extension-api
```

### 手动测试命令

如果脚本不工作，可以在CMD中手动测试：

```cmd
REM 关闭Chrome
taskkill /F /IM chrome.exe

REM 等待2秒
timeout /t 2

REM 启动Chrome
"C:\Program Files\Google\Chrome\Application\chrome.exe" --silent-debugger-extension-api
```

---

## 💡 进阶使用

### 自定义启动参数

编辑脚本，在启动命令中添加更多参数：

```batch
start "" "%CHROME_PATH%" ^
  --silent-debugger-extension-api ^
  --disable-blink-features=AutomationControlled ^
  --user-data-dir=%USERPROFILE%\ChromeDevProfile ^
  --no-first-run
```

**常用参数：**
- `--user-data-dir=PATH` - 使用独立配置文件
- `--disable-blink-features=AutomationControlled` - 禁用自动化检测
- `--no-first-run` - 跳过首次运行向导
- `--no-default-browser-check` - 不检查默认浏览器
- `--start-maximized` - 最大化启动
- `--incognito` - 无痕模式

### 创建桌面快捷方式

1. 右键脚本 → 发送到 → 桌面快捷方式
2. 右键快捷方式 → 属性
3. 修改图标（可选）
4. 点击"确定"

### 添加到右键菜单

创建 `add-to-context-menu.reg`:

```reg
Windows Registry Editor Version 5.00

[HKEY_CLASSES_ROOT\Directory\Background\shell\StartChromeDebug]
@="启动Chrome (隐藏调试横幅)"
"Icon"="C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"

[HKEY_CLASSES_ROOT\Directory\Background\shell\StartChromeDebug\command]
@="\"C:\\path\\to\\START_CHROME.bat\""
```

修改路径后双击运行。

---

## 📚 相关文档

- **完整指南**: [HIDE_DEBUGGER_BANNER.md](../../HIDE_DEBUGGER_BANNER.md)
- **CDP使用**: [CDP_USAGE.md](../../CDP_USAGE.md)
- **测试指南**: [CDP_TEST.md](../../CDP_TEST.md)
- **主文档**: [README.md](../../README.md)

---

## 🆘 获取帮助

如果脚本仍然无法工作：

1. **查看详细日志**
   - 在CMD中手动运行脚本
   - 查看完整错误信息

2. **检查Chrome版本**
   ```
   chrome://version/
   ```
   确保Chrome版本 >= 90

3. **临时禁用杀毒软件**
   - 某些杀毒软件可能阻止脚本

4. **使用手动命令**
   - 参考"手动测试命令"部分
   - 逐步执行每个命令

5. **报告问题**
   - GitHub Issues: https://github.com/WanGuChou/dy-live-record/issues
   - 包含错误信息和系统信息

---

**更新时间**: 2025-11-15  
**版本**: v2.0.0  
**支持系统**: Windows 10/11, macOS, Linux
