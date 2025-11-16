package ui

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/server"
)

// FyneUI Fyne 图形界面
type FyneUI struct {
	app       fyne.App
	mainWin   fyne.Window
	db        *sql.DB
	wsServer  *server.WebSocketServer
	
	// 数据绑定
	giftCount    binding.String
	messageCount binding.String
	totalValue   binding.String
	onlineUsers  binding.String
	debugMode    binding.String
	
	// 表格数据
	giftTable    *widget.Table
	messageTable *widget.Table
	
	// 当前选中的房间
	currentRoom string
	
	// 配置
	cfg *config.Config
}

// NewFyneUI 创建 Fyne UI
func NewFyneUI(db *sql.DB, wsServer *server.WebSocketServer, cfg *config.Config) *FyneUI {
	fyneApp := app.NewWithID("com.dy-live-monitor")
	
	// 设置主题（确保中文字体支持）
	fyneApp.Settings().SetTheme(theme.DefaultTheme())
	
	ui := &FyneUI{
		app:          fyneApp,
		db:           db,
		wsServer:     wsServer,
		cfg:          cfg,
		giftCount:    binding.NewString(),
		messageCount: binding.NewString(),
		totalValue:   binding.NewString(),
		onlineUsers:  binding.NewString(),
		debugMode:    binding.NewString(),
	}
	
	// 初始化数据（使用英文标签避免乱码）
	ui.giftCount.Set("0")
	ui.messageCount.Set("0")
	ui.totalValue.Set("0")
	ui.onlineUsers.Set("0")
	
	// 设置调试模式状态
	if cfg.Debug.Enabled {
		ui.debugMode.Set("DEBUG MODE")
	} else {
		ui.debugMode.Set("")
	}
	
	return ui
}

// triggerBindingUpdates 触发所有绑定更新（用于初始化格式化标签）
func (ui *FyneUI) triggerBindingUpdates() {
	// 触发所有绑定的监听器，确保格式化标签正确显示
	val, _ := ui.giftCount.Get()
	ui.giftCount.Set(val)
	
	val, _ = ui.messageCount.Get()
	ui.messageCount.Set(val)
	
	val, _ = ui.totalValue.Get()
	ui.totalValue.Set(val)
	
	val, _ = ui.onlineUsers.Get()
	ui.onlineUsers.Set(val)
}

// Show 显示主窗口
func (ui *FyneUI) Show() {
	// 使用英文标题避免乱码
	title := "Douyin Live Monitor v3.2.1"
	if ui.cfg.Debug.Enabled {
		title += " [DEBUG MODE]"
	}
	
	ui.mainWin = ui.app.NewWindow(title)
	ui.mainWin.Resize(fyne.NewSize(1200, 800))
	ui.mainWin.CenterOnScreen()
	
	// 创建主界面
	content := ui.createMainContent()
	ui.mainWin.SetContent(content)
	
	// 触发初始绑定更新（确保格式化标签显示正确）
	ui.triggerBindingUpdates()
	
	// 启动数据刷新
	go ui.startDataRefresh()
	
	ui.mainWin.ShowAndRun()
}

// createMainContent 创建主界面内容
func (ui *FyneUI) createMainContent() fyne.CanvasObject {
	// 顶部统计卡片
	statsCard := ui.createStatsCard()
	
	// 创建 Tab 容器（使用英文避免乱码）
	tabs := container.NewAppTabs(
		container.NewTabItem("Overview", ui.createOverviewTab()),
		container.NewTabItem("Gifts", ui.createGiftsTab()),
		container.NewTabItem("Messages", ui.createMessagesTab()),
		container.NewTabItem("Anchors", ui.createAnchorsTab()),
		container.NewTabItem("Segments", ui.createSegmentsTab()),
		container.NewTabItem("Settings", ui.createSettingsTab()),
	)
	
	// 主布局
	return container.NewBorder(
		statsCard, // top
		nil,       // bottom
		nil,       // left
		nil,       // right
		tabs,      // center
	)
}

// createStatsCard 创建统计卡片
func (ui *FyneUI) createStatsCard() fyne.CanvasObject {
	// 创建格式化的绑定字符串（使用英文避免乱码）
	giftFormatted := binding.NewString()
	ui.giftCount.AddListener(binding.NewDataListener(func() {
		val, _ := ui.giftCount.Get()
		giftFormatted.Set(fmt.Sprintf("Gifts: %s", val))
	}))
	giftLabel := widget.NewLabelWithData(giftFormatted)
	giftLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	messageFormatted := binding.NewString()
	ui.messageCount.AddListener(binding.NewDataListener(func() {
		val, _ := ui.messageCount.Get()
		messageFormatted.Set(fmt.Sprintf("Messages: %s", val))
	}))
	messageLabel := widget.NewLabelWithData(messageFormatted)
	messageLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	valueFormatted := binding.NewString()
	ui.totalValue.AddListener(binding.NewDataListener(func() {
		val, _ := ui.totalValue.Get()
		valueFormatted.Set(fmt.Sprintf("Total Value: %s", val))
	}))
	valueLabel := widget.NewLabelWithData(valueFormatted)
	valueLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	onlineFormatted := binding.NewString()
	ui.onlineUsers.AddListener(binding.NewDataListener(func() {
		val, _ := ui.onlineUsers.Get()
		onlineFormatted.Set(fmt.Sprintf("Online: %s", val))
	}))
	onlineLabel := widget.NewLabelWithData(onlineFormatted)
	onlineLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	// 统计卡片
	statsCards := []fyne.CanvasObject{
		container.NewVBox(
			widget.NewIcon(theme.ContentAddIcon()),
			giftLabel,
		),
		container.NewVBox(
			widget.NewIcon(theme.MailComposeIcon()),
			messageLabel,
		),
		container.NewVBox(
			widget.NewIcon(theme.AccountIcon()),
			valueLabel,
		),
		container.NewVBox(
			widget.NewIcon(theme.ComputerIcon()),
			onlineLabel,
		),
	}
	
	// 如果启用调试模式，添加调试标识
	if ui.cfg.Debug.Enabled {
		debugLabel := widget.NewLabelWithData(ui.debugMode)
		debugLabel.TextStyle = fyne.TextStyle{Bold: true}
		debugCard := container.NewVBox(
			widget.NewIcon(theme.WarningIcon()),
			debugLabel,
		)
		statsCards = append(statsCards, debugCard)
	}
	
	card := container.NewGridWithColumns(len(statsCards), statsCards...)
	
	return container.NewPadded(card)
}

// createOverviewTab 创建数据概览 Tab
func (ui *FyneUI) createOverviewTab() fyne.CanvasObject {
	roomLabel := widget.NewLabel("Current Room: None")
	statusLabel := widget.NewLabel("Status: Waiting for connection...")
	
	refreshBtn := widget.NewButton("Refresh Data", func() {
		ui.refreshData()
	})
	
	infoText := `Real-time Monitor Guide:

1. Open browser and install extension
2. Visit Douyin live room
3. Extension will collect data automatically
4. Data displays here in real-time

Features:
- Gift statistics
- Message records
- Anchor management
- Segment scoring
- Data persistence
`
	
	// 如果启用调试模式，添加警告信息
	if ui.cfg.Debug.Enabled {
		infoText += `
DEBUG MODE ENABLED
`
		if ui.cfg.Debug.SkipLicense {
			infoText += `License validation skipped (Debug only)
`
		}
		if ui.cfg.Debug.VerboseLog {
			infoText += `Verbose logging enabled
`
		}
		infoText += `
WARNING: Debug mode for development only!
Please disable in config.json for production.
`
	}
	
	info := widget.NewLabel(infoText)
	
	return container.NewVBox(
		roomLabel,
		statusLabel,
		refreshBtn,
		widget.NewSeparator(),
		info,
	)
}

// createGiftsTab 创建礼物记录 Tab
func (ui *FyneUI) createGiftsTab() fyne.CanvasObject {
	// 创建礼物表格（使用英文表头）
	ui.giftTable = widget.NewTable(
		func() (int, int) { return 0, 6 }, // 行数, 列数
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			// 表头
			if id.Row == 0 {
				headers := []string{"Time", "User", "Gift", "Count", "Value", "Room"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				label.SetText(fmt.Sprintf("Data %d-%d", id.Row, id.Col))
			}
		},
	)
	
	ui.giftTable.SetColumnWidth(0, 150) // 时间
	ui.giftTable.SetColumnWidth(1, 120) // 用户
	ui.giftTable.SetColumnWidth(2, 120) // 礼物
	ui.giftTable.SetColumnWidth(3, 80)  // 数量
	ui.giftTable.SetColumnWidth(4, 100) // 价值
	ui.giftTable.SetColumnWidth(5, 100) // 房间
	
	refreshBtn := widget.NewButton("Refresh", func() {
		ui.loadGiftData()
	})
	
	exportBtn := widget.NewButton("Export", func() {
		// TODO: 实现导出功能
		log.Println("Export gift data")
	})
	
	toolbar := container.NewHBox(
		refreshBtn,
		exportBtn,
	)
	
	return container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewScroll(ui.giftTable),
	)
}

// createMessagesTab 创建消息记录 Tab
func (ui *FyneUI) createMessagesTab() fyne.CanvasObject {
	// 创建消息表格（使用英文表头）
	ui.messageTable = widget.NewTable(
		func() (int, int) { return 0, 4 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				headers := []string{"Time", "User", "Content", "Type"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				label.SetText(fmt.Sprintf("Message %d-%d", id.Row, id.Col))
			}
		},
	)
	
	ui.messageTable.SetColumnWidth(0, 150)
	ui.messageTable.SetColumnWidth(1, 120)
	ui.messageTable.SetColumnWidth(2, 400)
	ui.messageTable.SetColumnWidth(3, 100)
	
	refreshBtn := widget.NewButton("Refresh", func() {
		ui.loadMessageData()
	})
	
	clearBtn := widget.NewButton("Clear", func() {
		// TODO: 实现清空功能
		log.Println("Clear message records")
	})
	
	toolbar := container.NewHBox(
		refreshBtn,
		clearBtn,
	)
	
	return container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewScroll(ui.messageTable),
	)
}

// createAnchorsTab 创建主播管理 Tab
func (ui *FyneUI) createAnchorsTab() fyne.CanvasObject {
	// 主播列表（使用英文）
	anchorList := widget.NewList(
		func() int { return 0 }, // TODO: 从数据库加载
		func() fyne.CanvasObject {
			return widget.NewLabel("Anchor Name")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// TODO: 更新列表项
		},
	)
	
	// 添加主播表单
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Anchor Name")
	
	boundGiftsEntry := widget.NewEntry()
	boundGiftsEntry.SetPlaceHolder("Bound Gifts (comma separated)")
	
	addBtn := widget.NewButton("Add Anchor", func() {
		name := nameEntry.Text
		gifts := boundGiftsEntry.Text
		if name != "" {
			// TODO: 保存到数据库
			log.Printf("Add anchor: %s, gifts: %s", name, gifts)
			nameEntry.SetText("")
			boundGiftsEntry.SetText("")
		}
	})
	
	form := container.NewVBox(
		widget.NewLabel("Add New Anchor"),
		nameEntry,
		boundGiftsEntry,
		addBtn,
	)
	
	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("Anchor List"),
			nil, nil, nil,
			anchorList,
		),
		container.NewPadded(form),
	)
}

// createSegmentsTab 创建分段记分 Tab
func (ui *FyneUI) createSegmentsTab() fyne.CanvasObject {
	segmentEntry := widget.NewEntry()
	segmentEntry.SetPlaceHolder("Segment name (e.g., Round 1)")
	
	createBtn := widget.NewButton("Create New Segment", func() {
		name := segmentEntry.Text
		if name != "" {
			// TODO: 创建分段
			log.Printf("Create segment: %s", name)
			segmentEntry.SetText("")
		}
	})
	
	endBtn := widget.NewButton("End Current Segment", func() {
		// TODO: 结束分段
		log.Println("End current segment")
	})
	
	// 分段列表
	segmentList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Segment Record")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// TODO: 更新列表
		},
	)
	
	toolbar := container.NewVBox(
		widget.NewLabel("Segment Management"),
		segmentEntry,
		container.NewHBox(createBtn, endBtn),
		widget.NewSeparator(),
	)
	
	return container.NewBorder(
		toolbar,
		nil, nil, nil,
		segmentList,
	)
}

// createSettingsTab 创建设置 Tab
func (ui *FyneUI) createSettingsTab() fyne.CanvasObject {
	// 端口设置
	portEntry := widget.NewEntry()
	portEntry.SetText("8080")
	portEntry.SetPlaceHolder("WebSocket Port")
	
	portForm := container.NewVBox(
		widget.NewLabel("WebSocket Port"),
		portEntry,
		widget.NewButton("Save", func() {
			// TODO: 保存端口设置
			log.Printf("Save port: %s", portEntry.Text)
		}),
	)
	
	// 插件管理
	installBtn := widget.NewButton("Install Browser Extension", func() {
		// TODO: 安装插件
		log.Println("Install browser extension")
	})
	
	removeBtn := widget.NewButton("Uninstall Browser Extension", func() {
		// TODO: 卸载插件
		log.Println("Uninstall browser extension")
	})
	
	pluginSection := container.NewVBox(
		widget.NewLabel("Browser Extension Management"),
		installBtn,
		removeBtn,
	)
	
	// License 设置
	licenseEntry := widget.NewEntry()
	licenseEntry.SetPlaceHolder("Paste License Key")
	licenseEntry.MultiLine = true
	licenseEntry.SetMinRowsVisible(3)
	
	activateBtn := widget.NewButton("Activate", func() {
		// TODO: 激活 License
		log.Printf("Activate License: %s", licenseEntry.Text)
	})
	
	licenseSection := container.NewVBox(
		widget.NewLabel("License Management"),
		licenseEntry,
		activateBtn,
		widget.NewLabel("Status: Not Activated"),
	)
	
	return container.NewVBox(
		portForm,
		widget.NewSeparator(),
		pluginSection,
		widget.NewSeparator(),
		licenseSection,
	)
}

// startDataRefresh 启动数据刷新
func (ui *FyneUI) startDataRefresh() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		ui.refreshData()
	}
}

// refreshData 刷新数据
func (ui *FyneUI) refreshData() {
	// TODO: 从数据库查询最新数据
	// 这里是示例，实际需要查询数据库
	
	// 查询礼物总数
	var giftCount int
	err := ui.db.QueryRow("SELECT COUNT(*) FROM gifts").Scan(&giftCount)
	if err == nil {
		ui.giftCount.Set(fmt.Sprintf("%d", giftCount))
	}
	
	// 查询消息总数
	var messageCount int
	err = ui.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&messageCount)
	if err == nil {
		ui.messageCount.Set(fmt.Sprintf("%d", messageCount))
	}
	
	// 查询礼物总值
	var totalValue int
	err = ui.db.QueryRow("SELECT COALESCE(SUM(diamond_count), 0) FROM gifts").Scan(&totalValue)
	if err == nil {
		ui.totalValue.Set(fmt.Sprintf("%d", totalValue))
	}
	
	// 在线用户（示例）
	ui.onlineUsers.Set("N/A")
}

// loadGiftData 加载礼物数据
func (ui *FyneUI) loadGiftData() {
	// TODO: 从数据库加载礼物数据并更新表格
	log.Println("加载礼物数据")
	ui.giftTable.Refresh()
}

// loadMessageData 加载消息数据
func (ui *FyneUI) loadMessageData() {
	// TODO: 从数据库加载消息数据并更新表格
	log.Println("加载消息数据")
	ui.messageTable.Refresh()
}
