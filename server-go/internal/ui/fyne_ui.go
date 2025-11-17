package ui

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"

	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/server"
)

const maxStoredMessages = 200

func init() {
	// 设置中文字体：解决中文乱码问题
	log.Println("🔍 正在查找系统中文字体...")

	fontPaths := findfont.List()
	fontFound := false

	// 优先级顺序：微软雅黑 > 黑体 > 宋体 > 楷体
	fontPriority := []string{"msyh.ttf", "msyhbd.ttf", "simhei.ttf", "simsun.ttc", "simkai.ttf"}

	for _, fontName := range fontPriority {
		for _, path := range fontPaths {
			if strings.Contains(strings.ToLower(path), strings.ToLower(fontName)) {
				os.Setenv("FYNE_FONT", path)
				log.Printf("✅ 找到中文字体: %s", path)
				fontFound = true
				break
			}
		}
		if fontFound {
			break
		}
	}

	if !fontFound {
		log.Println("⚠️  警告：未找到常见中文字体，将使用系统默认字体")
		log.Println("💡 提示：如果中文显示异常，请安装 Microsoft YaHei 字体")
	}
}

// MessagePair 记录原始消息
type MessagePair struct {
	ID         int64
	RawMessage string
	Timestamp  time.Time
}

// ParsedMessageRecord 保存解析后的消息与原始消息的关联
type ParsedMessageRecord struct {
	ID        int64
	RawID     int64
	Summary   string
	Detail    map[string]interface{}
	Timestamp time.Time
}

// RoomTab 房间Tab数据
type RoomTab struct {
	RoomID        string
	Tab           *container.TabItem
	RawMessages   *widget.List
	ParsedMsgs    *widget.List
	RawData       []string
	MessagePairs  []*MessagePair // 消息对列表
	ParsedRecords []*ParsedMessageRecord
	StatsLabel    *widget.Label
	DetailWindow  fyne.Window // 详情窗口
	nextRawID     int64
	nextParsedID  int64
}

// FyneUI Fyne 图形界面
type FyneUI struct {
	app      fyne.App
	mainWin  fyne.Window
	db       *sql.DB
	wsServer *server.WebSocketServer

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

	// 动态房间 Tabs
	roomTabs     map[string]*RoomTab
	tabContainer *container.AppTabs

	// 配置
	cfg *config.Config
}

// NewFyneUI 创建 Fyne UI
func NewFyneUI(db *sql.DB, wsServer *server.WebSocketServer, cfg *config.Config) *FyneUI {
	fyneApp := app.NewWithID("com.dy-live-monitor")

	// 设置支持中文的主题
	fyneApp.Settings().SetTheme(NewChineseTheme())

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
		roomTabs:     make(map[string]*RoomTab),
	}

	// 初始化数据
	ui.giftCount.Set("0")
	ui.messageCount.Set("0")
	ui.totalValue.Set("0")
	ui.onlineUsers.Set("0")

	// 设置调试模式状态
	if cfg.Debug.Enabled {
		ui.debugMode.Set("⚠️ 调试模式")
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
	// 使用中文标题
	title := "抖音直播监控系统 v3.2.1"
	if ui.cfg.Debug.Enabled {
		title += " [调试模式]"
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

	// 创建 Tab 容器
	ui.tabContainer = container.NewAppTabs(
		container.NewTabItem("📊 数据概览", ui.createOverviewTab()),
		container.NewTabItem("🎁 礼物记录", ui.createGiftsTab()),
		container.NewTabItem("💬 消息记录", ui.createMessagesTab()),
		container.NewTabItem("👤 主播管理", ui.createAnchorsTab()),
		container.NewTabItem("📈 分段记分", ui.createSegmentsTab()),
		container.NewTabItem("⚙️ 设置", ui.createSettingsTab()),
	)

	// 主布局
	return container.NewBorder(
		statsCard,       // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		ui.tabContainer, // center
	)
}

// createStatsCard 创建统计卡片
func (ui *FyneUI) createStatsCard() fyne.CanvasObject {
	// 创建格式化的绑定字符串
	giftFormatted := binding.NewString()
	ui.giftCount.AddListener(binding.NewDataListener(func() {
		val, _ := ui.giftCount.Get()
		giftFormatted.Set(fmt.Sprintf("礼物总数: %s", val))
	}))
	giftLabel := widget.NewLabelWithData(giftFormatted)
	giftLabel.TextStyle = fyne.TextStyle{Bold: true}

	messageFormatted := binding.NewString()
	ui.messageCount.AddListener(binding.NewDataListener(func() {
		val, _ := ui.messageCount.Get()
		messageFormatted.Set(fmt.Sprintf("消息总数: %s", val))
	}))
	messageLabel := widget.NewLabelWithData(messageFormatted)
	messageLabel.TextStyle = fyne.TextStyle{Bold: true}

	valueFormatted := binding.NewString()
	ui.totalValue.AddListener(binding.NewDataListener(func() {
		val, _ := ui.totalValue.Get()
		valueFormatted.Set(fmt.Sprintf("礼物总值: %s 钻石", val))
	}))
	valueLabel := widget.NewLabelWithData(valueFormatted)
	valueLabel.TextStyle = fyne.TextStyle{Bold: true}

	onlineFormatted := binding.NewString()
	ui.onlineUsers.AddListener(binding.NewDataListener(func() {
		val, _ := ui.onlineUsers.Get()
		onlineFormatted.Set(fmt.Sprintf("在线用户: %s", val))
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
	roomLabel := widget.NewLabel("当前监控房间: 无")
	statusLabel := widget.NewLabel("状态: 等待连接...")

	refreshBtn := widget.NewButton("刷新数据", func() {
		ui.refreshData()
	})

	infoText := `📊 实时监控说明

1. 打开浏览器并安装插件
2. 访问抖音直播间
3. 插件会自动采集数据
4. 数据实时显示在这里

当前功能：
✅ 礼物统计
✅ 消息记录
✅ 主播管理
✅ 分段记分
✅ 数据持久化
`

	// 如果启用调试模式，添加警告信息
	if ui.cfg.Debug.Enabled {
		infoText += `
⚠️  调试模式已启用
`
		if ui.cfg.Debug.SkipLicense {
			infoText += `⚠️  License 验证已跳过（仅供调试）
`
		}
		if ui.cfg.Debug.VerboseLog {
			infoText += `⚠️  详细日志已启用
`
		}
		infoText += `
❗ 警告：调试模式仅供开发使用，
   生产环境请在 config.json 中禁用！
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
	// 创建礼物表格
	ui.giftTable = widget.NewTable(
		func() (int, int) { return 0, 6 }, // 行数, 列数
		func() fyne.CanvasObject {
			return widget.NewLabel("模板")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			// 表头
			if id.Row == 0 {
				headers := []string{"时间", "用户", "礼物", "数量", "价值", "房间"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				label.SetText(fmt.Sprintf("数据 %d-%d", id.Row, id.Col))
			}
		},
	)

	ui.giftTable.SetColumnWidth(0, 150) // 时间
	ui.giftTable.SetColumnWidth(1, 120) // 用户
	ui.giftTable.SetColumnWidth(2, 120) // 礼物
	ui.giftTable.SetColumnWidth(3, 80)  // 数量
	ui.giftTable.SetColumnWidth(4, 100) // 价值
	ui.giftTable.SetColumnWidth(5, 100) // 房间

	refreshBtn := widget.NewButton("刷新", func() {
		ui.loadGiftData()
	})

	exportBtn := widget.NewButton("导出", func() {
		// TODO: 实现导出功能
		log.Println("导出礼物数据")
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
	// 创建消息表格
	ui.messageTable = widget.NewTable(
		func() (int, int) { return 0, 4 },
		func() fyne.CanvasObject {
			return widget.NewLabel("模板")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				headers := []string{"时间", "用户", "内容", "类型"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				label.SetText(fmt.Sprintf("消息 %d-%d", id.Row, id.Col))
			}
		},
	)

	ui.messageTable.SetColumnWidth(0, 150)
	ui.messageTable.SetColumnWidth(1, 120)
	ui.messageTable.SetColumnWidth(2, 400)
	ui.messageTable.SetColumnWidth(3, 100)

	refreshBtn := widget.NewButton("刷新", func() {
		ui.loadMessageData()
	})

	clearBtn := widget.NewButton("清空", func() {
		// TODO: 实现清空功能
		log.Println("清空消息记录")
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
	// 主播列表
	anchorList := widget.NewList(
		func() int { return 0 }, // TODO: 从数据库加载
		func() fyne.CanvasObject {
			return widget.NewLabel("主播名称")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// TODO: 更新列表项
		},
	)

	// 添加主播表单
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("主播名称")

	boundGiftsEntry := widget.NewEntry()
	boundGiftsEntry.SetPlaceHolder("绑定礼物（用逗号分隔）")

	addBtn := widget.NewButton("添加主播", func() {
		name := nameEntry.Text
		gifts := boundGiftsEntry.Text
		if name != "" {
			// TODO: 保存到数据库
			log.Printf("添加主播: %s, 礼物: %s", name, gifts)
			nameEntry.SetText("")
			boundGiftsEntry.SetText("")
		}
	})

	form := container.NewVBox(
		widget.NewLabel("添加新主播"),
		nameEntry,
		boundGiftsEntry,
		addBtn,
	)

	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("主播列表"),
			nil, nil, nil,
			anchorList,
		),
		container.NewPadded(form),
	)
}

// createSegmentsTab 创建分段记分 Tab
func (ui *FyneUI) createSegmentsTab() fyne.CanvasObject {
	segmentEntry := widget.NewEntry()
	segmentEntry.SetPlaceHolder("分段名称（如：第一轮PK）")

	createBtn := widget.NewButton("创建新分段", func() {
		name := segmentEntry.Text
		if name != "" {
			// TODO: 创建分段
			log.Printf("创建分段: %s", name)
			segmentEntry.SetText("")
		}
	})

	endBtn := widget.NewButton("结束当前分段", func() {
		// TODO: 结束分段
		log.Println("结束当前分段")
	})

	// 分段列表
	segmentList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("分段记录")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// TODO: 更新列表
		},
	)

	toolbar := container.NewVBox(
		widget.NewLabel("分段记分管理"),
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
	portEntry.SetPlaceHolder("WebSocket 端口")

	portForm := container.NewVBox(
		widget.NewLabel("WebSocket 端口"),
		portEntry,
		widget.NewButton("保存", func() {
			// TODO: 保存端口设置
			log.Printf("保存端口: %s", portEntry.Text)
		}),
	)

	// 插件管理
	installBtn := widget.NewButton("安装浏览器插件", func() {
		// TODO: 安装插件
		log.Println("安装浏览器插件")
	})

	removeBtn := widget.NewButton("卸载浏览器插件", func() {
		// TODO: 卸载插件
		log.Println("卸载浏览器插件")
	})

	pluginSection := container.NewVBox(
		widget.NewLabel("浏览器插件管理"),
		installBtn,
		removeBtn,
	)

	// License 设置
	licenseEntry := widget.NewEntry()
	licenseEntry.SetPlaceHolder("粘贴 License Key")
	licenseEntry.MultiLine = true
	licenseEntry.SetMinRowsVisible(3)

	activateBtn := widget.NewButton("激活", func() {
		// TODO: 激活 License
		log.Printf("激活 License: %s", licenseEntry.Text)
	})

	licenseSection := container.NewVBox(
		widget.NewLabel("License 管理"),
		licenseEntry,
		activateBtn,
		widget.NewLabel("当前状态: 未激活"),
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

// AddOrUpdateRoom 添加或更新房间Tab
func (ui *FyneUI) AddOrUpdateRoom(roomID string) {
	// 检查房间是否已存在
	if _, exists := ui.roomTabs[roomID]; exists {
		return
	}

	log.Printf("🎬 创建房间 Tab: %s", roomID)

	// 创建房间Tab
	roomTab := &RoomTab{
		RoomID:        roomID,
		RawData:       make([]string, 0, maxStoredMessages),
		MessagePairs:  make([]*MessagePair, 0, maxStoredMessages),
		ParsedRecords: make([]*ParsedMessageRecord, 0, maxStoredMessages),
		nextRawID:     1,
		nextParsedID:  1,
	}

	// 创建统计标签
	roomTab.StatsLabel = widget.NewLabel(fmt.Sprintf("房间: %s | 消息: 0 条", roomID))

	// 创建原始消息列表（支持点击查看详情）
	roomTab.RawMessages = widget.NewList(
		func() int {
			return len(roomTab.RawData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("消息模板")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(roomTab.RawData) {
				item.(*widget.Label).SetText(roomTab.RawData[id])
			}
		},
	)

	// 原始消息点击事件：选中对应的解析消息
	roomTab.RawMessages.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(roomTab.MessagePairs) {
			return
		}
		rawID := roomTab.MessagePairs[id].ID
		if parsedIndex := roomTab.findParsedIndexByRawID(rawID); parsedIndex >= 0 {
			roomTab.ParsedMsgs.Select(parsedIndex)
			roomTab.ParsedMsgs.ScrollTo(parsedIndex)
		}
	}

	// 创建解析后消息列表（支持点击查看详情）
	roomTab.ParsedMsgs = widget.NewList(
		func() int {
			return len(roomTab.ParsedRecords)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("消息模板")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(roomTab.ParsedRecords) {
				item.(*widget.Label).SetText(roomTab.ParsedRecords[id].Summary)
			}
		},
	)

	// 解析消息点击事件：显示完整详情
	roomTab.ParsedMsgs.OnSelected = func(id widget.ListItemID) {
		ui.showMessageDetail(roomTab, id)
	}

	// 创建分割视图
	rawContainer := container.NewBorder(
		widget.NewLabel("📡 原始 WebSocket 消息"),
		nil, nil, nil,
		container.NewScroll(roomTab.RawMessages),
	)

	parsedContainer := container.NewBorder(
		widget.NewLabel("📋 解析后的消息"),
		nil, nil, nil,
		container.NewScroll(roomTab.ParsedMsgs),
	)

	split := container.NewHSplit(rawContainer, parsedContainer)
	split.Offset = 0.5 // 50/50 分割

	content := container.NewBorder(
		roomTab.StatsLabel,
		nil, nil, nil,
		split,
	)

	// 创建Tab项
	tabTitle := fmt.Sprintf("🏠 房间 %s", roomID)
	roomTab.Tab = container.NewTabItem(tabTitle, content)

	// 添加到容器
	ui.roomTabs[roomID] = roomTab
	ui.tabContainer.Append(roomTab.Tab)
	ui.tabContainer.Select(roomTab.Tab)

	log.Printf("✅ 房间 Tab 创建成功: %s", roomID)
}

func (roomTab *RoomTab) findParsedIndexByRawID(rawID int64) int {
	for i := len(roomTab.ParsedRecords) - 1; i >= 0; i-- {
		if roomTab.ParsedRecords[i].RawID == rawID {
			return i
		}
	}
	return -1
}

func (roomTab *RoomTab) latestRawID() (int64, bool) {
	if len(roomTab.MessagePairs) == 0 {
		return 0, false
	}
	return roomTab.MessagePairs[len(roomTab.MessagePairs)-1].ID, true
}

func (roomTab *RoomTab) findRawPair(rawID int64) *MessagePair {
	for _, pair := range roomTab.MessagePairs {
		if pair.ID == rawID {
			return pair
		}
	}
	return nil
}

func (roomTab *RoomTab) updateStats(roomID string) {
	if roomTab.StatsLabel == nil {
		return
	}
	roomTab.StatsLabel.SetText(fmt.Sprintf(
		"房间: %s | 原始消息: %d 条 | 解析消息: %d 条",
		roomID,
		len(roomTab.RawData),
		len(roomTab.ParsedRecords),
	))
}

func (ui *FyneUI) appendParsedRecord(roomTab *RoomTab, roomID string, message string, detail map[string]interface{}) {
	timestamp := time.Now()
	summary := fmt.Sprintf("[%s] %s", timestamp.Format("15:04:05"), message)
	record := &ParsedMessageRecord{
		ID:        roomTab.nextParsedID,
		Summary:   summary,
		Detail:    detail,
		Timestamp: timestamp,
	}
	roomTab.nextParsedID++
	if rawID, ok := roomTab.latestRawID(); ok {
		record.RawID = rawID
	}
	roomTab.ParsedRecords = append(roomTab.ParsedRecords, record)
	if len(roomTab.ParsedRecords) > maxStoredMessages {
		roomTab.ParsedRecords = roomTab.ParsedRecords[1:]
	}
	roomTab.updateStats(roomID)
	if roomTab.ParsedMsgs != nil {
		roomTab.ParsedMsgs.Refresh()
		roomTab.ParsedMsgs.ScrollToBottom()
	}
}

// AddRawMessage 添加原始消息
func (ui *FyneUI) AddRawMessage(roomID string, message string) {
	roomTab, exists := ui.roomTabs[roomID]
	if !exists {
		log.Printf("⚠️  房间不存在，自动创建: %s", roomID)
		ui.AddOrUpdateRoom(roomID)
		roomTab = ui.roomTabs[roomID]
	}

	// 添加消息（保留最新100条）
	timestamp := time.Now()
	msg := fmt.Sprintf("[%s] %s", timestamp.Format("15:04:05"), message)

	roomTab.RawData = append(roomTab.RawData, msg)
	if len(roomTab.RawData) > maxStoredMessages {
		roomTab.RawData = roomTab.RawData[1:]
	}

	// 创建新的消息对
	pair := &MessagePair{
		ID:         roomTab.nextRawID,
		RawMessage: message,
		Timestamp:  timestamp,
	}
	roomTab.nextRawID++
	roomTab.MessagePairs = append(roomTab.MessagePairs, pair)
	if len(roomTab.MessagePairs) > maxStoredMessages {
		roomTab.MessagePairs = roomTab.MessagePairs[1:]
	}

	// 刷新UI
	if roomTab.RawMessages != nil {
		roomTab.RawMessages.Refresh()
		roomTab.RawMessages.ScrollToBottom()
	}

	roomTab.updateStats(roomID)
}

// AddParsedMessage 添加解析后的消息
func (ui *FyneUI) AddParsedMessage(roomID string, message string) {
	roomTab, exists := ui.roomTabs[roomID]
	if !exists {
		return
	}

	ui.appendParsedRecord(roomTab, roomID, message, nil)
}

// AddParsedMessageWithDetail 添加解析后的消息（包含详细信息）
func (ui *FyneUI) AddParsedMessageWithDetail(roomID string, message string, detail map[string]interface{}) {
	roomTab, exists := ui.roomTabs[roomID]
	if !exists {
		return
	}

	ui.appendParsedRecord(roomTab, roomID, message, detail)
}

// showMessageDetail 显示消息详情对话框
func (ui *FyneUI) showMessageDetail(roomTab *RoomTab, id widget.ListItemID) {
	if id < 0 || id >= len(roomTab.ParsedRecords) {
		return
	}

	record := roomTab.ParsedRecords[id]
	rawMessage := ""
	var rawTimestamp time.Time
	if record.RawID != 0 {
		if pair := roomTab.findRawPair(record.RawID); pair != nil {
			rawMessage = pair.RawMessage
			rawTimestamp = pair.Timestamp
		}
	}

	detailText := fmt.Sprintf("📅 解析时间: %s\n", record.Timestamp.Format("2006-01-02 15:04:05"))
	if !rawTimestamp.IsZero() {
		detailText += fmt.Sprintf("📡 原始时间: %s\n", rawTimestamp.Format("2006-01-02 15:04:05"))
	}
	detailText += "\n📋 解析后消息:\n" + record.Summary + "\n"
	if rawMessage != "" {
		detailText += "\n📡 原始消息:\n" + rawMessage + "\n"
	}

	if record.Detail != nil && len(record.Detail) > 0 {
		detailText += "\n🔍 详细信息:\n"
		if pretty, err := json.MarshalIndent(record.Detail, "", "  "); err == nil {
			detailText += string(pretty) + "\n"
		} else {
			detailText += fmt.Sprintf("%v\n", record.Detail)
		}
	} else {
		detailText += "\n🔍 详细信息:\n(无结构化数据)\n"
	}

	// 创建详情窗口
	detailWin := ui.app.NewWindow(fmt.Sprintf("消息详情 - 房间 %s", roomTab.RoomID))
	detailWin.Resize(fyne.NewSize(800, 600))
	detailWin.CenterOnScreen()

	// 创建多行文本组件
	detailLabel := widget.NewLabel(detailText)
	detailLabel.Wrapping = fyne.TextWrapWord

	// 创建滚动容器
	scrollContainer := container.NewScroll(detailLabel)

	// 关闭按钮
	closeBtn := widget.NewButton("关闭", func() {
		detailWin.Close()
	})

	// 复制按钮
	copyBtn := widget.NewButton("复制详情", func() {
		detailWin.Clipboard().SetContent(detailText)
		log.Println("✅ 已复制消息详情到剪贴板")
	})

	buttonBar := container.NewHBox(copyBtn, closeBtn)

	content := container.NewBorder(
		nil,
		buttonBar,
		nil,
		nil,
		scrollContainer,
	)

	detailWin.SetContent(content)
	detailWin.Show()
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
