package parser

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ByteBuffer Go 版本的 ByteBuffer（模仿 dycast 的实现）
type ByteBuffer struct {
	bytes  []byte
	offset int
	limit  int
}

// NewByteBuffer 创建 ByteBuffer
func NewByteBuffer(data []byte) *ByteBuffer {
	return &ByteBuffer{
		bytes:  data,
		offset: 0,
		limit:  len(data),
	}
}

// IsAtEnd 检查是否到达末尾
func (bb *ByteBuffer) IsAtEnd() bool {
	return bb.offset >= bb.limit
}

// Advance 前进指定字节数
func (bb *ByteBuffer) Advance(count int) (int, error) {
	oldOffset := bb.offset
	if oldOffset+count > bb.limit {
		return 0, errors.New("read past limit")
	}
	bb.offset += count
	return oldOffset, nil
}

// ReadByte 读取一个字节
func (bb *ByteBuffer) ReadByte() (byte, error) {
	offset, err := bb.Advance(1)
	if err != nil {
		return 0, err
	}
	return bb.bytes[offset], nil
}

// ReadBytes 读取指定数量的字节
func (bb *ByteBuffer) ReadBytes(count int) ([]byte, error) {
	offset, err := bb.Advance(count)
	if err != nil {
		return nil, err
	}
	return bb.bytes[offset : offset+count], nil
}

// ReadVarint32 读取 varint32
func (bb *ByteBuffer) ReadVarint32() (int32, error) {
	var value uint32
	var shift uint
	
	for {
		b, err := bb.ReadByte()
		if err != nil {
			return 0, err
		}
		
		if shift < 32 {
			value |= uint32(b&0x7f) << shift
		}
		shift += 7
		
		if (b & 0x80) == 0 {
			break
		}
	}
	
	return int32(value), nil
}

// ReadVarint64 读取 varint64
func (bb *ByteBuffer) ReadVarint64(unsigned bool) (string, error) {
	var part0, part1, part2 uint32
	
	b, err := bb.ReadByte()
	if err != nil {
		return "", err
	}
	part0 = uint32(b & 0x7f)
	
	if (b & 0x80) != 0 {
		b, _ = bb.ReadByte()
		part0 |= uint32(b&0x7f) << 7
		if (b & 0x80) != 0 {
			b, _ = bb.ReadByte()
			part0 |= uint32(b&0x7f) << 14
			if (b & 0x80) != 0 {
				b, _ = bb.ReadByte()
				part0 |= uint32(b&0x7f) << 21
				if (b & 0x80) != 0 {
					b, _ = bb.ReadByte()
					part1 = uint32(b & 0x7f)
					if (b & 0x80) != 0 {
						b, _ = bb.ReadByte()
						part1 |= uint32(b&0x7f) << 7
						if (b & 0x80) != 0 {
							b, _ = bb.ReadByte()
							part1 |= uint32(b&0x7f) << 14
							if (b & 0x80) != 0 {
								b, _ = bb.ReadByte()
								part1 |= uint32(b&0x7f) << 21
								if (b & 0x80) != 0 {
									b, _ = bb.ReadByte()
									part2 = uint32(b & 0x7f)
									if (b & 0x80) != 0 {
										b, _ = bb.ReadByte()
										part2 |= uint32(b&0x7f) << 7
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	low := part0 | (part1 << 28)
	high := (part1 >> 4) | (part2 << 24)
	
	if high == 0 {
		return fmt.Sprintf("%d", low), nil
	}
	
	// 64位数值
	result := uint64(high)*4294967296 + uint64(low)
	return fmt.Sprintf("%d", result), nil
}

// ReadString 读取字符串（UTF-8）
func (bb *ByteBuffer) ReadString(length int) (string, error) {
	bytes, err := bb.ReadBytes(length)
	if err != nil {
		return "", err
	}
	
	// UTF-8 解码
	result := ""
	i := 0
	for i < len(bytes) {
		c1 := bytes[i]
		
		if (c1 & 0x80) == 0 {
			// 单字节字符
			result += string(rune(c1))
			i++
		} else if (c1 & 0xe0) == 0xc0 {
			// 双字节字符
			if i+1 >= len(bytes) {
				result += "\uFFFD"
				break
			}
			c2 := bytes[i+1]
			if (c2 & 0xc0) != 0x80 {
				result += "\uFFFD"
				i++
			} else {
				c := ((uint32(c1) & 0x1f) << 6) | (uint32(c2) & 0x3f)
				if c < 0x80 {
					result += "\uFFFD"
				} else {
					result += string(rune(c))
				}
				i += 2
			}
		} else if (c1 & 0xf0) == 0xe0 {
			// 三字节字符
			if i+2 >= len(bytes) {
				result += "\uFFFD"
				break
			}
			c2 := bytes[i+1]
			c3 := bytes[i+2]
			if ((c2 | uint32(c3)<<8) & 0xc0c0) != 0x8080 {
				result += "\uFFFD"
				i++
			} else {
				c := ((uint32(c1) & 0x0f) << 12) | ((uint32(c2) & 0x3f) << 6) | (uint32(c3) & 0x3f)
				if c < 0x0800 || (c >= 0xd800 && c <= 0xdfff) {
					result += "\uFFFD"
				} else {
					result += string(rune(c))
				}
				i += 3
			}
		} else if (c1 & 0xf8) == 0xf0 {
			// 四字节字符
			if i+3 >= len(bytes) {
				result += "\uFFFD"
				break
			}
			c2 := bytes[i+1]
			c3 := bytes[i+2]
			c4 := bytes[i+3]
			if ((c2 | uint32(c3)<<8 | uint32(c4)<<16) & 0xc0c0c0) != 0x808080 {
				result += "\uFFFD"
				i++
			} else {
				c := ((uint32(c1) & 0x07) << 18) | ((uint32(c2) & 0x3f) << 12) | ((uint32(c3) & 0x3f) << 6) | (uint32(c4) & 0x3f)
				if c < 0x10000 || c > 0x10ffff {
					result += "\uFFFD"
				} else {
					c -= 0x10000
					result += string(rune((c>>10) + 0xd800))
					result += string(rune((c&0x3ff) + 0xdc00))
				}
				i += 4
			}
		} else {
			result += "\uFFFD"
			i++
		}
	}
	
	return result, nil
}

// PushTemporaryLength 读取长度并设置临时 limit
func (bb *ByteBuffer) PushTemporaryLength() (int, error) {
	length, err := bb.ReadVarint32()
	if err != nil {
		return 0, err
	}
	
	oldLimit := bb.limit
	bb.limit = bb.offset + int(length)
	
	if bb.limit > len(bb.bytes) {
		return 0, errors.New("length exceeds buffer size")
	}
	
	return oldLimit, nil
}

// SkipUnknownField 跳过未知字段
func (bb *ByteBuffer) SkipUnknownField(wireType int) error {
	switch wireType {
	case 0: // Varint
		for {
			b, err := bb.ReadByte()
			if err != nil {
				return err
			}
			if (b & 0x80) == 0 {
				break
			}
		}
	case 1: // 64-bit
		_, err := bb.Advance(8)
		return err
	case 2: // Length-delimited
		length, err := bb.ReadVarint32()
		if err != nil {
			return err
		}
		_, err = bb.Advance(int(length))
		return err
	case 3: // Start group (deprecated)
		for !bb.IsAtEnd() {
			tag, err := bb.ReadVarint32()
			if err != nil {
				return err
			}
			wireType := int(tag & 7)
			if wireType == 4 {
				break
			}
			if err := bb.SkipUnknownField(wireType); err != nil {
				return err
			}
		}
	case 4: // End group (deprecated)
		// Do nothing
	case 5: // 32-bit
		_, err := bb.Advance(4)
		return err
	default:
		return fmt.Errorf("invalid wire type: %d", wireType)
	}
	return nil
}

// Read32Fixed 读取固定32位整数
func (bb *ByteBuffer) Read32Fixed() (uint32, error) {
	bytes, err := bb.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bytes), nil
}

// Read64Fixed 读取固定64位整数
func (bb *ByteBuffer) Read64Fixed() (uint64, error) {
	bytes, err := bb.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(bytes), nil
}

// SkipToEnd 跳到buffer末尾（用于跳过整个嵌套结构）
func (bb *ByteBuffer) SkipToEnd() {
	bb.offset = bb.limit
}
