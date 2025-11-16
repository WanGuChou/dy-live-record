package main

import (
	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/license"
	"dy-live-monitor/internal/server"
	"dy-live-monitor/internal/ui"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("🚀 抖音直播监控系统启动...")

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

	// 5. 启动系统托盘 UI
	log.Println("✅ 启动系统托盘...")
	ui.RunSystemTray(cfg, db, wsServer, licenseManager)
}
