# 主题弹窗修复文档

## 修复日期
2025-11-21

## 问题描述

每次启动应用程序时，都会弹出"主题已更新"的对话框，这是因为在初始化主题下拉框时，`SetSelected()` 方法会触发回调函数，从而显示对话框。

## 原因分析

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`

原代码：
```go
themeSelect := widget.NewSelect([]string{"系统默认", "浅色", "深色"}, func(val string) {
    ui.applyTheme(val)
    ui.saveThemePreference(val)
    // 提示用户主题已更改
    if ui.mainWin != nil {
        dialog.ShowInformation("主题已更新", "主题设置已保存并应用", ui.mainWin)
    }
})
themeSelect.SetSelected(ui.userTheme)  // ⚠️ 这里会触发回调函数
```

问题：
- `SetSelected()` 会立即触发回调函数
- 回调函数中无条件显示对话框
- 导致每次启动都弹出提示

## 解决方案

使用标志位 `isInitializing` 来区分是初始化还是用户手动更改：

```go
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
```

## 工作流程

### 初始化阶段（启动时）
1. `isInitializing = true`
2. 创建下拉框和回调函数
3. `SetSelected()` 触发回调
4. 回调检查 `!isInitializing` → false，不显示对话框 ✅
5. `isInitializing = false`

### 用户手动更改主题
1. 用户点击下拉框选择新主题
2. 触发回调函数
3. 回调检查 `!isInitializing` → true，显示对话框 ✅

## 修改位置

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`  
**行号**: 2357-2367

## 验证结果

✅ 语法检查通过（`go fmt`）  
✅ 使用标志位正确区分初始化和手动更改  
✅ 不影响主题功能的正常使用

## 预期效果

### 启动应用程序
- ✅ 不再弹出"主题已更新"对话框
- ✅ 主题正常加载并应用
- ✅ 下拉框显示当前主题

### 用户手动切换主题
- ✅ 主题正常切换
- ✅ 显示"主题已更新"确认对话框
- ✅ 设置保存成功

## 技术细节

### 标志位作用域
- `isInitializing` 是局部变量
- 作用域仅限于函数内
- 通过闭包被回调函数捕获

### 时序说明
```
启动时:
  isInitializing = true
    ↓
  创建 themeSelect
    ↓
  SetSelected(ui.userTheme) → 触发回调
    ↓
  回调: if !isInitializing (false) → 不显示对话框
    ↓
  isInitializing = false
    ↓
  完成

用户更改:
  用户选择新主题 → 触发回调
    ↓
  回调: if !isInitializing (true) → 显示对话框
```

## 总结

通过添加一个简单的布尔标志位，成功区分了初始化和用户手动操作，避免了启动时不必要的提示对话框。这是一个常见的 UI 编程模式，用于处理初始化时的事件触发问题。
