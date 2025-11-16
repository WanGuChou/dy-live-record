package main

import (
	"dy-live-license/internal/api"
	"dy-live-license/internal/config"
	"dy-live-license/internal/database"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 启动许可证授权服务 (server-active)...")

	// 1. 加载配置
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Printf("⚠️  加载配置失败，使用默认配置: %v", err)
		cfg = config.GetDefaultConfig()
	}
	log.Printf("✅ 配置加载成功: %+v", cfg)

	// 2. 初始化数据库
	log.Println("🗄️  初始化 MySQL 数据库...")
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer db.Close()
	log.Println("✅ 数据库初始化成功")

	// 3. 初始化 API 路由
	router := gin.Default()
	api.SetupRoutes(router, db, cfg)

	// 4. 启动服务
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🌐 服务启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}
