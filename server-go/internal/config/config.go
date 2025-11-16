package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 系统配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	License  LicenseConfig  `json:"license"`
	Browser  BrowserConfig  `json:"browser"`
	Debug    DebugConfig    `json:"debug"`
}

// ServerConfig WebSocket 服务器配置
type ServerConfig struct {
	Port int `json:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `json:"path"`
}

// LicenseConfig 许可证配置
type LicenseConfig struct {
	ServerURL     string `json:"server_url"`
	PublicKeyPath string `json:"public_key_path"`
	LocalPath     string `json:"local_path"`
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	ChromePath     string   `json:"chrome_path"`
	LaunchArgs     []string `json:"launch_args"`
	PluginPath     string   `json:"plugin_path"`
	AutoInstall    bool     `json:"auto_install"`
}

// DebugConfig 调试配置
type DebugConfig struct {
	Enabled      bool `json:"enabled"`       // 启用调试模式
	SkipLicense  bool `json:"skip_license"`  // 跳过 License 验证（仅调试模式下有效）
	VerboseLog   bool `json:"verbose_log"`   // 详细日志输出
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "./data.db",
		},
		License: LicenseConfig{
			ServerURL:     "https://license.example.com",
			PublicKeyPath: "./rsa_public.pem",
			LocalPath:     "./license.dat",
		},
		Browser: BrowserConfig{
			ChromePath: "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			LaunchArgs: []string{
				"--silent-debugger-extension-api",
				"--disable-blink-features=AutomationControlled",
			},
			PluginPath:  "./browser-monitor.zip",
			AutoInstall: true,
		},
		Debug: DebugConfig{
			Enabled:     false, // 默认关闭调试模式
			SkipLicense: false, // 默认不跳过 License
			VerboseLog:  false, // 默认不输出详细日志
		},
	}
}

// Load 从文件加载配置
func Load() (*Config, error) {
	configPath := "./config.json"
	
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := Default()
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	configPath := "./config.json"
	return os.WriteFile(configPath, data, 0644)
}

// GetExecutableDir 获取可执行文件所在目录
func GetExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}
