package ui

import (
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/tidwall/gjson"
	"github.com/xuri/excelize/v2"

	"dy-live-monitor/internal/config"
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/parser"
	"dy-live-monitor/internal/server"
)

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

// MessagePair 解析后的消息记录
type MessagePair struct {
	Parsed    *parser.ParsedProtoMessage
	Display   string
	Detail    map[string]interface{}
	Timestamp time.Time
	Source    string
}

type GiftRecord struct {
	ID           int
	GiftID       string
	Name         string
	DiamondValue int
	IconURL      string
	IconLocal    string
	Version      string
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type giftFilter struct {
	Name       string
	DiamondMin int
	DiamondMax int
	SortAsc    bool
	Page       int
	PageSize   int
}

// RoomTab 房间Tab数据
type RoomTab struct {
	RoomID               string
	RoomName             string
	Tab                  *container.TabItem
	MessagesList         *widget.List
	MessagePairs         []*MessagePair
	FilteredPairs        []*MessagePair
	StatsLabel           *widget.Label
	DetailWindow         fyne.Window // 详情窗口
	MessageFilter        string
	TotalMessages        int
	FilterSelect         *widget.Select
	SubTabs              *container.AppTabs
	GiftTable            *widget.Table
	AnchorTable          *widget.Table
	SegmentTable         *widget.Table
	GiftRows             [][]string
	AnchorRows           [][]string
	SegmentRows          [][]string
	AnchorIDEntry        *widget.Entry
	AnchorNameEntry      *widget.Entry
	AnchorGiftCountEntry *widget.Entry
	AnchorScoreEntry     *widget.Entry
	AnchorStatus         *widget.Label
	AnchorPicker         *widget.Select
	AnchorOptionMap      map[string]AnchorOption
}

type AnchorOption struct {
	ID     string
	Name   string
	Avatar string
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

	// 当前选中的房间
	currentRoom string

	// 动态房间 Tabs
	roomTabs     map[string]*RoomTab
	tabContainer *container.AppTabs

	// 手动房间连接
	roomConnMu  sync.Mutex
	manualRooms map[string]*manualRoomConnection

	overviewStatus   *widget.Label
	currentRoomLabel *widget.Label
	userTheme        string
	preferencesPath  string

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
		manualRooms:  make(map[string]*manualRoomConnection),
	}
	ui.preferencesPath = filepath.Join(".", "ui_preferences.json")
	ui.userTheme = ui.loadThemePreference()
	ui.applyTheme(ui.userTheme)

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
	ui.tabContainer = container.NewAppTabs(
		container.NewTabItem("房间管理", ui.createRoomManagementTab()),
		container.NewTabItem("礼物管理", ui.createGiftManagementTab()),
		container.NewTabItem("主播管理", ui.createGlobalAnchorTab()),
		container.NewTabItem("设置", ui.createSettingsTab()),
		container.NewTabItem("数据概览", ui.createOverviewTab()),
	)
	ui.tabContainer.SetTabLocation(container.TabLocationTop)
	return ui.tabContainer
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

	card := container.NewGridWithColumns(len(statsCards), statsCards...)

	return container.NewPadded(card)
}

// createOverviewTab 创建数据概览 Tab
func (ui *FyneUI) createOverviewTab() fyne.CanvasObject {
	roomLabel := widget.NewLabel("当前监控房间: 无")
	ui.currentRoomLabel = roomLabel
	ui.overviewStatus = widget.NewLabel("状态: 等待连接...")

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
		ui.createStatsCard(),
		roomLabel,
		ui.overviewStatus,
		refreshBtn,
		widget.NewSeparator(),
		info,
	)
}

func (ui *FyneUI) createGlobalAnchorTab() fyne.CanvasObject {
	data := ui.loadAllAnchors()

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignTrailing
	statusLabel.Wrapping = fyne.TextWrapOff

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("主播ID")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("主播昵称")
	avatarEntry := widget.NewEntry()
	avatarEntry.SetPlaceHolder("头像路径")
	avatarEntry.Disable()
	deletedCheck := widget.NewCheck("标记删除", nil)

	resetForm := func() {
		idEntry.SetText("")
		nameEntry.SetText("")
		avatarEntry.SetText("")
		deletedCheck.SetChecked(false)
		statusLabel.SetText("")
	}

	table := widget.NewTable(
		func() (int, int) {
			if len(data) == 0 {
				return 0, 0
			}
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row < len(data) && id.Col < len(data[id.Row]) {
				cell.(*widget.Label).SetText(data[id.Row][id.Col])
			}
		},
	)
	table.SetColumnWidth(0, 160)
	table.SetColumnWidth(1, 160)
	table.SetColumnWidth(2, 240)
	table.SetColumnWidth(3, 90)
	table.SetColumnWidth(4, 140)
	table.SetColumnWidth(5, 140)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row >= len(data) {
			return
		}
		row := data[id.Row]
		idEntry.SetText(row[0])
		nameEntry.SetText(row[1])
		avatarEntry.SetText(row[2])
		deletedCheck.SetChecked(row[3] == "是")
	}

	saveBtn := widget.NewButton("保存/更新主播", func() {
		if ui.db == nil {
			return
		}
		id := strings.TrimSpace(idEntry.Text)
		name := strings.TrimSpace(nameEntry.Text)
		if id == "" || name == "" {
			return
		}
		avatar := strings.TrimSpace(avatarEntry.Text)
		deleted := 0
		var deletedAt interface{}
		if deletedCheck.Checked {
			deleted = 1
			deletedAt = time.Now()
		} else {
			deletedAt = nil
		}

		_, err := ui.db.Exec(`
			INSERT INTO anchors (anchor_id, anchor_name, avatar_url, bound_gifts, is_deleted, deleted_at, updated_at)
			VALUES (?, ?, ?, '', ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(anchor_id) DO UPDATE SET 
				anchor_name=excluded.anchor_name,
				avatar_url=excluded.avatar_url,
				is_deleted=excluded.is_deleted,
				deleted_at=excluded.deleted_at,
				updated_at=CURRENT_TIMESTAMP
		`, id, name, avatar, deleted, deletedAt)
		if err != nil {
			log.Printf("⚠️  保存主播失败: %v", err)
			statusLabel.SetText(fmt.Sprintf("保存失败: %v", err))
			return
		}
		resetForm()
		data = ui.loadAllAnchors()
		table.Refresh()
		ui.refreshAllAnchorPickers()
		statusLabel.SetText("✅ 主播信息已保存")
	})

	refreshBtn := widget.NewButton("刷新", func() {
		data = ui.loadAllAnchors()
		table.Refresh()
		ui.refreshAllAnchorPickers()
		statusLabel.SetText("已刷新")
	})

	clearBtn := widget.NewButton("清空", func() {
		resetForm()
	})

	uploadBtn := widget.NewButton("上传头像", func() {
		if ui.mainWin == nil {
			statusLabel.SetText("请先打开主窗口")
			return
		}
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("选择文件失败: %v", err))
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			dataBytes, err := io.ReadAll(reader)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("读取文件失败: %v", err))
				return
			}

			ext := filepath.Ext(reader.URI().Name())
			if ext == "" {
				ext = ".png"
			}
			dstDir := filepath.Join("assets", "anchor_avatars")
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				statusLabel.SetText(fmt.Sprintf("创建目录失败: %v", err))
				return
			}
			filename := reader.URI().Name()
			if trimmed := strings.TrimSpace(idEntry.Text); trimmed != "" {
				filename = trimmed + ext
			}
			dstPath := filepath.Join(dstDir, filename)
			if err := os.WriteFile(dstPath, dataBytes, 0644); err != nil {
				statusLabel.SetText(fmt.Sprintf("保存头像失败: %v", err))
				return
			}
			avatarEntry.Enable()
			avatarEntry.SetText(filepath.ToSlash(dstPath))
			avatarEntry.Disable()
			statusLabel.SetText("头像上传成功")
		}, ui.mainWin)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".gif", ".webp"}))
		fileDialog.Show()
	})

	form := container.NewVBox(
		widget.NewLabel("主播管理"),
		idEntry,
		nameEntry,
		container.NewHBox(avatarEntry, uploadBtn),
		deletedCheck,
		statusLabel,
		container.NewHBox(saveBtn, refreshBtn, clearBtn),
	)

	return container.NewBorder(
		form,
		nil, nil, nil,
		container.NewScroll(table),
	)
}

func (ui *FyneUI) createGiftManagementTab() fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignTrailing
	statusLabel.Wrapping = fyne.TextWrapOff
	const defaultPageSize = 10
	filter := giftFilter{SortAsc: true, Page: 1, PageSize: defaultPageSize}

	nameFilter := widget.NewEntry()
	nameFilter.SetPlaceHolder("礼物名称关键词")
	minDiamondEntry := widget.NewEntry()
	minDiamondEntry.SetPlaceHolder("最小钻石")
	maxDiamondEntry := widget.NewEntry()
	maxDiamondEntry.SetPlaceHolder("最大钻石")

	listContent := container.NewVBox()
	listScroll := container.NewVScroll(listContent)

	pageLabel := widget.NewLabel("")
	pageLabel.Alignment = fyne.TextAlignCenter
	var prevBtn, nextBtn *widget.Button

	var renderList func()
	renderList = func() {
		total := ui.countGiftRecords(filter)
		maxPage := (total + filter.PageSize - 1) / filter.PageSize
		if maxPage == 0 {
			maxPage = 1
		}
		if filter.Page > maxPage {
			filter.Page = maxPage
		}
		if filter.Page < 1 {
			filter.Page = 1
		}

		records := ui.loadGiftRecords(filter)
		listContent.Objects = nil
		if len(records) == 0 {
			empty := widget.NewLabel("暂无礼物数据")
			empty.Alignment = fyne.TextAlignCenter
			listContent.Add(container.NewCenter(empty))
		} else {
			for idx, rec := range records {
				record := rec
				row := ui.buildGiftRow(record,
					func() {
						ui.showGiftEditor(&record, func() {
							statusLabel.SetText("已保存")
							renderList()
						})
					},
					func() {
						if err := ui.setGiftDeleted(record.ID, !record.IsDeleted); err != nil {
							statusLabel.SetText(fmt.Sprintf("操作失败: %v", err))
							return
						}
						if record.IsDeleted {
							statusLabel.SetText("已恢复礼物")
						} else {
							statusLabel.SetText("已删除礼物")
						}
						renderList()
					})
				listContent.Add(row)
				if idx < len(records)-1 {
					listContent.Add(ui.giftRowDivider())
				}
			}
		}
		listContent.Refresh()

		pageLabel.SetText(fmt.Sprintf("第 %d 页 / 共 %d 页 · 共 %d 条", filter.Page, maxPage, total))
		if prevBtn != nil {
			if filter.Page <= 1 {
				prevBtn.Disable()
			} else {
				prevBtn.Enable()
			}
		}
		if nextBtn != nil {
			if filter.Page >= maxPage {
				nextBtn.Disable()
			} else {
				nextBtn.Enable()
			}
		}
	}

	sortBtn := widget.NewButton("钻石排序 ↑", nil)
	sortBtn.OnTapped = func() {
		filter.SortAsc = !filter.SortAsc
		if filter.SortAsc {
			sortBtn.SetText("钻石排序 ↑")
		} else {
			sortBtn.SetText("钻石排序 ↓")
		}
		renderList()
	}
	sortBtn.Importance = widget.MediumImportance

	searchBtn := widget.NewButton("查询", func() {
		filter.Name = strings.TrimSpace(nameFilter.Text)
		filter.DiamondMin = parseTextInt(minDiamondEntry.Text)
		filter.DiamondMax = parseTextInt(maxDiamondEntry.Text)
		filter.Page = 1
		renderList()
	})
	searchBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButton("重置", func() {
		nameFilter.SetText("")
		minDiamondEntry.SetText("")
		maxDiamondEntry.SetText("")
		filter = giftFilter{SortAsc: true, Page: 1, PageSize: defaultPageSize}
		sortBtn.SetText("钻石排序 ↑")
		renderList()
	})

	addBtn := widget.NewButton("新增礼物", func() {
		ui.showGiftEditor(nil, func() {
			statusLabel.SetText("已添加礼物")
			renderList()
		})
	})
	addBtn.Importance = widget.HighImportance

	var latestBtn *widget.Button
	latestBtn = widget.NewButton("更新最新礼物列表", func() {
		if ui.db == nil {
			statusLabel.SetText("数据库未初始化")
			return
		}
		latestBtn.Disable()
		statusLabel.SetText("正在从抖音获取礼物列表...")
		go func() {
			count, err := ui.fetchAndStoreLatestGifts()
			ui.runOnMain(func() {
				latestBtn.Enable()
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("更新失败: %v", err))
					return
				}
				statusLabel.SetText(fmt.Sprintf("已同步 %d 个礼物", count))
				renderList()
			})
		}()
	})

	// 创建筛选输入框 - 使用网格布局均匀分布
	makeFilterField := func(label string, entry *widget.Entry) fyne.CanvasObject {
		lbl := widget.NewLabel(label)
		lbl.Alignment = fyne.TextAlignLeading
		return container.NewVBox(lbl, entry)
	}

	filterBar := container.New(layout.NewGridLayoutWithColumns(3),
		makeFilterField("礼物名称关键词", nameFilter),
		makeFilterField("最小钻石", minDiamondEntry),
		makeFilterField("最大钻石", maxDiamondEntry),
	)

	prevBtn = widget.NewButton("上一页", func() {
		if filter.Page > 1 {
			filter.Page--
			renderList()
		}
	})
	nextBtn = widget.NewButton("下一页", func() {
		filter.Page++
		renderList()
	})

	buttonRow := container.NewHBox(
		searchBtn,
		resetBtn,
		addBtn,
		latestBtn,
		sortBtn,
		layout.NewSpacer(),
		statusLabel,
	)

	// 每页行数设置
	pageSizeEntry := widget.NewEntry()
	pageSizeEntry.SetText(fmt.Sprintf("%d", defaultPageSize))
	pageSizeEntry.SetPlaceHolder("每页行数")

	pageSizeBtn := widget.NewButton("设置", func() {
		newSize := parseTextInt(pageSizeEntry.Text)
		if newSize > 0 && newSize <= 100 {
			filter.PageSize = newSize
			filter.Page = 1 // 重置到第一页
			renderList()
		} else {
			statusLabel.SetText("每页行数必须在 1-100 之间")
		}
	})

	pageSizeBox := container.NewHBox(
		widget.NewLabel("每页显示:"),
		pageSizeEntry,
		pageSizeBtn,
	)

	paginationButtons := container.NewCenter(container.NewHBox(prevBtn, nextBtn))
	paginationBar := container.NewVBox(
		container.NewHBox(
			paginationButtons,
			layout.NewSpacer(),
			pageSizeBox,
		),
		container.NewCenter(pageLabel),
	)

	renderList()

	headerRow := ui.buildGiftHeaderRow()

	// 列表区域 - 表头 + 分隔符 + 滚动列表
	listArea := container.NewBorder(
		container.NewVBox(headerRow, widget.NewSeparator()), // 顶部：表头
		nil, nil, nil,
		listScroll, // 中间：滚动列表自动扩展
	)

	// 顶部区域
	topSection := container.NewVBox(filterBar, buttonRow, widget.NewSeparator())

	// 主布局：顶部固定，底部固定，中间列表自动扩展
	mainContent := container.NewBorder(
		topSection,    // 顶部：筛选 + 按钮
		paginationBar, // 底部：分页控件
		nil, nil,
		listArea, // 中间：列表区域自动扩展填充
	)

	return mainContent
}

func (ui *FyneUI) createRoomManagementTab() fyne.CanvasObject {
	statusLabel := widget.NewLabel("")

	// 手动添加房间区域
	manualRoomEntry := widget.NewEntry()
	manualRoomEntry.SetPlaceHolder("输入抖音房间号 (短号或 room_id)")

	addRoomBtn := widget.NewButton("手动添加房间", func() {
		roomID := strings.TrimSpace(manualRoomEntry.Text)
		if roomID == "" {
			statusLabel.SetText("❌ 房间号不能为空")
			return
		}

		manualRoomEntry.SetText("")
		statusLabel.SetText(fmt.Sprintf("⏳ 正在连接房间 %s...", roomID))

		go func(id string) {
			if err := ui.startManualRoom(id); err != nil {
				log.Printf("❌ 启动房间 %s 失败: %v", id, err)
				ui.runOnMain(func() {
					statusLabel.SetText(fmt.Sprintf("❌ 房间 %s 连接失败: %v", id, err))
				})
			} else {
				ui.runOnMain(func() {
					statusLabel.SetText(fmt.Sprintf("✅ 正在监听房间 %s", id))
					ui.updateOverviewStatus(fmt.Sprintf("状态: 正在监听房间 %s", id))
				})
			}
		}(roomID)
	})
	addRoomBtn.Importance = widget.HighImportance

	manualRoomSection := container.NewVBox(
		widget.NewLabel("📡 手动添加直播间"),
		widget.NewLabel("无需浏览器插件，直接建立 WSS 连接获取直播消息"),
		container.NewBorder(nil, nil, nil, addRoomBtn, manualRoomEntry),
		widget.NewSeparator(),
	)

	// 历史房间查询区域
	roomFilter := widget.NewEntry()
	roomFilter.SetPlaceHolder("房间号")
	anchorFilter := widget.NewEntry()
	anchorFilter.SetPlaceHolder("主播名称")

	type roomSummary struct {
		ID      string
		Title   string
		Display string
	}

	data := ui.loadRoomSummaries("", "")
	summaries := make([]roomSummary, len(data))
	for i, row := range data {
		summaries[i] = roomSummary{ID: row[0], Title: row[1], Display: strings.Join(row, " | ")}
	}

	updateStatusLabel := func() {
		if statusLabel.Text == "" || strings.HasPrefix(statusLabel.Text, "共") {
			statusLabel.SetText(fmt.Sprintf("共 %d 条记录", len(summaries)))
		}
	}
	updateStatusLabel()

	roomList := widget.NewList(
		func() int { return len(summaries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, co fyne.CanvasObject) {
			if id < len(summaries) {
				co.(*widget.Label).SetText(summaries[id].Display)
			}
		},
	)

	selected := -1
	roomList.OnSelected = func(id widget.ListItemID) {
		selected = int(id)
	}

	queryBtn := widget.NewButton("查询历史房间", func() {
		data = ui.loadRoomSummaries(roomFilter.Text, anchorFilter.Text)
		summaries = make([]roomSummary, len(data))
		for i, row := range data {
			summaries[i] = roomSummary{ID: row[0], Title: row[1], Display: strings.Join(row, " | ")}
		}
		roomList.Refresh()
		selected = -1
		statusLabel.SetText(fmt.Sprintf("共 %d 条记录", len(summaries)))
	})

	openBtn := widget.NewButton("打开房间详情", func() {
		if selected >= 0 && selected < len(summaries) {
			ui.openHistoricalRoomTab(summaries[selected].ID)
			statusLabel.SetText(fmt.Sprintf("已打开房间 %s", summaries[selected].ID))
		} else {
			statusLabel.SetText("请先选择房间")
		}
	})

	exportGiftsBtn := widget.NewButton("导出礼物记录", func() {
		if selected >= 0 && selected < len(summaries) {
			path, err := ui.exportRoomGifts(summaries[selected].ID)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("导出失败: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("礼物记录已导出到 %s", path))
			}
		} else {
			statusLabel.SetText("请先选择房间")
		}
	})

	exportAnchorsBtn := widget.NewButton("导出主播得分", func() {
		if selected >= 0 && selected < len(summaries) {
			path, err := ui.exportRoomAnchorScores(summaries[selected].ID)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("导出失败: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("主播得分已导出到 %s", path))
			}
		} else {
			statusLabel.SetText("请先选择房间")
		}
	})

	historySection := container.NewVBox(
		widget.NewLabel("📋 历史房间查询"),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("房间号"), roomFilter),
			container.NewVBox(widget.NewLabel("主播"), anchorFilter),
		),
		container.NewHBox(queryBtn, openBtn, exportGiftsBtn, exportAnchorsBtn),
		widget.NewSeparator(),
	)

	topSection := container.NewVBox(
		manualRoomSection,
		historySection,
		statusLabel,
	)

	return container.NewBorder(
		topSection,
		nil, nil, nil,
		container.NewScroll(roomList),
	)
}

func (ui *FyneUI) loadAllAnchors() [][]string {
	rows := [][]string{{"主播ID", "主播昵称", "头像", "已删除", "添加时间", "删除时间"}}
	if ui.db == nil {
		return rows
	}

	query := `
		SELECT anchor_id, anchor_name, COALESCE(avatar_url, ''), COALESCE(is_deleted, 0),
		       created_at, deleted_at
		FROM anchors
		ORDER BY updated_at DESC
	`
	data, err := ui.db.Query(query)
	if err != nil {
		return rows
	}
	defer data.Close()

	for data.Next() {
		var id, name, avatar string
		var created time.Time
		var deleted sql.NullTime
		var isDeleted int
		if err := data.Scan(&id, &name, &avatar, &isDeleted, &created, &deleted); err != nil {
			continue
		}
		deletedStr := ""
		if deleted.Valid {
			deletedStr = deleted.Time.Format("01-02 15:04")
		}
		rows = append(rows, []string{
			id,
			name,
			avatar,
			formatBoolLabel(isDeleted == 1),
			created.Format("01-02 15:04"),
			deletedStr,
		})
	}
	return rows
}

func formatBoolLabel(val bool) string {
	if val {
		return "是"
	}
	return "否"
}

func (ui *FyneUI) loadAnchorOptions(includeDeleted bool) []AnchorOption {
	options := make([]AnchorOption, 0)
	if ui.db == nil {
		return options
	}
	query := `SELECT anchor_id, anchor_name, COALESCE(avatar_url, '') FROM anchors`
	if !includeDeleted {
		query += ` WHERE COALESCE(is_deleted, 0) = 0`
	}
	query += ` ORDER BY anchor_name`
	rows, err := ui.db.Query(query)
	if err != nil {
		return options
	}
	defer rows.Close()

	for rows.Next() {
		var opt AnchorOption
		if err := rows.Scan(&opt.ID, &opt.Name, &opt.Avatar); err != nil {
			continue
		}
		options = append(options, opt)
	}
	return options
}

func (ui *FyneUI) refreshRoomAnchorPicker(roomTab *RoomTab) {
	if roomTab == nil || roomTab.AnchorPicker == nil {
		return
	}
	options := ui.loadAnchorOptions(false)
	labels := make([]string, 0, len(options))
	roomTab.AnchorOptionMap = make(map[string]AnchorOption, len(options))
	for _, opt := range options {
		label := fmt.Sprintf("%s | %s", opt.ID, opt.Name)
		labels = append(labels, label)
		roomTab.AnchorOptionMap[label] = opt
	}
	roomTab.AnchorPicker.Options = labels
	roomTab.AnchorPicker.Selected = ""
	roomTab.AnchorPicker.Refresh()
}

func (ui *FyneUI) refreshAllAnchorPickers() {
	if ui.roomTabs == nil {
		return
	}
	for _, tab := range ui.roomTabs {
		ui.refreshRoomAnchorPicker(tab)
	}
}

func (ui *FyneUI) loadAllGifts() [][]string {
	rows := [][]string{{"礼物ID", "礼物名称", "钻石", "版本号", "更新时间"}}
	if ui.db == nil {
		return rows
	}

	query := `
		SELECT gift_id, gift_name, diamond_value, version, updated_at
		FROM gifts
		WHERE COALESCE(is_deleted, 0) = 0
		ORDER BY updated_at DESC
	`
	data, err := ui.db.Query(query)
	if err != nil {
		return rows
	}
	defer data.Close()

	for data.Next() {
		var id, name, version string
		var diamond int
		var updated time.Time
		if err := data.Scan(&id, &name, &diamond, &version, &updated); err != nil {
			continue
		}
		rows = append(rows, []string{
			id,
			name,
			fmt.Sprintf("%d", diamond),
			version,
			updated.Format("01-02 15:04"),
		})
	}
	return rows
}

const (
	douyinGiftListAPI   = "https://live.douyin.com/webcast/gift/list/?device_platform=webapp&aid=6383"
	giftIconStoragePath = "assets/gift_icons"
)

type douyinGiftItem struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	DiamondCount int             `json:"diamond_count"`
	Icon         douyinGiftIcon  `json:"icon"`
	Picture      douyinGiftIcon  `json:"picture"`
	Describe     string          `json:"describe"`
	GiftLabel    json.RawMessage `json:"gift_label"`
}

type douyinGiftIcon struct {
	URLList []string `json:"url_list"`
	URI     string   `json:"uri"`
}

func (icon douyinGiftIcon) FirstURL() string {
	for _, url := range icon.URLList {
		if trimmed := strings.TrimSpace(url); trimmed != "" {
			return trimmed
		}
	}
	if strings.TrimSpace(icon.URI) != "" {
		if strings.HasPrefix(icon.URI, "http") {
			return icon.URI
		}
		return "https://p3-webcast.douyinpic.com/" + strings.TrimLeft(icon.URI, "/")
	}
	return ""
}

func (ui *FyneUI) fetchAndStoreLatestGifts() (int, error) {
	if ui.db == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, douyinGiftListAPI, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("抖音接口返回状态 %d", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取礼物数据失败: %w", err)
	}

	bodyStr := strings.TrimSpace(string(rawBody))
	if bodyStr == "" || (!strings.HasPrefix(bodyStr, "{") && !strings.HasPrefix(bodyStr, "[")) {
		return 0, fmt.Errorf("礼物接口返回的不是 JSON 数据: %s", truncateString(bodyStr, 64))
	}

	if ui.cfg != nil && ui.cfg.Debug.VerboseLog {
		log.Printf("🧾 礼物接口原始 body: %s", truncateString(bodyStr, 256))
	}

	giftsArray := gjson.Get(bodyStr, "data.gifts")
	if !giftsArray.Exists() || !giftsArray.IsArray() {
		return 0, fmt.Errorf("礼物数据缺少 data.gifts 数组")
	}

	giftItems := make([]douyinGiftItem, 0, len(giftsArray.Array()))
	for _, item := range giftsArray.Array() {
		if !item.Exists() || !item.IsObject() {
			continue
		}
		var parsed douyinGiftItem
		if err := json.Unmarshal([]byte(item.Raw), &parsed); err != nil {
			log.Printf("⚠️  解析礼物对象失败: %v", err)
			continue
		}
		giftItems = append(giftItems, parsed)
	}
	log.Printf("ℹ️  抓取礼物列表 gift_items 条数: %d", len(giftItems))
	if len(giftItems) > 0 {
		if firstJSON, err := json.Marshal(giftItems[0]); err == nil {
			log.Printf("ℹ️  gift_items 第一个对象: %s", string(firstJSON))
		} else {
			log.Printf("ℹ️  gift_items 第一个对象解析失败: %v", err)
		}
	}
	if len(giftItems) == 0 {
		return 0, fmt.Errorf("未获取到礼物数据")
	}

	if err := os.MkdirAll(giftIconStoragePath, 0755); err != nil {
		return 0, err
	}

	tx, err := ui.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, gift := range giftItems {
		giftID := strconv.FormatInt(gift.ID, 10)
		iconURL := gift.Icon.FirstURL()
		if iconURL == "" {
			iconURL = gift.Picture.FirstURL()
		}
		iconPath := ""
		if iconURL != "" {
			path, err := ui.downloadGiftIcon(giftID, iconURL)
			if err != nil {
				log.Printf("⚠️  下载礼物图标失败(%s): %v", giftID, err)
			} else {
				iconPath = path
			}
		}

		_, err := tx.Exec(`
			INSERT INTO gifts (gift_id, gift_name, diamond_value, icon_url, icon_local, version, is_deleted)
			VALUES (?, ?, ?, ?, ?, ?, 0)
			ON CONFLICT(gift_id) DO UPDATE SET 
				gift_name=excluded.gift_name,
				diamond_value=excluded.diamond_value,
				icon_url=excluded.icon_url,
				icon_local=excluded.icon_local,
				version=excluded.version,
				is_deleted=0,
				updated_at=CURRENT_TIMESTAMP
		`, giftID, strings.TrimSpace(gift.Name), gift.DiamondCount, iconURL, iconPath, "douyin_api")
		if err != nil {
			log.Printf("⚠️  保存礼物 %s 失败: %v", giftID, err)
			continue
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return inserted, nil
}

func (ui *FyneUI) downloadGiftIcon(giftID string, rawURL string) (string, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", nil
	}
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载礼物图标失败: %s", resp.Status)
	}

	ext := filepath.Ext(strings.Split(rawURL, "?")[0])
	if ext == "" || len(ext) > 5 {
		ext = ".png"
	}

	if err := os.MkdirAll(giftIconStoragePath, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(giftIconStoragePath, fmt.Sprintf("%s%s", giftID, ext))
	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	return filepath.ToSlash(fullPath), nil
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func parseTextInt(text string) int {
	value, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0
	}
	return value
}

func formatDisplayTime(t time.Time) string {
	if t.IsZero() {
		return "--"
	}
	return t.Format("01-02 15:04")
}

func (ui *FyneUI) loadGiftRecords(filter giftFilter) []GiftRecord {
	records := make([]GiftRecord, 0)
	if ui.db == nil {
		return records
	}

	whereClause, args := buildGiftWhereClause(filter)
	orderClause := " ORDER BY diamond_value ASC, updated_at DESC"
	if !filter.SortAsc {
		orderClause = " ORDER BY diamond_value DESC, updated_at DESC"
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, gift_id, gift_name, diamond_value, icon_url, icon_local, version,
		       COALESCE(is_deleted, 0), created_at, updated_at
		FROM gifts
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, orderClause)
	args = append(args, pageSize, offset)

	rows, err := ui.db.Query(query, args...)
	if err != nil {
		return records
	}
	defer rows.Close()

	for rows.Next() {
		var rec GiftRecord
		var created, updated sql.NullTime
		var isDeleted int
		if err := rows.Scan(&rec.ID, &rec.GiftID, &rec.Name, &rec.DiamondValue, &rec.IconURL, &rec.IconLocal, &rec.Version, &isDeleted, &created, &updated); err != nil {
			continue
		}
		rec.IsDeleted = isDeleted == 1
		if created.Valid {
			rec.CreatedAt = created.Time
		}
		if updated.Valid {
			rec.UpdatedAt = updated.Time
		}
		records = append(records, rec)
	}
	return records
}

func (ui *FyneUI) saveGiftRecord(rec *GiftRecord) error {
	if ui.db == nil || rec == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if rec.ID > 0 {
		_, err := ui.db.Exec(`
			UPDATE gifts
			SET gift_id = ?, gift_name = ?, diamond_value = ?, icon_url = ?, icon_local = ?, version = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, rec.GiftID, rec.Name, rec.DiamondValue, rec.IconURL, rec.IconLocal, rec.Version, rec.ID)
		return err
	}

	_, err := ui.db.Exec(`
		INSERT INTO gifts (gift_id, gift_name, diamond_value, icon_url, icon_local, version, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(gift_id) DO UPDATE SET
			gift_name=excluded.gift_name,
			diamond_value=excluded.diamond_value,
			icon_url=excluded.icon_url,
			icon_local=excluded.icon_local,
			version=excluded.version,
			is_deleted=excluded.is_deleted,
			updated_at=CURRENT_TIMESTAMP
	`, rec.GiftID, rec.Name, rec.DiamondValue, rec.IconURL, rec.IconLocal, rec.Version, boolToInt(rec.IsDeleted))
	return err
}

func (ui *FyneUI) setGiftDeleted(id int, deleted bool) error {
	if ui.db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := ui.db.Exec(`UPDATE gifts SET is_deleted = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, boolToInt(deleted), id)
	return err
}

func (ui *FyneUI) showGiftEditor(existing *GiftRecord, onSaved func()) {
	if ui.mainWin == nil {
		return
	}

	isEdit := existing != nil
	title := "新增礼物"
	if isEdit {
		title = "编辑礼物"
	}

	giftIDEntry := widget.NewEntry()
	nameEntry := widget.NewEntry()
	diamondEntry := widget.NewEntry()
	versionEntry := widget.NewEntry()
	iconURLEntry := widget.NewEntry()
	iconLocalEntry := widget.NewEntry()
	iconLocalEntry.Disable()
	statusLabel := widget.NewLabel("")
	preview := canvas.NewImageFromResource(theme.DocumentIcon())
	preview.FillMode = canvas.ImageFillContain
	preview.SetMinSize(fyne.NewSize(150, 150))

	updatePreview := func() {
		path := strings.TrimSpace(iconLocalEntry.Text)
		if fileExists(path) {
			preview.File = path
			preview.Resource = nil
		} else {
			preview.File = ""
			preview.Resource = theme.DocumentIcon()
		}
		preview.Refresh()
	}

	if isEdit {
		giftIDEntry.SetText(existing.GiftID)
		giftIDEntry.Disable()
		nameEntry.SetText(existing.Name)
		diamondEntry.SetText(fmt.Sprintf("%d", existing.DiamondValue))
		versionEntry.SetText(existing.Version)
		iconURLEntry.SetText(existing.IconURL)
		iconLocalEntry.Enable()
		iconLocalEntry.SetText(existing.IconLocal)
		iconLocalEntry.Disable()
		updatePreview()
	}

	uploadBtn := widget.NewButton("上传图标", func() {
		if ui.mainWin == nil {
			return
		}
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("选择文件失败: %v", err))
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			dataBytes, err := io.ReadAll(reader)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("读取文件失败: %v", err))
				return
			}

			ext := filepath.Ext(reader.URI().Name())
			if ext == "" {
				ext = ".png"
			}
			if err := os.MkdirAll(giftIconStoragePath, 0755); err != nil {
				statusLabel.SetText(fmt.Sprintf("创建目录失败: %v", err))
				return
			}
			fileName := fmt.Sprintf("manual_%d%s", time.Now().UnixNano(), ext)
			if strings.TrimSpace(giftIDEntry.Text) != "" {
				fileName = fmt.Sprintf("%s%s", strings.TrimSpace(giftIDEntry.Text), ext)
			}
			dstPath := filepath.Join(giftIconStoragePath, fileName)
			if err := os.WriteFile(dstPath, dataBytes, 0644); err != nil {
				statusLabel.SetText(fmt.Sprintf("保存图标失败: %v", err))
				return
			}
			iconLocalEntry.Enable()
			iconLocalEntry.SetText(filepath.ToSlash(dstPath))
			iconLocalEntry.Disable()
			statusLabel.SetText("图标上传成功")
			updatePreview()
		}, ui.mainWin).Show()
	})

	form := container.NewVBox(
		widget.NewLabel("礼物ID"),
		giftIDEntry,
		widget.NewLabel("礼物名称"),
		nameEntry,
		widget.NewLabel("钻石数"),
		diamondEntry,
		widget.NewLabel("版本号"),
		versionEntry,
		widget.NewLabel("图标链接"),
		iconURLEntry,
		widget.NewLabel("本地图标"),
		container.NewBorder(nil, nil, nil, uploadBtn, iconLocalEntry),
		container.NewCenter(preview),
		statusLabel,
	)

	scroll := container.NewVScroll(form)
	scroll.SetMinSize(fyne.NewSize(480, 400))

	var giftDialog dialog.Dialog
	giftDialog = dialog.NewCustomConfirm(title, "保存", "取消", scroll, func(ok bool) {
		if !ok {
			return
		}
		rec := &GiftRecord{
			GiftID:       strings.TrimSpace(giftIDEntry.Text),
			Name:         strings.TrimSpace(nameEntry.Text),
			DiamondValue: parseTextInt(diamondEntry.Text),
			Version:      strings.TrimSpace(versionEntry.Text),
			IconURL:      strings.TrimSpace(iconURLEntry.Text),
			IconLocal:    strings.TrimSpace(iconLocalEntry.Text),
		}
		if rec.GiftID == "" || rec.Name == "" {
			statusLabel.SetText("礼物ID和名称不能为空")
			return
		}
		if rec.DiamondValue < 0 {
			statusLabel.SetText("钻石数必须为正数")
			return
		}
		if isEdit {
			rec.ID = existing.ID
			rec.IsDeleted = existing.IsDeleted
		}
		if err := ui.saveGiftRecord(rec); err != nil {
			statusLabel.SetText(fmt.Sprintf("保存失败: %v", err))
			return
		}
		if onSaved != nil {
			onSaved()
		}
		giftDialog.Hide()
	}, ui.mainWin)
	giftDialog.Resize(fyne.NewSize(520, 560))
	giftDialog.Show()
}

func (ui *FyneUI) buildGiftRow(rec GiftRecord, onEdit func(), onToggleDeleted func()) fyne.CanvasObject {
	// 1. ID列 - 居中显示
	idLabel := widget.NewLabel(rec.GiftID)
	idLabel.Alignment = fyne.TextAlignCenter
	idLabel.Wrapping = fyne.TextWrapOff
	idCell := container.NewCenter(idLabel)

	// 2. 名称列 - 左对齐，加粗
	nameLabel := widget.NewLabel(rec.Name)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Wrapping = fyne.TextWrapOff
	nameLabel.Truncation = fyne.TextTruncateEllipsis
	nameLabel.Alignment = fyne.TextAlignLeading
	nameCell := container.NewPadded(nameLabel)

	// 3. 图标列 - 居中显示
	icon := canvas.NewImageFromResource(theme.DocumentIcon())
	if fileExists(rec.IconLocal) {
		icon = canvas.NewImageFromFile(rec.IconLocal)
	}
	icon.SetMinSize(fyne.NewSize(32, 32))
	icon.FillMode = canvas.ImageFillContain
	iconCell := container.NewCenter(icon)

	// 4. 钻石数列 - 居中显示
	diamondLabel := widget.NewLabel(fmt.Sprintf("%d", rec.DiamondValue))
	diamondLabel.Alignment = fyne.TextAlignCenter
	diamondLabel.Wrapping = fyne.TextWrapOff
	diamondCell := container.NewCenter(diamondLabel)

	// 5. 版本号列 - 居中显示
	versionLabel := widget.NewLabel(rec.Version)
	versionLabel.Alignment = fyne.TextAlignCenter
	versionLabel.Wrapping = fyne.TextWrapOff
	versionLabel.Truncation = fyne.TextTruncateEllipsis
	versionCell := container.NewCenter(versionLabel)

	// 6. 更新时间列 - 右对齐
	timeLabel := widget.NewLabel(formatDisplayTime(rec.UpdatedAt))
	timeLabel.Alignment = fyne.TextAlignTrailing
	timeLabel.Wrapping = fyne.TextWrapOff
	timeCell := container.NewPadded(timeLabel)

	// 7. 操作列 - 按钮横向排列
	editBtn := widget.NewButton("编辑", func() {
		if onEdit != nil {
			onEdit()
		}
	})
	editBtn.Importance = widget.LowImportance

	deleteLabel := "删除"
	if rec.IsDeleted {
		deleteLabel = "恢复"
	}
	deleteBtn := widget.NewButton(deleteLabel, func() {
		if onToggleDeleted != nil {
			onToggleDeleted()
		}
	})
	deleteBtn.Importance = widget.LowImportance

	actionBox := container.NewHBox(editBtn, deleteBtn)

	// 使用网格布局，7列：ID、名称、图标、钻石、版本号、更新时间、操作
	grid := container.New(layout.NewGridLayoutWithColumns(7),
		idCell,
		nameCell,
		iconCell,
		diamondCell,
		versionCell,
		timeCell,
		actionBox,
	)

	// 使用卡片样式，跟随主题
	card := widget.NewCard("", "", grid)
	return card
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type giftPalette struct {
	Text             color.Color
	MutedText        color.Color
	ListBackground   color.Color
	HeaderBackground color.Color
	HeaderText       color.Color
	RowBackground    color.Color
	RowBorder        color.Color
	RowDivider       color.Color
	EntryBackground  color.Color
	EntryBorder      color.Color
	ButtonBackground color.Color
	ButtonHover      color.Color
	ButtonText       color.Color
	ButtonBorder     color.Color
	PagerBackground  color.Color
	PagerHover       color.Color
	PagerText        color.Color
}

func (ui *FyneUI) giftPalette() giftPalette {
	if ui.isDarkThemeVariant() {
		return giftPalette{
			Text:             color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
			MutedText:        color.NRGBA{R: 0xB5, G: 0xB5, B: 0xB5, A: 0xFF},
			ListBackground:   color.NRGBA{R: 0x22, G: 0x24, B: 0x2A, A: 0xFF},
			HeaderBackground: color.NRGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF},
			HeaderText:       color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			RowBackground:    color.NRGBA{R: 0x2C, G: 0x2E, B: 0x36, A: 0xFF},
			RowBorder:        color.NRGBA{R: 0x46, G: 0x4B, B: 0x55, A: 0xFF},
			RowDivider:       color.NRGBA{R: 0x2A, G: 0x2A, B: 0x2A, A: 0xFF},
			EntryBackground:  color.NRGBA{R: 0x2D, G: 0x2D, B: 0x2D, A: 0xFF},
			EntryBorder:      color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xFF},
			ButtonBackground: color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
			ButtonHover:      color.NRGBA{R: 0x4A, G: 0x4A, B: 0x4A, A: 0xFF},
			ButtonText:       color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			ButtonBorder:     color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xFF},
			PagerBackground:  color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
			PagerHover:       color.NRGBA{R: 0x4A, G: 0x4A, B: 0x4A, A: 0xFF},
			PagerText:        color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		}
	}
	return giftPalette{
		Text:             color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
		MutedText:        color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xFF},
		ListBackground:   color.NRGBA{R: 0xF7, G: 0xF8, B: 0xFC, A: 0xFF},
		HeaderBackground: color.NRGBA{R: 0xEE, G: 0xEE, B: 0xEE, A: 0xFF},
		HeaderText:       color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
		RowBackground:    color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		RowBorder:        color.NRGBA{R: 0xDD, G: 0xDD, B: 0xDD, A: 0xFF},
		RowDivider:       color.NRGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0xFF},
		EntryBackground:  color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF},
		EntryBorder:      color.NRGBA{R: 0xDD, G: 0xDD, B: 0xDD, A: 0xFF},
		ButtonBackground: color.NRGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF},
		ButtonHover:      color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
		ButtonText:       color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
		ButtonBorder:     color.NRGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0xFF},
		PagerBackground:  color.NRGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF},
		PagerHover:       color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
		PagerText:        color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
	}
}

func (ui *FyneUI) fixedSpacer(width float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.NRGBA{A: 0})
	rect.SetMinSize(fyne.NewSize(width, 1))
	return rect
}

func (ui *FyneUI) giftEntryField(entry *widget.Entry, width float32) fyne.CanvasObject {
	// 移除自定义背景，让 Entry 使用 Fyne 默认主题
	// 这样可以确保主题切换时 Entry 也会正确响应
	// 注意：Container 没有 SetMinSize 方法，直接返回包装后的 entry
	wrapper := container.NewPadded(entry)
	return wrapper
}

func (ui *FyneUI) giftTableCell(text string, align fyne.TextAlign, bold bool) fyne.CanvasObject {
	lbl := widget.NewLabel(text)
	lbl.Alignment = align
	lbl.Wrapping = fyne.TextWrapOff
	lbl.Truncation = fyne.TextTruncateEllipsis
	if bold {
		lbl.TextStyle = fyne.TextStyle{Bold: true}
	}

	// 使用简单的容器包装，确保文本水平显示
	return container.NewPadded(lbl)
}

func (ui *FyneUI) giftHeaderCell(text string, align fyne.TextAlign) fyne.CanvasObject {
	palette := ui.giftPalette()
	lbl := canvas.NewText(text, palette.HeaderText)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.Alignment = align
	lbl.TextSize = theme.TextSize()
	switch align {
	case fyne.TextAlignTrailing:
		return container.NewHBox(layout.NewSpacer(), lbl)
	case fyne.TextAlignCenter:
		return container.NewHBox(layout.NewSpacer(), lbl, layout.NewSpacer())
	default:
		return container.NewHBox(lbl, layout.NewSpacer())
	}
}

func (ui *FyneUI) giftRowDivider() fyne.CanvasObject {
	// 使用标准分隔符以跟随主题
	return widget.NewSeparator()
}

type giftButton struct {
	widget.BaseWidget
	ui       *FyneUI
	text     string
	minWidth float32
	onTapped func()
	hover    bool
	disabled bool
}

func (ui *FyneUI) newGiftButton(text string, minWidth float32, tapped func()) *giftButton {
	btn := &giftButton{
		ui:       ui,
		text:     text,
		minWidth: minWidth,
		onTapped: tapped,
	}
	if btn.minWidth <= 0 {
		btn.minWidth = 96
	}
	btn.ExtendBaseWidget(btn)
	return btn
}

func (b *giftButton) SetText(text string) {
	if b.text == text {
		return
	}
	b.text = text
	b.Refresh()
}

func (b *giftButton) SetMinWidth(width float32) {
	if width <= 0 {
		width = 60
	}
	b.minWidth = width
	b.Refresh()
}

func (b *giftButton) Disable() {
	b.disabled = true
	b.hover = false
	b.Refresh()
}

func (b *giftButton) Enable() {
	b.disabled = false
	b.Refresh()
}

func (b *giftButton) Disabled() bool {
	return b.disabled
}

func (b *giftButton) MinSize() fyne.Size {
	height := float32(36)
	width := b.minWidth
	if width < 60 {
		width = 60
	}
	return fyne.NewSize(width, height)
}

func (b *giftButton) Tapped(*fyne.PointEvent) {
	if b.Disabled() {
		return
	}
	if b.onTapped != nil {
		b.onTapped()
	}
}

func (b *giftButton) MouseIn(*desktop.MouseEvent) {
	if b.Disabled() {
		return
	}
	b.hover = true
	b.Refresh()
}

func (b *giftButton) MouseMoved(*desktop.MouseEvent) {}

func (b *giftButton) MouseOut(*desktop.MouseEvent) {
	if b.hover {
		b.hover = false
		b.Refresh()
	}
}

func (b *giftButton) CreateRenderer() fyne.WidgetRenderer {
	palette := b.ui.giftPalette()
	bg := canvas.NewRectangle(palette.ButtonBackground)
	bg.CornerRadius = 8
	bg.StrokeWidth = 1
	bg.StrokeColor = palette.ButtonBorder

	label := canvas.NewText(b.text, palette.ButtonText)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = theme.TextSize()

	content := container.NewCenter(label)

	return &giftButtonRenderer{
		button:     b,
		background: bg,
		label:      label,
		content:    content,
	}
}

type giftButtonRenderer struct {
	button     *giftButton
	background *canvas.Rectangle
	label      *canvas.Text
	content    *fyne.Container
}

func (r *giftButtonRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *giftButtonRenderer) MinSize() fyne.Size {
	return r.button.MinSize()
}

func (r *giftButtonRenderer) Refresh() {
	palette := r.button.ui.giftPalette()
	bgColor := palette.ButtonBackground
	if r.button.Disabled() {
		bgColor = palette.ButtonBackground
	} else if r.button.hover {
		bgColor = palette.ButtonHover
	}
	r.background.FillColor = bgColor
	r.background.StrokeColor = palette.ButtonBorder
	textColor := palette.ButtonText
	if r.button.Disabled() {
		textColor = palette.MutedText
	}
	r.label.Text = r.button.text
	r.label.Color = textColor
	canvas.Refresh(r.background)
	canvas.Refresh(r.label)
}

func (r *giftButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.content}
}

func (r *giftButtonRenderer) Destroy() {}

func (ui *FyneUI) isDarkThemeVariant() bool {
	if ui.app == nil {
		return ui.userTheme == "深色"
	}
	settings := ui.app.Settings()
	if settings == nil {
		return ui.userTheme == "深色"
	}
	return settings.ThemeVariant() == theme.VariantDark
}

func (ui *FyneUI) giftListBackgroundColor() color.Color {
	return ui.giftPalette().ListBackground
}

func (ui *FyneUI) giftHeaderBackgroundColor() color.Color {
	return ui.giftPalette().HeaderBackground
}

func (ui *FyneUI) giftRowBackgroundColor() color.Color {
	return ui.giftPalette().RowBackground
}

func (ui *FyneUI) giftRowBorderColor() color.Color {
	return ui.giftPalette().RowBorder
}

func (ui *FyneUI) buildGiftHeaderRow() fyne.CanvasObject {
	// 7列：ID、名称、图标、钻石、版本号、更新时间、操作
	headers := []string{"ID", "名称", "图标", "钻石", "版本号", "更新时间", "操作"}
	cells := make([]fyne.CanvasObject, 0, len(headers))

	// 为每个表头创建标签，使用标准 widget 以跟随主题
	for _, h := range headers {
		lbl := widget.NewLabel(h)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		lbl.Wrapping = fyne.TextWrapOff

		switch h {
		case "名称":
			lbl.Alignment = fyne.TextAlignLeading
		case "版本号", "更新时间":
			lbl.Alignment = fyne.TextAlignTrailing
		default:
			lbl.Alignment = fyne.TextAlignCenter
		}
		cells = append(cells, container.NewPadded(lbl))
	}

	row := container.New(layout.NewGridLayoutWithColumns(len(headers)), cells...)
	return container.NewPadded(row)
}

func buildGiftWhereClause(filter giftFilter) (string, []interface{}) {
	clauses := []string{"COALESCE(is_deleted, 0) = 0"}
	args := make([]interface{}, 0)
	if strings.TrimSpace(filter.Name) != "" {
		clauses = append(clauses, "gift_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(filter.Name)+"%")
	}
	if filter.DiamondMin > 0 {
		clauses = append(clauses, "diamond_value >= ?")
		args = append(args, filter.DiamondMin)
	}
	if filter.DiamondMax > 0 {
		clauses = append(clauses, "diamond_value <= ?")
		args = append(args, filter.DiamondMax)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	return where, args
}

func (ui *FyneUI) countGiftRecords(filter giftFilter) int {
	if ui.db == nil {
		return 0
	}
	where, args := buildGiftWhereClause(filter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM gifts %s`, where)
	var total int
	if err := ui.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0
	}
	return total
}

func (ui *FyneUI) loadRoomSummaries(roomID, anchor string) [][]string {
	rows := [][]string{{"房间号", "标题", "首次出现", "最近活动"}}
	if ui.db == nil {
		return rows
	}

	query := `SELECT room_id, COALESCE(room_title,''), first_seen_at, last_seen_at FROM rooms`
	var args []interface{}
	clauses := []string{}

	if roomID != "" {
		clauses = append(clauses, "room_id LIKE ?")
		args = append(args, "%"+roomID+"%")
	}

	if anchor != "" {
		clauses = append(clauses, "room_id IN (SELECT room_id FROM room_anchors WHERE anchor_name LIKE ?)")
		args = append(args, "%"+anchor+"%")
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY last_seen_at DESC"

	data, err := ui.db.Query(query, args...)
	if err != nil {
		return rows
	}
	defer data.Close()

	for data.Next() {
		var id, title string
		var first, last sql.NullTime
		if err := data.Scan(&id, &title, &first, &last); err != nil {
			continue
		}
		firstStr := ""
		if first.Valid {
			firstStr = first.Time.Format("01-02 15:04")
		}
		lastStr := ""
		if last.Valid {
			lastStr = last.Time.Format("01-02 15:04")
		}
		rows = append(rows, []string{
			id,
			title,
			firstStr,
			lastStr,
		})
	}
	return rows
}

func (ui *FyneUI) openHistoricalRoomTab(roomID string) {
	if roomID == "" {
		return
	}
	historyKey := fmt.Sprintf("%s#history", roomID)
	if _, exists := ui.roomTabs[historyKey]; exists {
		ui.tabContainer.Select(ui.roomTabs[historyKey].Tab)
		return
	}

	ui.AddOrUpdateRoom(historyKey)
	roomTab := ui.roomTabs[historyKey]
	roomTab.RoomID = roomID
	roomTab.RoomName = fmt.Sprintf("%s (历史)", roomID)
	roomTab.Tab.Text = fmt.Sprintf("房间 %s(历史)", roomID)

	historyPairs := ui.loadHistoricalMessages(roomID)
	roomTab.MessagePairs = historyPairs
	roomTab.TotalMessages = ui.fetchRoomMessageCount(roomID)
	if roomTab.TotalMessages == 0 {
		roomTab.TotalMessages = len(roomTab.MessagePairs)
	}
	ui.applyRoomFilter(roomTab)
	ui.refreshRoomTables(roomTab)
	if roomTab.MessagesList != nil {
		roomTab.MessagesList.Refresh()
	}
	ui.updateRoomStats(roomTab)
}

func (ui *FyneUI) loadHistoricalMessages(roomID string) []*MessagePair {
	if ui.db == nil {
		return nil
	}
	tableName := database.RoomMessageTableName(roomID)
	query := fmt.Sprintf(`SELECT COALESCE(create_time, timestamp), display, message_type, method, raw_payload, parsed_json FROM %s ORDER BY COALESCE(create_time, timestamp) DESC LIMIT 200`, tableName)
	rows, err := ui.db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make([]*MessagePair, 0)
	for rows.Next() {
		var ts time.Time
		var display, msgType, method, parsedJSON string
		var rawPayload []byte
		if err := rows.Scan(&ts, &display, &msgType, &method, &rawPayload, &parsedJSON); err != nil {
			continue
		}
		parsed := &parser.ParsedProtoMessage{
			Method:      method,
			Display:     display,
			MessageType: msgType,
			RawPayload:  rawPayload,
			RawJSON:     parsedJSON,
			ReceivedAt:  ts,
			Detail: map[string]interface{}{
				"messageType": msgType,
				"method":      method,
			},
		}
		result = append(result, &MessagePair{
			Parsed: parsed,
			Display: ui.decorateMessageDisplay(&MessagePair{
				Parsed:    parsed,
				Display:   display,
				Detail:    parsed.Detail,
				Timestamp: ts,
			}),
			Detail:    parsed.Detail,
			Timestamp: ts,
			Source:    "history",
		})
	}
	return result
}

func normalizeRoomID(roomID string) string {
	if idx := strings.Index(roomID, "#"); idx >= 0 {
		return roomID[:idx]
	}
	return roomID
}

func (ui *FyneUI) fetchRoomMessageCount(roomID string) int {
	if ui.db == nil || roomID == "" {
		return 0
	}
	tableName := database.RoomMessageTableName(normalizeRoomID(roomID))
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)
	var total int
	if err := ui.db.QueryRow(query).Scan(&total); err != nil {
		return 0
	}
	return total
}

func (ui *FyneUI) exportRoomGifts(roomID string) (string, error) {
	if ui.db == nil || roomID == "" {
		return "", fmt.Errorf("缺少房间号")
	}
	if err := os.MkdirAll("exports", 0755); err != nil {
		return "", err
	}

	path := filepath.Join("exports", fmt.Sprintf("room_%s_gifts.xlsx", roomID))
	file := excelize.NewFile()
	defer file.Close()

	const sheet = "礼物记录"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{"时间", "礼物名称", "礼物数量", "送礼人", "钻石", "接收主播"}
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		file.SetCellValue(sheet, cell, header)
	}

	rows, err := ui.db.Query(`
		SELECT COALESCE(gr.create_time, gr.timestamp), gr.gift_name, gr.gift_count, gr.user_nickname,
		       gr.gift_diamond_value, COALESCE(gr.anchor_name, a.anchor_name, gr.anchor_id) AS receiver
		FROM gift_records gr
		LEFT JOIN anchors a ON gr.anchor_id = a.anchor_id
		WHERE gr.room_id = ?
		ORDER BY COALESCE(gr.create_time, gr.timestamp) ASC
	`, roomID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	rowIdx := 2
	for rows.Next() {
		var ts time.Time
		var giftName, user, receiver sql.NullString
		var count, diamond int
		if err := rows.Scan(&ts, &giftName, &count, &user, &diamond, &receiver); err != nil {
			continue
		}
		totalDiamond := diamond * count
		if totalDiamond == 0 {
			totalDiamond = diamond
		}
		values := []interface{}{
			ts.Format("2006-01-02 15:04:05"),
			giftName.String,
			count,
			user.String,
			totalDiamond,
			strings.TrimSpace(receiver.String),
		}
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
			file.SetCellValue(sheet, cell, value)
		}
		rowIdx++
	}

	if err := file.SetColWidth(sheet, "A", "A", 20); err != nil {
		return "", err
	}
	if err := file.SetColWidth(sheet, "B", "F", 18); err != nil {
		return "", err
	}
	if err := file.SaveAs(path); err != nil {
		return "", err
	}

	return path, nil
}

func (ui *FyneUI) exportRoomAnchorScores(roomID string) (string, error) {
	if ui.db == nil || roomID == "" {
		return "", fmt.Errorf("缺少房间号")
	}
	path := filepath.Join("exports", fmt.Sprintf("room_%s_anchors.csv", roomID))
	if err := os.MkdirAll("exports", 0755); err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"主播ID", "主播名称", "礼物计数", "得分"})
	rows, err := ui.db.Query(`
		SELECT anchor_id, anchor_name, gift_count, score
		FROM room_anchors WHERE room_id = ? ORDER BY score DESC
	`, roomID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var anchorID, anchorName string
		var count, score int
		if err := rows.Scan(&anchorID, &anchorName, &count, &score); err != nil {
			continue
		}
		writer.Write([]string{
			anchorID,
			anchorName,
			fmt.Sprintf("%d", count),
			fmt.Sprintf("%d", score),
		})
	}
	return path, nil
}

func (ui *FyneUI) loadThemePreference() string {
	if ui.preferencesPath == "" {
		return "系统默认"
	}
	data, err := os.ReadFile(ui.preferencesPath)
	if err != nil {
		return "系统默认"
	}
	var pref struct {
		Theme string `json:"theme"`
	}
	if err := json.Unmarshal(data, &pref); err != nil || pref.Theme == "" {
		return "系统默认"
	}
	return pref.Theme
}

func (ui *FyneUI) saveThemePreference(themeName string) {
	if ui.preferencesPath == "" {
		return
	}
	pref := struct {
		Theme string `json:"theme"`
	}{Theme: themeName}
	data, _ := json.MarshalIndent(pref, "", "  ")
	_ = os.WriteFile(ui.preferencesPath, data, 0644)
}

func (ui *FyneUI) applyTheme(themeName string) {
	switch themeName {
	case "浅色":
		ui.app.Settings().SetTheme(theme.LightTheme())
	case "深色":
		ui.app.Settings().SetTheme(theme.DarkTheme())
	default:
		ui.app.Settings().SetTheme(NewChineseTheme())
	}
	ui.userTheme = themeName

	// 刷新所有界面组件以响应主题变化
	ui.refreshAllUIComponents()
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

	debugLabel := widget.NewLabelWithData(ui.debugMode)
	debugLabel.TextStyle = fyne.TextStyle{Bold: true}
	debugSection := container.NewVBox(
		widget.NewLabel("调试状态"),
		debugLabel,
	)

	isInitializing := true
	themeSelect := widget.NewSelect([]string{"系统默认", "浅色", "深色"}, func(val string) {
		ui.applyTheme(val)
		ui.saveThemePreference(val)
		// 只在用户手动更改时提示，初始化时不提示
		if !isInitializing && ui.mainWin != nil {
			dialog.ShowInformation("主题已更新", "主题设置已保存并应用", ui.mainWin)
		}
	})
	themeSelect.SetSelected(ui.userTheme)
	isInitializing = false
	themeSection := container.NewVBox(
		widget.NewLabel("主题设置"),
		themeSelect,
		widget.NewLabel("提示：切换主题后界面将自动刷新"),
	)

	return container.NewVBox(
		portForm,
		widget.NewSeparator(),
		pluginSection,
		widget.NewSeparator(),
		licenseSection,
		widget.NewSeparator(),
		themeSection,
		widget.NewSeparator(),
		debugSection,
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

// closeRoom 关闭房间并清理资源
func (ui *FyneUI) closeRoom(roomID string) {
	ui.stopManualRoom(roomID)

	roomTab, exists := ui.roomTabs[roomID]
	if !exists {
		ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 已关闭", roomID))
		return
	}

	if ui.tabContainer != nil {
		ui.tabContainer.Remove(roomTab.Tab)
	}
	delete(ui.roomTabs, roomID)
	ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 已关闭", roomID))
}

func (ui *FyneUI) updateRoomStats(roomTab *RoomTab) {
	if roomTab == nil || roomTab.StatsLabel == nil {
		return
	}
	displayed := len(roomTab.MessagePairs)
	if displayed > roomTab.TotalMessages {
		roomTab.TotalMessages = displayed
	}
	total := roomTab.TotalMessages
	if total == 0 {
		total = displayed
	}
	extra := ""
	if total > displayed {
		extra = fmt.Sprintf(" (展示 %d 条)", displayed)
	}
	roomTab.StatsLabel.SetText(fmt.Sprintf("房间: %s | 消息: %d 条%s", roomTab.RoomID, total, extra))
}

// AddOrUpdateRoom 添加或更新房间Tab
func (ui *FyneUI) AddOrUpdateRoom(roomID string) {
	if _, exists := ui.roomTabs[roomID]; exists {
		return
	}

	roomTab := &RoomTab{
		RoomID:        roomID,
		RoomName:      roomID,
		MessagePairs:  make([]*MessagePair, 0, 200),
		FilteredPairs: make([]*MessagePair, 0, 200),
	}

	if ui.currentRoomLabel != nil {
		ui.currentRoomLabel.SetText(fmt.Sprintf("当前监控房间: %s", roomID))
	}

	roomTab.StatsLabel = widget.NewLabel(fmt.Sprintf("房间: %s | 消息: 0 条", roomID))
	roomTab.TotalMessages = ui.fetchRoomMessageCount(roomID)
	ui.updateRoomStats(roomTab)

	roomTab.FilterSelect = widget.NewSelect([]string{"全部", "聊天消息", "礼物消息", "点赞消息", "进场消息", "关注消息"}, func(val string) {
		roomTab.MessageFilter = val
		ui.applyRoomFilter(roomTab)
		if roomTab.MessagesList != nil {
			roomTab.MessagesList.Refresh()
			roomTab.MessagesList.ScrollToTop()
		}
	})

	roomTab.MessagesList = widget.NewList(
		func() int {
			return len(roomTab.FilteredPairs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("消息")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(roomTab.FilteredPairs) {
				item.(*widget.Label).SetText(roomTab.FilteredPairs[id].Display)
			}
		},
	)

	roomTab.MessagesList.OnSelected = func(id widget.ListItemID) {
		ui.showMessageDetail(roomTab, id)
	}

	roomTab.FilterSelect.SetSelected("全部")

	giftOnlyBtn := widget.NewButton("礼物记录视图", func() {
		ui.showGiftRecordWindow(roomID)
	})

	messagesHeader := container.NewHBox(
		widget.NewLabel("筛选:"),
		roomTab.FilterSelect,
		giftOnlyBtn,
		layout.NewSpacer(),
	)

	messagesTab := container.NewBorder(
		container.NewVBox(messagesHeader, widget.NewSeparator()),
		nil, nil, nil,
		container.NewScroll(roomTab.MessagesList),
	)

	ui.initRoomGiftTable(roomTab)
	anchorContent := ui.initRoomAnchorTable(roomTab)
	ui.initRoomSegmentTable(roomTab)

	roomTab.SubTabs = container.NewAppTabs(
		container.NewTabItem("礼物记录", container.NewScroll(roomTab.GiftTable)),
		container.NewTabItem("主播管理", anchorContent),
		container.NewTabItem("消息记录", messagesTab),
		container.NewTabItem("分段记分", container.NewScroll(roomTab.SegmentTable)),
	)

	closeBtn := widget.NewButtonWithIcon("关闭", theme.CancelIcon(), func() {
		ui.closeRoom(roomID)
	})

	header := container.NewHBox(
		roomTab.StatsLabel,
		layout.NewSpacer(),
		closeBtn,
	)

	content := container.NewBorder(
		header,
		nil, nil, nil,
		roomTab.SubTabs,
	)

	roomTab.Tab = container.NewTabItem(fmt.Sprintf("房间 %s", roomID), content)

	ui.roomTabs[roomID] = roomTab
	ui.tabContainer.Append(roomTab.Tab)
	ui.tabContainer.Select(roomTab.Tab)

	log.Printf("✅ 房间 Tab 创建成功: %s", roomID)
}

// AddParsedMessage 添加解析后的消息（纯文本）
func (ui *FyneUI) AddParsedMessage(roomID string, message string) {
	parsed := &parser.ParsedProtoMessage{
		Method:      "System",
		Display:     message,
		Detail:      map[string]interface{}{"messageType": "系统", "content": message},
		RawJSON:     message,
		RawPayload:  []byte(message),
		ReceivedAt:  time.Now(),
		MessageType: "系统",
	}
	ui.recordParsedMessage(roomID, parsed, false)
}

// AddParsedMessageWithDetail 添加解析后的消息（包含详细信息）
func (ui *FyneUI) AddParsedMessageWithDetail(roomID string, message string, detail map[string]interface{}) {
	if detail != nil {
		if parsed, ok := detail["_parsed"].(*parser.ParsedProtoMessage); ok {
			ui.recordParsedMessage(roomID, parsed, false)
			// 如果是礼物消息，刷新礼物表格（因为 WebSocket 已经保存到数据库了）
			if parsed.MessageType == "礼物消息" {
				if roomTab, ok := ui.roomTabs[roomID]; ok {
					log.Printf("🔄 [房间 %s] 浏览器插件礼物消息，刷新礼物表格", roomID)
					ui.refreshRoomTables(roomTab)
				}
			}
			return
		}
	}

	if detail == nil {
		detail = make(map[string]interface{})
	}

	method := fmt.Sprintf("%v", detail["method"])
	msgType := fmt.Sprintf("%v", detail["messageType"])

	rawJSON, _ := json.Marshal(detail)
	parsed := &parser.ParsedProtoMessage{
		Method:      method,
		Display:     message,
		Detail:      detail,
		RawJSON:     string(rawJSON),
		RawPayload:  []byte(message),
		ReceivedAt:  time.Now(),
		MessageType: msgType,
	}

	ui.recordParsedMessage(roomID, parsed, false)

	// 如果是礼物消息，刷新礼物表格
	if msgType == "礼物消息" {
		if roomTab, ok := ui.roomTabs[roomID]; ok {
			log.Printf("🔄 [房间 %s] 浏览器插件礼物消息，刷新礼物表格", roomID)
			ui.refreshRoomTables(roomTab)
		}
	}
}

func formatDisplayWithTimestamp(ts time.Time, original string) string {
	if ts.IsZero() {
		return original
	}

	clean := original
	if strings.HasPrefix(clean, "[") {
		if idx := strings.Index(clean, "]"); idx > 0 && idx+2 <= len(clean) {
			candidate := clean[1:idx]
			if len(candidate) == len("15:04:05") {
				if _, err := time.Parse("15:04:05", candidate); err == nil {
					clean = strings.TrimSpace(clean[idx+1:])
				}
			} else if len(candidate) == len("01-02 15:04:05") {
				if _, err := time.Parse("01-02 15:04:05", candidate); err == nil {
					clean = strings.TrimSpace(clean[idx+1:])
				}
			}
		}
	}

	return fmt.Sprintf("[%s] %s", ts.Format("01-02 15:04:05"), clean)
}

func (ui *FyneUI) decorateMessageDisplay(pair *MessagePair) string {
	if pair == nil {
		return ""
	}
	if pair.Detail == nil {
		pair.Detail = make(map[string]interface{})
	}

	display := formatDisplayWithTimestamp(pair.Timestamp, pair.Display)

	if mt, ok := pair.Detail["messageType"].(string); ok && mt == "礼物消息" {
		group := toInt(pair.Detail["groupCount"])
		if group == 0 {
			group = toInt(pair.Detail["giftCount"])
		}
		if group == 0 {
			group = 1
		}
		diamond := toInt(pair.Detail["diamondCount"])
		total := diamond * group
		if total == 0 {
			total = toInt(pair.Detail["diamondTotal"])
		}
		if total > 0 {
			pair.Detail["diamondTotal"] = total
			if !strings.Contains(display, "💎") {
				display = fmt.Sprintf("%s | 💎%d", display, total)
			}
		}
	}

	return display
}

func (ui *FyneUI) recordParsedMessage(roomID string, parsed *parser.ParsedProtoMessage, persist bool) {
	if parsed == nil {
		return
	}

	if parsed.ReceivedAt.IsZero() {
		parsed.ReceivedAt = time.Now()
	}
	if parsed.Detail == nil {
		parsed.Detail = make(map[string]interface{})
	}
	parsed.Detail["timestamp"] = parsed.ReceivedAt.Format(time.RFC3339)
	tempPair := &MessagePair{
		Parsed:    parsed,
		Display:   parsed.Display,
		Detail:    parsed.Detail,
		Timestamp: parsed.ReceivedAt,
	}
	displayText := ui.decorateMessageDisplay(tempPair)

	ui.AddOrUpdateRoom(roomID)
	roomTab := ui.roomTabs[roomID]
	if roomTab.MessageFilter == "" {
		roomTab.MessageFilter = "全部"
	}

	source := fmt.Sprintf("%v", parsed.Detail["source"])
	if source == "<nil>" || source == "" {
		source = "browser"
	}
	pair := &MessagePair{
		Parsed:    parsed,
		Display:   displayText,
		Detail:    parsed.Detail,
		Timestamp: parsed.ReceivedAt,
		Source:    source,
	}

	if parsed.MessageType == "礼物消息" {
		ui.handleGiftAssignment(roomID, pair.Detail)

		// 保存礼物记录到 gift_records 表
		if persist && ui.db != nil {
			log.Printf("🎁 [房间 %s] 手动连接收到礼物消息，准备保存到 gift_records", roomID)
			if err := ui.saveManualGiftRecord(roomID, parsed); err != nil {
				log.Printf("❌ [房间 %s] 保存手动房间礼物记录失败: %v", roomID, err)
			} else {
				// 保存成功后刷新礼物表格
				if roomTab, ok := ui.roomTabs[roomID]; ok {
					log.Printf("🔄 [房间 %s] 刷新礼物表格", roomID)
					ui.refreshRoomTables(roomTab)
				}
			}
		}
	}

	roomTab.MessagePairs = append([]*MessagePair{pair}, roomTab.MessagePairs...)

	ui.applyRoomFilter(roomTab)
	if roomTab.MessagesList != nil {
		roomTab.MessagesList.Refresh()
		roomTab.MessagesList.ScrollToTop()
	}
	roomTab.TotalMessages++
	ui.updateRoomStats(roomTab)

	if persist && ui.wsServer != nil {
		source := pair.Source
		if source == "" {
			source = "manual"
		}
		if err := ui.wsServer.PersistRoomMessage(roomID, parsed, source); err != nil {
			log.Printf("⚠️  保存房间 %s 消息失败: %v", roomID, err)
		}
	}
}

func (ui *FyneUI) applyRoomFilter(roomTab *RoomTab) {
	filter := roomTab.MessageFilter
	if filter == "" {
		filter = "全部"
		roomTab.MessageFilter = filter
	}
	if filter == "全部" {
		roomTab.FilteredPairs = append([]*MessagePair(nil), roomTab.MessagePairs...)
		sort.SliceStable(roomTab.FilteredPairs, func(i, j int) bool {
			return roomTab.FilteredPairs[i].Timestamp.After(roomTab.FilteredPairs[j].Timestamp)
		})
		return
	}

	filtered := make([]*MessagePair, 0, len(roomTab.MessagePairs))
	for _, pair := range roomTab.MessagePairs {
		if mt, ok := pair.Detail["messageType"].(string); ok && mt == filter {
			filtered = append(filtered, pair)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})
	roomTab.FilteredPairs = filtered
}

func (ui *FyneUI) handleGiftAssignment(roomID string, detail map[string]interface{}) {
	if ui.db == nil {
		return
	}

	giftName := fmt.Sprintf("%v", detail["giftName"])
	if giftName == "" {
		return
	}

	anchorID := fmt.Sprintf("%v", detail["anchorId"])
	anchorName := fmt.Sprintf("%v", detail["anchorName"])

	if anchorID == "" {
		anchorID, anchorName = ui.lookupGiftBinding(roomID, giftName)
		if anchorID == "" {
			return
		}
	}

	ui.ensureRoomAnchorRecord(roomID, anchorID, anchorName)
	ui.ensureGlobalAnchor(anchorID, anchorName)
	ui.incrementAnchorScore(roomID, anchorID, toInt(detail["groupCount"]), toInt(detail["diamondCount"]))

	if roomTab, ok := ui.roomTabs[roomID]; ok {
		ui.refreshRoomTables(roomTab)
	}
}

func (ui *FyneUI) ensureRoomAnchorRecord(roomID, anchorID, anchorName string) {
	if ui.db == nil || anchorID == "" {
		return
	}

	_, err := ui.db.Exec(`
		INSERT INTO room_anchors (room_id, anchor_id, anchor_name, gift_count, score)
		VALUES (?, ?, ?, 0, 0)
		ON CONFLICT(room_id, anchor_id) DO UPDATE SET anchor_name=excluded.anchor_name
	`, roomID, anchorID, anchorName)
	if err != nil {
		log.Printf("⚠️  更新房间主播失败: %v", err)
	}
}

func (ui *FyneUI) incrementAnchorScore(roomID, anchorID string, giftCount, diamond int) {
	if ui.db == nil || anchorID == "" {
		return
	}

	_, err := ui.db.Exec(`
		UPDATE room_anchors
		SET gift_count = gift_count + ?, score = score + ?
		WHERE room_id = ? AND anchor_id = ?
	`, giftCount, giftCount*diamond, roomID, anchorID)

	if err != nil {
		log.Printf("⚠️  更新主播得分失败: %v", err)
	}
}

func (ui *FyneUI) ensureGlobalAnchor(anchorID, anchorName string) {
	if ui.db == nil || anchorID == "" {
		return
	}
	anchorName = strings.TrimSpace(anchorName)
	if anchorName == "" {
		anchorName = anchorID
	}

	_, err := ui.db.Exec(`
		INSERT INTO anchors (anchor_id, anchor_name, bound_gifts, created_at, updated_at)
		VALUES (?, ?, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(anchor_id)
		DO UPDATE SET anchor_name=CASE WHEN excluded.anchor_name = '' THEN anchors.anchor_name ELSE excluded.anchor_name END,
		          updated_at=CURRENT_TIMESTAMP
	`, anchorID, anchorName)
	if err != nil {
		log.Printf("⚠️  同步全局主播失败: %v", err)
	}
}

func (ui *FyneUI) initializeRoomAnchors(roomTab *RoomTab) {
	if roomTab == nil || ui.db == nil {
		return
	}

	updateStatus := func(text string) {
		if roomTab.AnchorStatus != nil {
			roomTab.AnchorStatus.SetText(text)
		}
	}

	defaultID := fmt.Sprintf("%s_anchor", roomTab.RoomID)
	defaultName := roomTab.RoomName
	if strings.TrimSpace(defaultName) == "" {
		defaultName = defaultID
	}

	_, err := ui.db.Exec(`
		INSERT INTO room_anchors (room_id, anchor_id, anchor_name, bound_gifts, gift_count, score)
		VALUES (?, ?, ?, '', 0, 0)
		ON CONFLICT(room_id, anchor_id) DO NOTHING
	`, roomTab.RoomID, defaultID, defaultName)
	if err != nil {
		updateStatus(fmt.Sprintf("⚠️ 初始化失败: %v", err))
		return
	}

	ui.ensureGlobalAnchor(defaultID, defaultName)
	ui.refreshRoomTables(roomTab)
	updateStatus("✅ 已添加默认主播，可继续编辑")
}

func (ui *FyneUI) bindGiftsToAnchor(roomID, anchorID, gifts string) {
	if ui.db == nil || roomID == "" || anchorID == "" || strings.TrimSpace(gifts) == "" {
		return
	}

	giftList := strings.Split(gifts, ",")
	for _, name := range giftList {
		giftName := strings.TrimSpace(name)
		if giftName == "" {
			continue
		}
		if _, err := ui.db.Exec(`
			INSERT INTO room_gift_bindings (room_id, gift_name, anchor_id)
			VALUES (?, ?, ?)
			ON CONFLICT(room_id, gift_name) DO UPDATE SET anchor_id=excluded.anchor_id
		`, roomID, giftName, anchorID); err != nil {
			log.Printf("⚠️  绑定礼物 %s 到主播 %s 失败: %v", giftName, anchorID, err)
			continue
		}
		anchorName := ui.lookupAnchorName(anchorID)
		ui.ensureGlobalAnchor(anchorID, anchorName)
		ui.ensureRoomAnchorRecord(roomID, anchorID, anchorName)
	}
}

func (ui *FyneUI) lookupGiftBinding(roomID, giftName string) (string, string) {
	if ui.db == nil {
		return "", ""
	}
	var anchorID string
	err := ui.db.QueryRow(`
		SELECT anchor_id FROM room_gift_bindings
		WHERE room_id = ? AND gift_name = ?
	`, roomID, giftName).Scan(&anchorID)
	if err != nil {
		return "", ""
	}

	var anchorName string
	_ = ui.db.QueryRow(`SELECT anchor_name FROM anchors WHERE anchor_id = ?`, anchorID).Scan(&anchorName)
	return anchorID, anchorName
}

func (ui *FyneUI) lookupAnchorName(anchorID string) string {
	if ui.db == nil || anchorID == "" {
		return anchorID
	}
	var anchorName string
	if err := ui.db.QueryRow(`SELECT anchor_name FROM anchors WHERE anchor_id = ?`, anchorID).Scan(&anchorName); err != nil {
		return anchorID
	}
	anchorName = strings.TrimSpace(anchorName)
	if anchorName == "" {
		return anchorID
	}
	return anchorName
}

func (ui *FyneUI) refreshRoomTables(roomTab *RoomTab) {
	log.Printf("🔄 [房间 %s] refreshRoomTables 开始刷新表格", roomTab.RoomID)

	roomTab.GiftRows = ui.loadRoomGiftRows(roomTab.RoomID)
	log.Printf("📊 [房间 %s] GiftRows 更新完成，当前行数: %d", roomTab.RoomID, len(roomTab.GiftRows))

	roomTab.AnchorRows = ui.loadRoomAnchorRows(roomTab.RoomID)
	roomTab.SegmentRows = ui.loadRoomSegmentRows(roomTab.RoomID)

	if roomTab.GiftTable != nil {
		log.Printf("🔄 [房间 %s] 刷新 GiftTable UI", roomTab.RoomID)
		roomTab.GiftTable.Refresh()
	} else {
		log.Printf("⚠️  [房间 %s] GiftTable 为 nil，无法刷新", roomTab.RoomID)
	}

	if roomTab.AnchorTable != nil {
		roomTab.AnchorTable.Refresh()
	}
	if roomTab.SegmentTable != nil {
		roomTab.SegmentTable.Refresh()
	}
	ui.refreshRoomAnchorPicker(roomTab)

	log.Printf("✅ [房间 %s] refreshRoomTables 完成", roomTab.RoomID)
}

func (ui *FyneUI) initRoomGiftTable(roomTab *RoomTab) {
	log.Printf("🏗️  [房间 %s] 初始化礼物表格", roomTab.RoomID)

	roomTab.GiftRows = ui.loadRoomGiftRows(roomTab.RoomID)
	log.Printf("📊 [房间 %s] 初始化时加载了 %d 行数据", roomTab.RoomID, len(roomTab.GiftRows))

	table := widget.NewTable(
		func() (int, int) {
			if len(roomTab.GiftRows) == 0 {
				log.Printf("⚠️  [房间 %s] GiftRows 为空，返回 0 行", roomTab.RoomID)
				return 0, 0
			}
			rows := len(roomTab.GiftRows)
			cols := len(roomTab.GiftRows[0])
			log.Printf("📐 [房间 %s] 表格尺寸: %d 行 x %d 列", roomTab.RoomID, rows, cols)
			return rows, cols
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row < len(roomTab.GiftRows) && id.Col < len(roomTab.GiftRows[id.Row]) {
				cell.(*widget.Label).SetText(roomTab.GiftRows[id.Row][id.Col])
			}
		},
	)
	table.SetColumnWidth(0, 140)
	table.SetColumnWidth(1, 140)
	table.SetColumnWidth(2, 80)
	table.SetColumnWidth(3, 80)
	table.SetColumnWidth(4, 120)
	table.SetColumnWidth(5, 140)
	roomTab.GiftTable = table

	log.Printf("✅ [房间 %s] 礼物表格初始化完成", roomTab.RoomID)
}

func (ui *FyneUI) initRoomAnchorTable(roomTab *RoomTab) fyne.CanvasObject {
	roomTab.AnchorRows = ui.loadRoomAnchorRows(roomTab.RoomID)

	table := widget.NewTable(
		func() (int, int) {
			if len(roomTab.AnchorRows) == 0 {
				return 0, 0
			}
			return len(roomTab.AnchorRows), len(roomTab.AnchorRows[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row < len(roomTab.AnchorRows) && id.Col < len(roomTab.AnchorRows[id.Row]) {
				cell.(*widget.Label).SetText(roomTab.AnchorRows[id.Row][id.Col])
			}
		},
	)
	table.SetColumnWidth(0, 120)
	table.SetColumnWidth(1, 140)
	table.SetColumnWidth(2, 200)
	table.SetColumnWidth(3, 100)
	table.SetColumnWidth(4, 100)
	roomTab.AnchorTable = table

	roomTab.AnchorOptionMap = make(map[string]AnchorOption)
	anchorPicker := widget.NewSelect([]string{}, func(val string) {
		if roomTab.AnchorOptionMap == nil {
			return
		}
		if opt, ok := roomTab.AnchorOptionMap[val]; ok {
			roomTab.AnchorIDEntry.SetText(opt.ID)
			roomTab.AnchorNameEntry.SetText(opt.Name)
		}
	})
	anchorPicker.PlaceHolder = "选择全局主播"
	roomTab.AnchorPicker = anchorPicker

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("主播ID")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("主播名称")

	// 创建礼物选择器
	giftFilterEntry := widget.NewEntry()
	giftFilterEntry.SetPlaceHolder("筛选礼物...")

	selectedGifts := make(map[string]bool)
	giftListData := make([]string, 0)

	giftList := widget.NewList(
		func() int { return len(giftListData) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			return container.NewHBox(check, widget.NewLabel(""))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(giftListData) {
				giftName := giftListData[id]
				hbox := item.(*fyne.Container)
				check := hbox.Objects[0].(*widget.Check)
				label := hbox.Objects[1].(*widget.Label)

				label.SetText(giftName)
				check.SetChecked(selectedGifts[giftName])
				check.OnChanged = func(checked bool) {
					if checked {
						selectedGifts[giftName] = true
					} else {
						delete(selectedGifts, giftName)
					}
				}
			}
		},
	)
	giftList.Resize(fyne.NewSize(300, 200))

	// 加载礼物列表
	loadGiftList := func(filter string) {
		allGifts := ui.loadAllGiftNames()
		giftListData = make([]string, 0)

		filterLower := strings.ToLower(strings.TrimSpace(filter))
		for _, gift := range allGifts {
			if filterLower == "" || strings.Contains(strings.ToLower(gift), filterLower) {
				giftListData = append(giftListData, gift)
			}
		}
		giftList.Refresh()
	}

	giftFilterEntry.OnChanged = func(text string) {
		loadGiftList(text)
	}

	// 初始加载礼物列表
	loadGiftList("")

	giftsDisplay := widget.NewLabel("")
	giftsDisplay.Wrapping = fyne.TextWrapWord

	updateGiftsDisplay := func() {
		gifts := make([]string, 0, len(selectedGifts))
		for gift := range selectedGifts {
			gifts = append(gifts, gift)
		}
		sort.Strings(gifts)
		giftsDisplay.SetText(strings.Join(gifts, ", "))
	}

	giftCountEntry := widget.NewEntry()
	giftCountEntry.SetPlaceHolder("礼物数量")
	scoreEntry := widget.NewEntry()
	scoreEntry.SetPlaceHolder("钻石总值")
	statusLabel := widget.NewLabel("")

	roomTab.AnchorIDEntry = idEntry
	roomTab.AnchorNameEntry = nameEntry
	roomTab.AnchorGiftCountEntry = giftCountEntry
	roomTab.AnchorScoreEntry = scoreEntry
	roomTab.AnchorStatus = statusLabel

	updateInitBtnState := func(btn *widget.Button) {
		if btn == nil {
			return
		}
		if len(roomTab.AnchorRows) <= 1 {
			btn.Enable()
		} else {
			btn.Disable()
		}
	}

	var initBtn *widget.Button
	initBtn = widget.NewButton("初始化主播", func() {
		ui.initializeRoomAnchors(roomTab)
		ui.refreshRoomTables(roomTab)
		updateInitBtnState(initBtn)
	})
	updateInitBtnState(initBtn)

	saveBtn := widget.NewButton("保存/更新", func() {
		updateGiftsDisplay()
		gifts := make([]string, 0, len(selectedGifts))
		for gift := range selectedGifts {
			gifts = append(gifts, gift)
		}

		// 将选中的礼物保存到主播
		anchorID := strings.TrimSpace(idEntry.Text)
		anchorName := strings.TrimSpace(nameEntry.Text)

		if anchorID == "" || anchorName == "" {
			statusLabel.SetText("⚠️ 请填写主播ID和名称")
			return
		}

		giftCount, _ := strconv.Atoi(strings.TrimSpace(giftCountEntry.Text))
		score, _ := strconv.Atoi(strings.TrimSpace(scoreEntry.Text))
		giftStr := strings.Join(gifts, ",")

		tx, err := ui.db.Begin()
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("⚠️ 数据库错误: %v", err))
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(`
			INSERT INTO room_anchors (room_id, anchor_id, anchor_name, bound_gifts, gift_count, score)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(room_id, anchor_id)
			DO UPDATE SET anchor_name=excluded.anchor_name,
			             bound_gifts=excluded.bound_gifts,
			             gift_count=excluded.gift_count,
			             score=excluded.score
		`, roomTab.RoomID, anchorID, anchorName, giftStr, giftCount, score)

		if err != nil {
			statusLabel.SetText(fmt.Sprintf("⚠️ 保存失败: %v", err))
			return
		}

		if err := tx.Commit(); err != nil {
			statusLabel.SetText(fmt.Sprintf("⚠️ 保存失败: %v", err))
			return
		}

		ui.ensureGlobalAnchor(anchorID, anchorName)
		ui.bindGiftsToAnchor(roomTab.RoomID, anchorID, giftStr)
		ui.refreshRoomTables(roomTab)
		updateInitBtnState(initBtn)
		statusLabel.SetText("✅ 主播信息已保存")
	})

	refreshBtn := widget.NewButton("刷新", func() {
		ui.refreshRoomTables(roomTab)
		updateInitBtnState(initBtn)
	})

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row >= len(roomTab.AnchorRows) {
			return
		}
		row := roomTab.AnchorRows[id.Row]
		if len(row) >= 5 {
			idEntry.SetText(row[0])
			nameEntry.SetText(row[1])

			// 解析绑定的礼物并更新选择状态
			selectedGifts = make(map[string]bool)
			if row[2] != "" {
				gifts := strings.Split(row[2], ",")
				for _, gift := range gifts {
					gift = strings.TrimSpace(gift)
					if gift != "" {
						selectedGifts[gift] = true
					}
				}
			}
			updateGiftsDisplay()
			giftList.Refresh()

			giftCountEntry.SetText(row[3])
			scoreEntry.SetText(row[4])
		}
	}

	giftSection := container.NewVBox(
		widget.NewLabel("绑定礼物"),
		giftFilterEntry,
		container.NewScroll(giftList),
		widget.NewLabel("已选择:"),
		container.NewScroll(giftsDisplay),
	)

	form := container.NewVBox(
		widget.NewLabel("选择全局主播"),
		container.NewHBox(anchorPicker, widget.NewButton("刷新", func() {
			ui.refreshRoomAnchorPicker(roomTab)
		})),
		widget.NewSeparator(),
		widget.NewLabel("主播信息"),
		widget.NewLabel("主播ID"),
		idEntry,
		widget.NewLabel("主播名称"),
		nameEntry,
		giftSection,
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("礼物数量"), giftCountEntry),
			container.NewVBox(widget.NewLabel("钻石总值"), scoreEntry),
		),
		container.NewHBox(saveBtn, refreshBtn, initBtn),
		statusLabel,
	)

	content := container.NewHSplit(
		container.NewScroll(table),
		container.NewPadded(form),
	)
	content.SetOffset(0.55)

	ui.refreshRoomAnchorPicker(roomTab)

	return content
}

func (ui *FyneUI) initRoomSegmentTable(roomTab *RoomTab) {
	roomTab.SegmentRows = ui.loadRoomSegmentRows(roomTab.RoomID)
	table := widget.NewTable(
		func() (int, int) {
			if len(roomTab.SegmentRows) == 0 {
				return 0, 0
			}
			return len(roomTab.SegmentRows), len(roomTab.SegmentRows[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row < len(roomTab.SegmentRows) && id.Col < len(roomTab.SegmentRows[id.Row]) {
				cell.(*widget.Label).SetText(roomTab.SegmentRows[id.Row][id.Col])
			}
		},
	)
	table.SetColumnWidth(0, 160)
	table.SetColumnWidth(1, 140)
	table.SetColumnWidth(2, 140)
	table.SetColumnWidth(3, 120)
	roomTab.SegmentTable = table
}

// showMessageDetail 显示消息详情对话框
func (ui *FyneUI) showMessageDetail(roomTab *RoomTab, id widget.ListItemID) {
	if id >= len(roomTab.FilteredPairs) {
		return
	}

	pair := roomTab.FilteredPairs[id]

	// 构建详情内容
	detailText := fmt.Sprintf("📅 时间: %s\n来源: %s\n\n", pair.Timestamp.Format("2006-01-02 15:04:05"), pair.Source)
	detailText += "📋 展示:\n" + pair.Display + "\n\n"

	if pair.Detail != nil {
		detailText += "🔍 详细信息:\n"
		for key, value := range pair.Detail {
			detailText += fmt.Sprintf("  %s: %v\n", key, value)
		}
		detailText += "\n"
	}

	if pair.Parsed != nil {
		detailText += "🧾 JSON:\n" + pair.Parsed.RawJSON + "\n\n"
		if len(pair.Parsed.RawPayload) > 0 {
			detailText += "📦 原始Payload(Base64):\n" + base64.StdEncoding.EncodeToString(pair.Parsed.RawPayload) + "\n"
		}
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

func (ui *FyneUI) showGiftRecordWindow(roomID string) {
	rows := ui.loadRoomGiftRows(roomID)
	if len(rows) <= 1 {
		dialog.ShowInformation("提示", "暂无礼物记录", ui.mainWin)
		return
	}

	statusLabel := widget.NewLabel(fmt.Sprintf("共 %d 条礼物记录", len(rows)-1))

	table := widget.NewTable(
		func() (int, int) { return len(rows), len(rows[0]) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row < len(rows) && id.Col < len(rows[id.Row]) {
				cell.(*widget.Label).SetText(rows[id.Row][id.Col])
			}
		},
	)

	// 设置右键菜单
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row >= len(rows) {
			return
		}

		row := rows[id.Row]
		if len(row) < 5 {
			return
		}

		giftName := row[1]
		currentAnchor := row[4]

		// 如果没有主播，显示绑定选项
		if strings.TrimSpace(currentAnchor) == "" {
			ui.showBindAnchorMenu(roomID, giftName, func() {
				// 刷新礼物记录
				newRows := ui.loadRoomGiftRows(roomID)
				rows = newRows
				table.Refresh()
				statusLabel.SetText(fmt.Sprintf("共 %d 条礼物记录 (已刷新)", len(rows)-1))
			})
		}
	}

	win := ui.app.NewWindow(fmt.Sprintf("房间 %s 礼物记录", roomID))
	content := container.NewBorder(
		statusLabel,
		nil, nil, nil,
		container.NewScroll(table),
	)
	win.SetContent(content)
	win.Resize(fyne.NewSize(800, 500))
	win.Show()
}

// showBindAnchorMenu 显示绑定主播的菜单
func (ui *FyneUI) showBindAnchorMenu(roomID, giftName string, onBound func()) {
	if ui.mainWin == nil || ui.db == nil {
		return
	}

	// 查询该房间的主播列表
	anchors, err := ui.loadRoomAnchors(roomID)
	if err != nil || len(anchors) == 0 {
		dialog.ShowInformation("提示", "该房间暂无主播，请先在主播管理中添加主播", ui.mainWin)
		return
	}

	anchorOptions := make([]string, 0, len(anchors))
	anchorMap := make(map[string]string)
	for _, anchor := range anchors {
		option := fmt.Sprintf("%s | %s", anchor.ID, anchor.Name)
		anchorOptions = append(anchorOptions, option)
		anchorMap[option] = anchor.ID
	}

	anchorSelect := widget.NewSelect(anchorOptions, nil)
	anchorSelect.PlaceHolder = "选择主播"

	statusLabel := widget.NewLabel("")

	bindDialog := dialog.NewCustomConfirm(
		"绑定礼物到主播",
		"绑定",
		"取消",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("礼物: %s", giftName)),
			widget.NewLabel("选择接收主播:"),
			anchorSelect,
			statusLabel,
		),
		func(ok bool) {
			if !ok || anchorSelect.Selected == "" {
				return
			}

			anchorID := anchorMap[anchorSelect.Selected]
			if anchorID == "" {
				return
			}

			// 绑定礼物到主播
			_, err := ui.db.Exec(`
				INSERT INTO room_gift_bindings (room_id, gift_name, anchor_id)
				VALUES (?, ?, ?)
				ON CONFLICT(room_id, gift_name) DO UPDATE SET anchor_id=excluded.anchor_id
			`, roomID, giftName, anchorID)

			if err != nil {
				statusLabel.SetText(fmt.Sprintf("绑定失败: %v", err))
				return
			}

			// 更新现有的礼物记录
			_, err = ui.db.Exec(`
				UPDATE gift_records
				SET anchor_id = ?, anchor_name = (SELECT anchor_name FROM anchors WHERE anchor_id = ?)
				WHERE room_id = ? AND gift_name = ? AND (anchor_id IS NULL OR anchor_id = '')
			`, anchorID, anchorID, roomID, giftName)

			if err != nil {
				log.Printf("⚠️ 更新礼物记录失败: %v", err)
			}

			if onBound != nil {
				onBound()
			}
		},
		ui.mainWin,
	)

	bindDialog.Resize(fyne.NewSize(400, 200))
	bindDialog.Show()
}

type RoomAnchor struct {
	ID   string
	Name string
}

// loadRoomAnchors 加载房间主播列表
func (ui *FyneUI) loadRoomAnchors(roomID string) ([]RoomAnchor, error) {
	if ui.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := ui.db.Query(`
		SELECT anchor_id, anchor_name
		FROM room_anchors
		WHERE room_id = ?
		ORDER BY score DESC
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	anchors := make([]RoomAnchor, 0)
	for rows.Next() {
		var anchor RoomAnchor
		if err := rows.Scan(&anchor.ID, &anchor.Name); err != nil {
			continue
		}
		anchors = append(anchors, anchor)
	}

	return anchors, nil
}

// loadAllGiftNames 加载所有礼物名称
func (ui *FyneUI) loadAllGiftNames() []string {
	if ui.db == nil {
		return []string{}
	}

	rows, err := ui.db.Query(`
		SELECT DISTINCT gift_name
		FROM gifts
		WHERE COALESCE(is_deleted, 0) = 0
		ORDER BY gift_name
	`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		if strings.TrimSpace(name) != "" {
			names = append(names, name)
		}
	}

	return names
}

func (ui *FyneUI) loadRoomGiftRows(roomID string) [][]string {
	rows := [][]string{{"时间", "礼物", "数量", "钻石", "接收主播", "送礼用户"}}

	log.Printf("📊 [房间 %s] 开始加载礼物记录", roomID)

	if ui.db == nil {
		log.Printf("⚠️  [房间 %s] 数据库连接为空", roomID)
		return rows
	}

	query := `
		SELECT COALESCE(gr.create_time, gr.timestamp, CURRENT_TIMESTAMP), gr.gift_name, gr.gift_count, gr.gift_diamond_value,
		       COALESCE(gr.anchor_name, a.anchor_name, gr.anchor_id, '') AS receiver, gr.user_nickname
		FROM gift_records gr
		LEFT JOIN anchors a ON gr.anchor_id = a.anchor_id
		WHERE gr.room_id = ?
		ORDER BY COALESCE(gr.create_time, gr.timestamp) DESC
		LIMIT 200
	`

	log.Printf("🔍 [房间 %s] 执行查询: WHERE room_id = '%s'", roomID, roomID)

	data, err := ui.db.Query(query, roomID)
	if err != nil {
		log.Printf("❌ [房间 %s] 查询礼物记录失败: %v", roomID, err)
		return rows
	}
	defer data.Close()

	recordCount := 0
	for data.Next() {
		var ts time.Time
		var giftName, receiver, user sql.NullString
		var count, diamond int
		if err := data.Scan(&ts, &giftName, &count, &diamond, &receiver, &user); err != nil {
			log.Printf("⚠️  [房间 %s] 扫描记录失败: %v", roomID, err)
			continue
		}
		totalDiamond := diamond * count
		if totalDiamond == 0 {
			totalDiamond = diamond
		}
		rows = append(rows, []string{
			ts.Format("01-02 15:04:05"),
			giftName.String,
			fmt.Sprintf("%d", count),
			fmt.Sprintf("%d", totalDiamond),
			strings.TrimSpace(receiver.String),
			user.String,
		})
		recordCount++
	}

	log.Printf("✅ [房间 %s] 加载了 %d 条礼物记录（包含表头共 %d 行）", roomID, recordCount, len(rows))

	return rows
}

func (ui *FyneUI) loadRoomAnchorRows(roomID string) [][]string {
	rows := [][]string{{"主播ID", "主播名称", "绑定礼物", "礼物计数", "得分"}}
	if ui.db == nil {
		return rows
	}

	query := `
		SELECT anchor_id, anchor_name, bound_gifts, gift_count, score
		FROM room_anchors
		WHERE room_id = ?
		ORDER BY score DESC
	`

	data, err := ui.db.Query(query, roomID)
	if err != nil {
		return rows
	}
	defer data.Close()

	for data.Next() {
		var anchorID, anchorName, gifts string
		var giftCount, score int
		if err := data.Scan(&anchorID, &anchorName, &gifts, &giftCount, &score); err != nil {
			continue
		}
		rows = append(rows, []string{
			anchorID,
			anchorName,
			gifts,
			fmt.Sprintf("%d", giftCount),
			fmt.Sprintf("%d", score),
		})
	}

	return rows
}

func (ui *FyneUI) loadRoomSegmentRows(roomID string) [][]string {
	rows := [][]string{{"分段名称", "开始时间", "结束时间", "礼物总值"}}
	if ui.db == nil {
		return rows
	}

	query := `
		SELECT segment_name, start_time, end_time, total_gift_value
		FROM score_segments
		WHERE room_id = ?
		ORDER BY start_time DESC
		LIMIT 100
	`

	data, err := ui.db.Query(query, roomID)
	if err != nil {
		return rows
	}
	defer data.Close()

	for data.Next() {
		var name string
		var start, end sql.NullTime
		var total int
		if err := data.Scan(&name, &start, &end, &total); err != nil {
			continue
		}

		startStr := ""
		if start.Valid {
			startStr = start.Time.Format("01-02 15:04")
		}

		endStr := "进行中"
		if end.Valid {
			endStr = end.Time.Format("01-02 15:04")
		}

		rows = append(rows, []string{
			name,
			startStr,
			endStr,
			fmt.Sprintf("%d", total),
		})
	}

	return rows
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if v == "" {
			return 0
		}
		var i int
		fmt.Sscanf(v, "%d", &i)
		return i
	default:
		return 0
	}
}

func (ui *FyneUI) runOnMain(f func()) {
	if f == nil {
		return
	}
	if ui == nil || ui.app == nil {
		f()
		return
	}
	if drv := ui.app.Driver(); drv != nil {
		if runner, ok := drv.(interface{ RunOnMain(func()) }); ok {
			runner.RunOnMain(f)
			return
		}
	}
	f()
}

// updateOverviewStatus 更新概览页状态文本
func (ui *FyneUI) updateOverviewStatus(text string) {
	if ui.overviewStatus == nil {
		return
	}
	ui.overviewStatus.SetText(text)
}

// refreshAllUIComponents 刷新所有UI组件以响应主题变化
func (ui *FyneUI) refreshAllUIComponents() {
	if ui.mainWin != nil {
		// 刷新主窗口内容
		if ui.tabContainer != nil {
			ui.tabContainer.Refresh()
		}

		// 刷新所有房间Tab
		for _, roomTab := range ui.roomTabs {
			if roomTab.GiftTable != nil {
				roomTab.GiftTable.Refresh()
			}
			if roomTab.AnchorTable != nil {
				roomTab.AnchorTable.Refresh()
			}
			if roomTab.SegmentTable != nil {
				roomTab.SegmentTable.Refresh()
			}
			if roomTab.MessagesList != nil {
				roomTab.MessagesList.Refresh()
			}
			if roomTab.SubTabs != nil {
				roomTab.SubTabs.Refresh()
			}
		}

		// 刷新主窗口
		ui.mainWin.Content().Refresh()
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
