package main

const (
	// Version 版本号
	Version = "v3.3.0"

	// BuildDate 构建日期
	BuildDate = "2025-11-16"

	// AppName 应用名称
	AppName = "抖音直播监控系统"

	// AppNameEN 应用英文名称
	AppNameEN = "Douyin Live Monitor"
)

// GetVersionInfo 获取版本信息
func GetVersionInfo() string {
	return AppName + " " + Version + " (" + BuildDate + ")"
}
