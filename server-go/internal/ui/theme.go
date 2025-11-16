package ui

import (
	"image/color"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ChineseTheme 支持中文的自定义主题
type ChineseTheme struct {
	fyne.Theme
}

// NewChineseTheme 创建支持中文的主题
func NewChineseTheme() fyne.Theme {
	return &ChineseTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Font 返回支持中文的字体
func (t *ChineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 使用 Fyne 内置字体，支持 CJK 字符
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		if style.Italic {
			return theme.DefaultTheme().Font(style)
		}
		return theme.DefaultTheme().Font(style)
	}
	if style.Italic {
		return theme.DefaultTheme().Font(style)
	}
	// 使用默认字体，Fyne 会自动加载系统中文字体
	return theme.DefaultTheme().Font(style)
}

// Color 返回主题颜色
func (t *ChineseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

// Icon 返回主题图标
func (t *ChineseTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size 返回主题尺寸
func (t *ChineseTheme) Size(name fyne.ThemeSizeName) float32 {
	// 稍微增大字体以便中文显示
	if name == theme.SizeNameText {
		return theme.DefaultTheme().Size(name) + 1
	}
	return theme.DefaultTheme().Size(name)
}
