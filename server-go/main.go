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
)

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

	// 3. 许可证校验（强制）
	licenseManager := license.NewManager(cfg.License.ServerURL, cfg.License.PublicKeyPath)
	
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

	// 4. 启动 WebSocket 服务器
	wsServer := server.NewWebSocketServer(cfg.Server.Port, db)
	go func() {
		if err := wsServer.Start(); err != nil {
			log.Fatalf("❌ WebSocket 服务器启动失败: %v", err)
		}
	}()
	log.Printf("✅ WebSocket 服务器启动成功 (端口: %d)", cfg.Server.Port)

	// 5. 启动 Fyne GUI（主窗口）
	log.Println("✅ 启动图形界面...")
	
	// 在单独的 goroutine 中运行系统托盘
	go ui.RunSystemTray(cfg, db, wsServer, licenseManager)
	
	// 主线程运行 Fyne GUI
	fyneUI := ui.NewFyneUI(db, wsServer)
	fyneUI.Show() // 这会阻塞直到窗口关闭
}
