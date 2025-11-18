package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	douyinLive "github.com/jwwsjlm/douyinLive"
	newdouyin "github.com/jwwsjlm/douyinLive/generated/new_douyin"

	"dy-live-monitor/internal/parser"
)

type manualRoomConnection struct {
	roomID         string
	live           *douyinLive.DouyinLive
	subscriptionID string
}

// startManualRoom launches a standalone Douyin WSS session for a room.
func (ui *FyneUI) startManualRoom(roomID string) error {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return errors.New("房间号不能为空")
	}

	ui.roomConnMu.Lock()
	if _, exists := ui.manualRooms[roomID]; exists {
		ui.roomConnMu.Unlock()
		return fmt.Errorf("房间 %s 已在监听中", roomID)
	}
	ui.roomConnMu.Unlock()

	logger := log.New(os.Stdout, fmt.Sprintf("[手动房间 %s] ", roomID), log.LstdFlags)
	live, err := douyinLive.NewDouyinLive(roomID, logger)
	if err != nil {
		return err
	}

	conn := &manualRoomConnection{
		roomID: roomID,
		live:   live,
	}

	conn.subscriptionID = live.Subscribe(func(eventData *newdouyin.Webcast_Im_Message) {
		ui.handleManualEvent(roomID, eventData)
	})

	ui.roomConnMu.Lock()
	ui.manualRooms[roomID] = conn
	ui.roomConnMu.Unlock()

	ui.AddOrUpdateRoom(roomID)
	ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 已连接", roomID))

	go func() {
		live.Start()
		ui.cleanupManualRoom(roomID)
		ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 连接结束", roomID))
	}()

	return nil
}

func (ui *FyneUI) stopManualRoom(roomID string) {
	conn := ui.detachManualRoom(roomID)
	if conn == nil {
		return
	}

	if conn.subscriptionID != "" {
		conn.live.Unsubscribe(conn.subscriptionID)
	}
	conn.live.Close()
}

func (ui *FyneUI) cleanupManualRoom(roomID string) {
	conn := ui.detachManualRoom(roomID)
	if conn == nil {
		return
	}
	if conn.subscriptionID != "" {
		conn.live.Unsubscribe(conn.subscriptionID)
	}
}

func (ui *FyneUI) detachManualRoom(roomID string) *manualRoomConnection {
	ui.roomConnMu.Lock()
	defer ui.roomConnMu.Unlock()

	conn, exists := ui.manualRooms[roomID]
	if !exists {
		return nil
	}

	delete(ui.manualRooms, roomID)
	return conn
}

func (ui *FyneUI) handleManualEvent(roomID string, eventData *newdouyin.Webcast_Im_Message) {
	if eventData == nil {
		return
	}

	ui.AddOrUpdateRoom(roomID)

	parsed, err := parser.ParseProtoMessage(eventData.Method, eventData.Payload)
	if err != nil {
		ui.AddParsedMessage(roomID, fmt.Sprintf("解析 %s 失败: %v", eventData.Method, err))
		return
	}

	ui.recordParsedMessage(roomID, parsed, true)
}
