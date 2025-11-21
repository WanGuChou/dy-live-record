# 编译错误修复 - toInt 重复声明

## 问题描述

启动 server-go 时报错：
```
# dy-live-monitor/internal/ui
internal\ui\manual_room.go:289:6: toInt redeclared in this block
	internal\ui\fyne_ui.go:3639:6: other declaration of toInt
```

## 原因分析

在同一个包 (`internal/ui`) 中，`toInt` 函数被声明了两次：
1. `fyne_ui.go:3639` - 已有的函数
2. `manual_room.go:289` - 新添加时重复声明

## 解决方案

删除 `manual_room.go` 中重复的 `toInt` 函数定义，保留 `toString` 函数（因为 `fyne_ui.go` 中没有此函数）。

### 修改内容

**文件**: `/workspace/server-go/internal/ui/manual_room.go`

**删除的代码**:
```go
// 辅助函数：转换接口类型为整数
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
```

**保留的代码**:
```go
// 辅助函数：转换接口类型为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
```

## 函数使用说明

### ui 包中的辅助函数

| 函数 | 位置 | 说明 |
|-----|------|------|
| `toInt` | `fyne_ui.go:3639` | 在整个 ui 包中使用 |
| `toString` | `manual_room.go:281` | 在整个 ui 包中使用 |

### server 包中的辅助函数

`internal/server/websocket.go` 中也有自己的 `toString` 和 `toInt` 函数（这是正常的，因为属于不同的包）：
- `toString` - `websocket.go:365`
- `toInt` - `websocket.go:372`

## 验证结果

✅ 使用 `go vet` 检查：没有重复声明错误  
✅ 包结构正常  
✅ 只保留一个 `toInt` 函数在 `fyne_ui.go` 中

## 总结

现在编译不会再报 `toInt redeclared` 错误了。程序应该可以正常启动。
