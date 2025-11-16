package ui

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// 尝试嵌入插件文件，如果不存在则在运行时加载
//go:embed assets/browser-monitor.zip
var embeddedPlugin []byte

// SettingsManager 设置管理器
type SettingsManager struct{}

// NewSettingsManager 创建设置管理器
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

// InstallPlugin 安装浏览器插件
func (s *SettingsManager) InstallPlugin() error {
	log.Println("📦 开始安装浏览器插件...")

	// 1. 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "browser-monitor")
	os.RemoveAll(tempDir) // 清理旧目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	log.Printf("📂 临时目录: %s", tempDir)

	// 2. 从嵌入文件或外部文件读取插件
	var zipData []byte
	var err error
	
	// 优先使用嵌入的插件
	if len(embeddedPlugin) > 0 {
		zipData = embeddedPlugin
		log.Println("使用内嵌插件文件")
	} else {
		// 如果嵌入文件不存在，尝试从外部读取
		externalPath := "assets/browser-monitor.zip"
		zipData, err = os.ReadFile(externalPath)
		if err != nil {
			return fmt.Errorf("读取插件文件失败: %w\n提示：请先运行 browser-monitor/pack.bat 打包插件", err)
		}
		log.Println("使用外部插件文件")
	}

	// 3. 解压到临时目录
	if err := s.unzipPlugin(zipData, tempDir); err != nil {
		return fmt.Errorf("解压插件失败: %w", err)
	}

	log.Println("✅ 插件文件已解压")

	// 4. 打开浏览器扩展页面
	if err := s.openExtensionsPage(); err != nil {
		log.Printf("⚠️  自动打开扩展页面失败: %v", err)
	}

	// 5. 提示用户手动加载
	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║          请手动加载插件                                  ║")
	log.Println("╠══════════════════════════════════════════════════════════╣")
	log.Println("║ 1. 在浏览器中打开 chrome://extensions/                  ║")
	log.Println("║ 2. 启用右上角的「开发者模式」                            ║")
	log.Println("║ 3. 点击「加载已解压的扩展程序」                          ║")
	log.Printf("║ 4. 选择目录: %-42s ║\n", tempDir)
	log.Println("╚══════════════════════════════════════════════════════════╝")

	return nil
}

// RemovePlugin 删除插件（清理临时目录）
func (s *SettingsManager) RemovePlugin() error {
	tempDir := filepath.Join(os.TempDir(), "browser-monitor")
	
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		return fmt.Errorf("插件目录不存在")
	}

	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("删除插件目录失败: %w", err)
	}

	log.Println("✅ 插件目录已清理")
	return nil
}

// unzipPlugin 解压插件
func (s *SettingsManager) unzipPlugin(zipData []byte, destDir string) error {
	// 创建临时 zip 文件
	tempZip := filepath.Join(os.TempDir(), "plugin.zip")
	if err := os.WriteFile(tempZip, zipData, 0644); err != nil {
		return err
	}
	defer os.Remove(tempZip)

	// 打开 zip 文件
	reader, err := zip.OpenReader(tempZip)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 解压所有文件
	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 解压文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// openExtensionsPage 打开浏览器扩展页面
func (s *SettingsManager) openExtensionsPage() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// 尝试 Chrome
		cmd = exec.Command("cmd", "/c", "start", "chrome://extensions/")
	case "darwin":
		cmd = exec.Command("open", "-a", "Google Chrome", "chrome://extensions/")
	case "linux":
		cmd = exec.Command("xdg-open", "chrome://extensions/")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}
