package license

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// GetHardwareFingerprint 获取硬件指纹
func GetHardwareFingerprint() string {
	components := []string{
		getCPUInfo(),
		getMotherboardInfo(),
		getDiskSerial(),
		getMACAddress(),
	}

	// 过滤空值
	var validComponents []string
	for _, comp := range components {
		if comp != "" {
			validComponents = append(validComponents, comp)
		}
	}

	// 排序确保一致性
	sort.Strings(validComponents)

	// 组合并哈希
	combined := strings.Join(validComponents, "|")
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// getCPUInfo 获取CPU信息
func getCPUInfo() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// 使用WMIC获取CPU序列号
	cmd := exec.Command("wmic", "cpu", "get", "ProcessorId")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

// getMotherboardInfo 获取主板信息
func getMotherboardInfo() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("wmic", "baseboard", "get", "SerialNumber")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

// getDiskSerial 获取硬盘序列号
func getDiskSerial() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("wmic", "diskdrive", "get", "SerialNumber")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

// getMACAddress 获取MAC地址
func getMACAddress() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("getmac", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// 提取第一个物理MAC地址
		parts := strings.Split(lines[0], ",")
		if len(parts) > 0 {
			mac := strings.Trim(parts[0], `"`)
			return strings.ReplaceAll(mac, "-", ":")
		}
	}
	return ""
}
