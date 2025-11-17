package parser

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strings"
)

// PushFrame Protobuf PushFrame 结构
type PushFrame struct {
	SeqID           string
	LogID           string
	Service         string
	Method          string
	HeadersList     map[string]string
	PayloadEncoding string
	PayloadType     string
	Payload         []byte
	LogIDNew        string
}

// Response Protobuf Response 结构
type Response struct {
	Messages          []*Message
	Cursor            string
	FetchInterval     string
	Now               string
	InternalExt       string
	FetchType         int32
	HeartbeatDuration string
	NeedAck           bool
	PushServer        string
	LiveCursor        string
	HistoryNoMore     bool
}

// Message Protobuf Message 结构
type Message struct {
	Method        string
	Payload       []byte
	MsgID         string
	MsgType       int32
	Offset        string
	NeedWrdsStore bool
	WrdsVersion   string
	WrdsSubKey    string
}

// DecodePushFrame 解码 PushFrame
func DecodePushFrame(data []byte) (*PushFrame, error) {
	bb := NewByteBuffer(data)
	frame := &PushFrame{
		HeadersList: make(map[string]string),
	}

	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // seqId
			frame.SeqID, _ = bb.ReadVarint64(true)
		case 2: // logId
			frame.LogID, _ = bb.ReadVarint64(true)
		case 3: // service
			frame.Service, _ = bb.ReadVarint64(true)
		case 4: // method
			frame.Method, _ = bb.ReadVarint64(true)
		case 5: // headersList (map<string, string>)
			oldLimit, _ := bb.PushTemporaryLength()
			var key, value string
			for !bb.IsAtEnd() {
				tag2, err := bb.ReadVarint32()
				if err != nil {
					break
				}
				fieldNumber2 := tag2 >> 3
				if fieldNumber2 == 0 {
					break
				}
				if fieldNumber2 == 1 {
					length, _ := bb.ReadVarint32()
					key, _ = bb.ReadString(int(length))
				} else if fieldNumber2 == 2 {
					length, _ := bb.ReadVarint32()
					value, _ = bb.ReadString(int(length))
				} else {
					bb.SkipUnknownField(int(tag2 & 7))
				}
			}
			if key != "" && value != "" {
				frame.HeadersList[key] = value
			}
			bb.limit = oldLimit
		case 6: // payloadEncoding
			length, _ := bb.ReadVarint32()
			frame.PayloadEncoding, _ = bb.ReadString(int(length))
		case 7: // payloadType
			length, _ := bb.ReadVarint32()
			frame.PayloadType, _ = bb.ReadString(int(length))
		case 8: // payload
			length, _ := bb.ReadVarint32()
			frame.Payload, _ = bb.ReadBytes(int(length))
		case 9: // logIdNew
			length, _ := bb.ReadVarint32()
			frame.LogIDNew, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return frame, nil
}

// DecodeResponse 解码 Response
func DecodeResponse(data []byte) (*Response, error) {
	bb := NewByteBuffer(data)
	resp := &Response{
		Messages: make([]*Message, 0),
	}

	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // messages (repeated Message)
			oldLimit, _ := bb.PushTemporaryLength()
			msg, err := DecodeMessage(bb)
			if err == nil {
				resp.Messages = append(resp.Messages, msg)
			}
			bb.limit = oldLimit
		case 2: // cursor
			length, _ := bb.ReadVarint32()
			resp.Cursor, _ = bb.ReadString(int(length))
		case 3: // fetchInterval
			resp.FetchInterval, _ = bb.ReadVarint64(false)
		case 4: // now
			resp.Now, _ = bb.ReadVarint64(false)
		case 5: // internalExt
			length, _ := bb.ReadVarint32()
			resp.InternalExt, _ = bb.ReadString(int(length))
		case 6: // fetchType
			resp.FetchType, _ = bb.ReadVarint32()
		case 8: // heartbeatDuration
			resp.HeartbeatDuration, _ = bb.ReadVarint64(false)
		case 9: // needAck
			b, _ := bb.ReadByte()
			resp.NeedAck = b != 0
		case 10: // pushServer
			length, _ := bb.ReadVarint32()
			resp.PushServer, _ = bb.ReadString(int(length))
		case 11: // liveCursor
			length, _ := bb.ReadVarint32()
			resp.LiveCursor, _ = bb.ReadString(int(length))
		case 12: // historyNoMore
			b, _ := bb.ReadByte()
			resp.HistoryNoMore = b != 0
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return resp, nil
}

// DecodeMessage 解码 Message
func DecodeMessage(bb *ByteBuffer) (*Message, error) {
	msg := &Message{}

	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // method
			length, _ := bb.ReadVarint32()
			msg.Method, _ = bb.ReadString(int(length))
		case 2: // payload
			length, _ := bb.ReadVarint32()
			msg.Payload, _ = bb.ReadBytes(int(length))
		case 3: // msgId
			msg.MsgID, _ = bb.ReadVarint64(false)
		case 4: // msgType
			msg.MsgType, _ = bb.ReadVarint32()
		case 5: // offset
			msg.Offset, _ = bb.ReadVarint64(false)
		case 6: // needWrdsStore
			b, _ := bb.ReadByte()
			msg.NeedWrdsStore = b != 0
		case 7: // wrdsVersion
			msg.WrdsVersion, _ = bb.ReadVarint64(false)
		case 8: // wrdsSubKey
			length, _ := bb.ReadVarint32()
			msg.WrdsSubKey, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// ParseDouyinMessage 解析抖音消息（主入口）
func ParseDouyinMessage(payloadData, url string) ([]map[string]interface{}, error) {
	buffer, err := decodePayloadBuffer(payloadData)
	if err != nil {
		return nil, err
	}

	responsePayload, err := decodeResponsePayload(buffer)
	if err != nil {
		return nil, err
	}

	response, err := DecodeResponse(responsePayload)
	if err != nil || len(response.Messages) == 0 {
		// 再尝试一次直接解析原始数据（兼容部分异常场景）
		if fallbackResp, fallbackErr := DecodeResponse(decompressIfGzip(buffer)); fallbackErr == nil && len(fallbackResp.Messages) > 0 {
			response = fallbackResp
		} else {
			return nil, fmt.Errorf("Response解析失败: %w", err)
		}
	}

	results := make([]map[string]interface{}, 0, len(response.Messages))
	for _, msg := range response.Messages {
		if msg == nil || msg.Method == "" || msg.Payload == nil {
			continue
		}
		payload := decompressIfGzip(msg.Payload)
		parsed := ParseMessagePayload(msg.Method, payload)
		if parsed != nil {
			results = append(results, parsed)
		}
	}

	return results, nil
}

func decodePayloadBuffer(payloadData string) ([]byte, error) {
	data := strings.TrimSpace(payloadData)
	if data == "" {
		return nil, fmt.Errorf("payloadData为空")
	}

	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		return decoded, nil
	} else if rawDecoded, rawErr := base64.RawStdEncoding.DecodeString(data); rawErr == nil {
		return rawDecoded, nil
	}

	log.Printf("⚠️ Payload 不是标准Base64，按原始文本处理")
	return []byte(data), nil
}

func decodeResponsePayload(buffer []byte) ([]byte, error) {
	pushFrame, err := DecodePushFrame(buffer)
	if err != nil {
		log.Printf("⚠️ PushFrame解析失败，直接尝试解析 Response: %v", err)
		return decompressIfGzip(buffer), nil
	}

	if pushFrame.Payload == nil {
		log.Printf("⚠️ PushFrame 不包含 Payload，使用原始数据")
		return decompressIfGzip(buffer), nil
	}

	payload := pushFrame.Payload
	if encoding := strings.ToLower(pushFrame.PayloadEncoding); encoding == "gzip" {
		payload = decompressIfGzip(payload)
	} else if compressType, ok := pushFrame.HeadersList["compress_type"]; ok && strings.EqualFold(compressType, "gzip") {
		payload = decompressIfGzip(payload)
	} else {
		payload = decompressIfGzip(payload)
	}

	return payload, nil
}

func decompressIfGzip(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	if data[0] == 0x1f && data[1] == 0x8b {
		if decompressed, err := gzipDecompress(data); err == nil {
			return decompressed
		} else {
			log.Printf("⚠️ GZIP 解压失败: %v", err)
		}
	}
	return data
}

func gzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
