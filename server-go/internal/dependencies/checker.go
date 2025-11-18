package dependencies

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CheckResult 依赖检查结果
type CheckResult struct {
	Name      string
	Installed bool
	Version   string
	Message   string
	Critical  bool // 是否为关键依赖
}

// Checker 依赖检查器
type Checker struct {
	results []CheckResult
}

// NewChecker 创建依赖检查器
func NewChecker() *Checker {
	return &Checker{
		results: make([]CheckResult, 0),
	}
}

// CheckAll 检查所有依赖
func (c *Checker) CheckAll() bool {
	log.Println("🔍 开始检查系统依赖...")

	// 1. 检查 WebView2 Runtime
	c.checkWebView2()

	// 2. 检查 SQLite 支持（CGO）
	c.checkSQLite()

	// 3. 检查网络连接
	c.checkNetwork()

	// 4. 检查磁盘空间
	c.checkDiskSpace()

	// 打印结果
	c.printResults()

	// 检查是否有关键依赖缺失
	hasError := false
	for _, result := range c.results {
		if result.Critical && !result.Installed {
			hasError = true
		}
	}

	if hasError {
		log.Println("❌ 关键依赖缺失，程序可能无法正常运行")
		c.showInstallGuide()
		return false
	}

	log.Println("✅ 所有依赖检查通过")
	return true
}

// checkWebView2 检查 WebView2 Runtime
func (c *Checker) checkWebView2() {
	if runtime.GOOS != "windows" {
		c.results = append(c.results, CheckResult{
			Name:      "WebView2 Runtime",
			Installed: false,
			Message:   "非 Windows 平台，跳过检查",
			Critical:  false,
		})
		return
	}

	// 检查 WebView2 安装路径
	paths := []string{
		`C:\Program Files (x86)\Microsoft\EdgeWebView\Application`,
		`C:\Program Files\Microsoft\EdgeWebView\Application`,
		os.Getenv("LOCALAPPDATA") + `\Microsoft\EdgeWebView\Application`,
	}

	installed := false
	version := ""

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// 尝试读取版本
			entries, err := os.ReadDir(path)
			if err == nil && len(entries) > 0 {
				// 第一个目录通常是版本号
				for _, entry := range entries {
					if entry.IsDir() && strings.Contains(entry.Name(), ".") {
						version = entry.Name()
						installed = true
						break
					}
				}
			}
			if installed {
				break
			}
		}
	}

	// Windows 11 和 Windows 10 (20H1+) 自带 WebView2
	if !installed {
		// 检查 Edge 浏览器（包含 WebView2）
		edgePath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
		if _, err := os.Stat(edgePath); err == nil {
			installed = true
			version = "Installed with Edge"
		}
	}

	c.results = append(c.results, CheckResult{
		Name:      "WebView2 Runtime",
		Installed: installed,
		Version:   version,
		Message:   c.getWebView2Message(installed),
		Critical:  true,
	})
}

// checkSQLite 检查 SQLite 支持
func (c *Checker) checkSQLite() {
	// 尝试检查 CGO 是否可用
	cmd := exec.Command("go", "env", "CGO_ENABLED")
	output, err := cmd.Output()

	cgoEnabled := false
	if err == nil {
		cgoEnabled = strings.TrimSpace(string(output)) == "1"
	}

	// 检查 gcc/mingw（Windows）
	gccInstalled := false
	if runtime.GOOS == "windows" {
		cmd := exec.Command("gcc", "--version")
		if err := cmd.Run(); err == nil {
			gccInstalled = true
		}
	}

	// SQLite 驱动在编译时就需要 CGO
	// 运行时无法直接检查，但可以提示
	c.results = append(c.results, CheckResult{
		Name:      "SQLite Driver (CGO)",
		Installed: cgoEnabled || gccInstalled,
		Version:   fmt.Sprintf("CGO_ENABLED=%v, GCC=%v", cgoEnabled, gccInstalled),
		Message:   c.getSQLiteMessage(cgoEnabled, gccInstalled),
		Critical:  true,
	})
}

// checkNetwork 检查网络连接
func (c *Checker) checkNetwork() {
	// 尝试 ping NTP 服务器（用于许可证校验）
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "2000", "pool.ntp.org")
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "2", "pool.ntp.org")
	}

	err := cmd.Run()
	networkOK := err == nil

	c.results = append(c.results, CheckResult{
		Name:      "网络连接",
		Installed: networkOK,
		Message:   c.getNetworkMessage(networkOK),
		Critical:  false, // 离线模式下也可以运行
	})
}

// checkDiskSpace 检查磁盘空间
func (c *Checker) checkDiskSpace() {
	// 简单检查当前目录是否可写
	testFile := ".test_write"
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err == nil {
		os.Remove(testFile)
	}

	c.results = append(c.results, CheckResult{
		Name:      "磁盘空间",
		Installed: err == nil,
		Message:   c.getDiskMessage(err == nil),
		Critical:  true,
	})
}

// printResults 打印检查结果
func (c *Checker) printResults() {
	log.Println()
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                    依赖检查结果                              ║")
	log.Println("╠══════════════════════════════════════════════════════════════╣")

	for _, result := range c.results {
		status := "❌"
		if result.Installed {
			status = "✅"
		}

		critical := ""
		if result.Critical {
			critical = " [关键]"
		}

		log.Printf("║ %s %-30s %s", status, result.Name+critical, "")
		if result.Version != "" {
			log.Printf("║    版本: %s", result.Version)
		}
		log.Printf("║    %s", result.Message)
		log.Println("╠──────────────────────────────────────────────────────────────╣")
	}

	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println()
}

// showInstallGuide 显示安装指南
func (c *Checker) showInstallGuide() {
	log.Println()
	log.Println("📖 依赖安装指南:")
	log.Println("────────────────────────────────────────────────────────────")

	for _, result := range c.results {
		if result.Critical && !result.Installed {
			log.Println()
			log.Printf("【%s】", result.Name)
			log.Println(c.getInstallGuide(result.Name))
		}
	}

	log.Println("\n────────────────────────────────────────────────────────────")
}

// getWebView2Message 获取 WebView2 检查消息
func (c *Checker) getWebView2Message(installed bool) string {
	if installed {
		return "已安装，UI 可以正常显示"
	}
	return "未检测到，需要安装 WebView2 Runtime"
}

// getSQLiteMessage 获取 SQLite 检查消息
func (c *Checker) getSQLiteMessage(cgo, gcc bool) string {
	if cgo || gcc {
		return "CGO 支持已启用，SQLite 可以正常使用"
	}
	return "CGO 未启用或 GCC 未安装，SQLite 可能无法工作"
}

// getNetworkMessage 获取网络检查消息
func (c *Checker) getNetworkMessage(ok bool) string {
	if ok {
		return "网络连接正常，可以进行在线许可证校验"
	}
	return "网络连接失败，将使用离线模式"
}

// getDiskMessage 获取磁盘检查消息
func (c *Checker) getDiskMessage(ok bool) string {
	if ok {
		return "磁盘可写，数据库可以正常存储"
	}
	return "磁盘不可写，请检查权限"
}

// getInstallGuide 获取安装指南
func (c *Checker) getInstallGuide(name string) string {
	switch name {
	case "WebView2 Runtime":
		return `下载链接: https://developer.microsoft.com/en-us/microsoft-edge/webview2/
安装方法:
  1. 访问上述链接
  2. 下载 "Evergreen Standalone Installer"
  3. 运行安装程序
  4. 重启本程序

提示: Windows 11 和 Windows 10 (20H1+) 通常已自带 WebView2`

	case "SQLite Driver (CGO)":
		return `Windows 平台需要安装 MinGW-w64:
安装方法 1 (推荐 - Chocolatey):
  choco install mingw

安装方法 2 (手动):
  1. 下载: https://sourceforge.net/projects/mingw-w64/
  2. 安装到 C:\mingw-w64
  3. 添加到 PATH: C:\mingw-w64\bin
  4. 重新编译本程序: go build

编译时确保 CGO_ENABLED=1`

	case "磁盘空间":
		return `请检查:
  1. 当前目录是否有写权限
  2. 磁盘是否已满
  3. 是否以管理员身份运行`

	default:
		return "请查阅文档获取安装指南"
	}
}

// AutoInstallWebView2 自动安装 WebView2（如果可能）
func (c *Checker) AutoInstallWebView2() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("非 Windows 平台")
	}

	log.Println("🔧 尝试自动下载并安装 WebView2 Runtime...")

	// 下载 URL
	url := "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
	installerPath := "webview2_installer.exe"

	// 使用 PowerShell 下载
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Invoke-WebRequest -Uri '%s' -OutFile '%s'", url, installerPath))

	log.Println("📥 正在下载 WebView2 安装程序...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 运行安装程序
	log.Println("📦 正在安装 WebView2...")
	cmd = exec.Command(installerPath, "/silent", "/install")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	// 清理
	os.Remove(installerPath)

	log.Println("✅ WebView2 安装完成！")
	return nil
}
