package ui

import (
	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/license"
	"dy-live-monitor/internal/server"
	"log"

	"github.com/getlantern/systray"
)

// RunSystemTray 运行系统托盘
func RunSystemTray(
	cfg *config.Config,
	db *database.DB,
	wsServer *server.WebSocketServer,
	licManager *license.Manager,
) {
	systray.Run(
		func() { onReady(cfg, db, wsServer, licManager) },
		onExit,
	)
}

// onReady 系统托盘初始化
func onReady(
	cfg *config.Config,
	db *database.DB,
	wsServer *server.WebSocketServer,
	licManager *license.Manager,
) {
	// 设置图标和标题
	// systray.SetIcon(icon.Data) // TODO: 添加图标
	systray.SetTitle("抖音直播监控")
	systray.SetTooltip("抖音直播数据统计系统")

	// 菜单项
	mOpen := systray.AddMenuItem("打开主界面", "显示数据看板")
	mRooms := systray.AddMenuItem("当前监控房间", "查看正在监控的直播间")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("设置", "配置系统参数")
	mLicense := systray.AddMenuItem("许可证管理", "查看和更新许可证")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出程序", "关闭应用")

	// 处理点击事件
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				log.Println("📊 打开主界面")
				// TODO: 打开 webview2 主界面
				ShowMainWindow(db, wsServer)

			case <-mRooms.ClickedCh:
				log.Println("🏠 查看监控房间")
				// TODO: 显示当前房间列表

			case <-mSettings.ClickedCh:
				log.Println("⚙️  打开设置")
				ShowSettingsDialog(cfg)

			case <-mLicense.ClickedCh:
				log.Println("🔑 许可证管理")
				ShowLicenseDialog(licManager)

			case <-mQuit.ClickedCh:
				log.Println("👋 退出程序")
				systray.Quit()
				return
			}
		}
	}()
}

// onExit 退出回调
func onExit() {
	log.Println("🛑 系统托盘已退出")
}

// ShowMainWindow 显示主界面
func ShowMainWindow(db *database.DB, wsServer *server.WebSocketServer) {
	// TODO: 使用 webview2 显示主界面
	// - Tab 标签页切换房间
	// - 数据看板（礼物、消息、统计）
	// - 历史记录
	log.Println("⚠️  主界面功能待实现")
}

// ShowSettingsDialog 显示设置对话框
func ShowSettingsDialog(cfg *config.Config) {
	// TODO: 设置界面
	// - 端口号
	// - 浏览器路径
	// - 插件管理
	log.Println("⚠️  设置界面待实现")
}

// ShowLicenseDialog 显示许可证对话框
func ShowLicenseDialog(licManager *license.Manager) {
	// TODO: 许可证界面
	// - 当前许可证信息
	// - 有效期
	// - 激活新许可证
	log.Println("⚠️  许可证界面待实现")
}

// ShowActivationDialog 显示激活对话框
func ShowActivationDialog(licManager *license.Manager) {
	// TODO: 激活界面
	// - 输入许可证密钥
	// - 在线激活
	// - 离线激活
	log.Println("⚠️  激活界面待实现")
}
