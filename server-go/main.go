package main

import (
	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/dependencies"
	"dy-live-monitor/internal/license"
	"dy-live-monitor/internal/server"
	"dy-live-monitor/internal/ui"
	"fmt"
	"log"
	"os"
	"syscall"
)

func init() {
	// 设置 Windows 控制台为 UTF-8 编码，避免中文乱码
	if kernel32, err := syscall.LoadDLL("kernel32.dll"); err == nil {
		if setConsoleCP, err := kernel32.FindProc("SetConsoleCP"); err == nil {
			setConsoleCP.Call(65001) // UTF-8
		}
		if setConsoleOutputCP, err := kernel32.FindProc("SetConsoleOutputCP"); err == nil {
			setConsoleOutputCP.Call(65001) // UTF-8
		}
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("🚀 " + GetVersionInfo() + " 启动...")

	// 0. 检查依赖
	checker := dependencies.NewChecker()
	if !checker.CheckAll() {
		log.Println("\n⚠️  检测到关键依赖缺失")
		fmt.Print("\n是否尝试自动安装 WebView2? (y/n): ")
		var response string
		fmt.Scanln(&response)
		
		if response == "y" || response == "Y" {
			if err := checker.AutoInstallWebView2(); err != nil {
				log.Printf("❌ 自动安装失败: %v", err)
				log.Println("请手动安装后重启程序")
			} else {
				log.Println("✅ 安装成功！请重启程序")
			}
		}
		
		log.Println("\n按任意键退出...")
		fmt.Scanln()
		os.Exit(1)
	}

	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("⚠️  加载配置失败: %v, 使用默认配置", err)
		cfg = config.Default()
	}

	// 2. 初始化数据库
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer db.Close()
	log.Println("✅ 数据库初始化成功")

	// 3. 许可证校验
	licenseManager := license.NewManager(cfg.License.ServerURL, cfg.License.PublicKeyPath)
	
	// 检查是否启用调试模式
	if cfg.Debug.Enabled && cfg.Debug.SkipLicense {
		log.Println("⚠️  调试模式已启用，跳过 License 验证")
		log.Println("⚠️  警告：调试模式仅供开发使用，生产环境请禁用！")
	} else {
		// 读取本地许可证
		localLicense, err := licenseManager.LoadLocal()
		if err != nil || localLicense == "" {
			log.Println("⚠️  未找到有效许可证，请激活软件")
			// 显示激活窗口
			ui.ShowActivationDialog(licenseManager)
			os.Exit(1)
		}

		// 校验许可证
		valid, expiryDate, err := licenseManager.Validate(localLicense)
		if err != nil || !valid {
			log.Printf("❌ 许可证校验失败: %v", err)
			ui.ShowActivationDialog(licenseManager)
			os.Exit(1)
		}

		log.Printf("✅ 许可证校验通过，有效期至: %s", expiryDate.Format("2006-01-02"))
	}

	// 4. 启动 WebSocket 服务器
	log.Printf("📡 正在启动 WebSocket 服务器 (端口: %d)...", cfg.Server.Port)
	wsServer := server.NewWebSocketServer(cfg.Server.Port, db)
	
	if err := wsServer.Start(); err != nil {
		log.Fatalf("❌ WebSocket 服务器启动失败: %v", err)
	}
	
	log.Printf("✅ WebSocket 服务器启动成功！")
	log.Printf("   📍 连接地址: ws://localhost:%d/ws", cfg.Server.Port)
	log.Printf("   📍 健康检查: http://localhost:%d/health", cfg.Server.Port)
	log.Printf("   💡 提示: 浏览器插件需连接到此地址")

	// 5. 启动 Fyne GUI（主窗口）
	log.Println("✅ 启动图形界面...")
	
	// 在单独的 goroutine 中运行系统托盘
	go ui.RunSystemTray(cfg, db, wsServer, licenseManager)
	
	// 主线程运行 Fyne GUI
	fyneUI := ui.NewFyneUI(db.GetConn(), wsServer, cfg)
	fyneUI.Show() // 这会阻塞直到窗口关闭
}
