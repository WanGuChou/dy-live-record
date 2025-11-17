package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ManualRoomPuller 直接拉取抖音 WebSocket 消息
type ManualRoomPuller struct {
	roomID   string
	wsURL    string
	server   *WebSocketServer
	stopCh   chan struct{}
	stopOnce sync.Once
	conn     *websocket.Conn
	connMu   sync.Mutex
}

func NewManualRoomPuller(server *WebSocketServer, roomID, wsURL string) *ManualRoomPuller {
	return &ManualRoomPuller{
		roomID: roomID,
		wsURL:  wsURL,
		server: server,
		stopCh: make(chan struct{}),
	}
}

func (p *ManualRoomPuller) Start() {
	log.Printf("🚀 手动拉流启动: 房间 %s (%s)", p.roomID, p.wsURL)
	defer log.Printf("🛑 手动拉流结束: 房间 %s", p.roomID)
	for {
		select {
		case <-p.stopCh:
			return
		default:
		}
		if err := p.runOnce(); err != nil {
			log.Printf("⚠️ 手动拉流错误 [房间 %s]: %v", p.roomID, err)
			select {
			case <-time.After(5 * time.Second):
			case <-p.stopCh:
				return
			}
		} else {
			return
		}
	}
}

func (p *ManualRoomPuller) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
		p.closeConn()
	})
}

func (p *ManualRoomPuller) runOnce() error {
	header := http.Header{}
	header.Set("Origin", "https://live.douyin.com")
	header.Set("Referer", fmt.Sprintf("https://live.douyin.com/%s", p.roomID))
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  15 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(p.wsURL, header)
	if err != nil {
		return err
	}
	p.setConn(conn)
	defer p.closeConn()

	conn.SetReadLimit(32 * 1024 * 1024)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		return nil
	})

	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-p.stopCh:
			return nil
		case <-pingTicker.C:
			_ = conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		default:
			_, data, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			p.forwardPayload(data)
		}
	}
}

func (p *ManualRoomPuller) forwardPayload(payload []byte) {
	encoded := base64.StdEncoding.EncodeToString(payload)
	data := map[string]interface{}{
		"url":         p.wsURL,
		"payloadData": encoded,
	}
	p.server.handleDouyinMessage(data, true)
}

func ensureRoomParam(wsURL, roomID string) string {
	if roomID == "" || wsURL == "" {
		return wsURL
	}
	if strings.Contains(wsURL, "room_id=") {
		return wsURL
	}
	separator := "?"
	if strings.Contains(wsURL, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%sroom_id=%s", wsURL, separator, roomID)
}

func buildDefaultDouyinWS(roomID string) string {
	params := url.Values{}
	params.Set("aid", "6383")
	params.Set("version_code", "170400")
	params.Set("webcast_sdk_version", "1.3.0")
	params.Set("update_version_code", "170400")
	params.Set("room_id", roomID)
	params.Set("sub_room_id", roomID)
	params.Set("live_id", "1")
	params.Set("did_rule", "3")
	params.Set("device_platform", "web")
	params.Set("device_type", "windows")
	params.Set("user_unique_id", fmt.Sprintf("manual_%s", roomID))
	params.Set("enter_from", "manual")
	params.Set("cookie_enabled", "true")
	params.Set("screen_width", "1920")
	params.Set("screen_height", "1080")
	params.Set("browser_language", "zh-CN")
	params.Set("browser_platform", "Win32")
	params.Set("browser_name", "Chrome")
	params.Set("browser_version", "126.0.0.0")
	params.Set("browser_online", "true")
	params.Set("tz_name", "Asia/Shanghai")
	params.Set("msToken", randomMsToken())
	params.Set("X-Bogus", randomMsToken())
	return "wss://webcast3-ws-web-lq.douyin.com/webcast/im/push/v2/?" + params.Encode()
}

func randomMsToken() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 48)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (p *ManualRoomPuller) setConn(conn *websocket.Conn) {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	p.conn = conn
}

func (p *ManualRoomPuller) closeConn() {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
}
