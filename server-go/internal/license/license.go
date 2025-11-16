package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// LicenseData 许可证数据结构
type LicenseData struct {
	LicenseKey          string    `json:"license_key"`
	SoftwareID          string    `json:"software_id"`
	Version             string    `json:"version"`
	CustomerID          string    `json:"customer_id"`
	ExpiryDate          time.Time `json:"expiry_date"`
	IssueDate           time.Time `json:"issue_date"`
	HardwareFingerprint string    `json:"hardware_fingerprint"`
	ActivationTime      time.Time `json:"activation_time"`
	MaxActivations      int       `json:"max_activations"`
	CurrentActivations  int       `json:"current_activations"`
	LicenseType         string    `json:"license_type"`
	Features            map[string]interface{} `json:"features"`
}

// Manager 许可证管理器
type Manager struct {
	serverURL     string
	publicKey     *rsa.PublicKey
	localLicPath  string
}

// NewManager 创建许可证管理器
func NewManager(serverURL, publicKeyPath string) *Manager {
	pubKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		// 如果公钥加载失败，使用硬编码的公钥（生产环境）
		pubKey = getEmbeddedPublicKey()
	}

	return &Manager{
		serverURL:    serverURL,
		publicKey:    pubKey,
		localLicPath: "./license.dat",
	}
}

// LoadLocal 加载本地许可证
func (m *Manager) LoadLocal() (string, error) {
	data, err := os.ReadFile(m.localLicPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveLocal 保存许可证到本地
func (m *Manager) SaveLocal(license string) error {
	return os.WriteFile(m.localLicPath, []byte(license), 0600)
}

// Validate 校验许可证（离线模式）
func (m *Manager) Validate(licenseString string) (bool, time.Time, error) {
	// 1. Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(licenseString)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("许可证格式错误: %w", err)
	}

	// 2. 分离数据和签名（最后256字节是RSA签名）
	if len(decoded) < 256 {
		return false, time.Time{}, errors.New("许可证数据不完整")
	}

	dataBytes := decoded[:len(decoded)-256]
	signatureBytes := decoded[len(decoded)-256:]

	// 3. 验证签名
	hashed := sha256.Sum256(dataBytes)
	err = rsa.VerifyPKCS1v15(m.publicKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("许可证签名验证失败: %w", err)
	}

	// 4. 解析许可证数据
	var licData LicenseData
	if err := json.Unmarshal(dataBytes, &licData); err != nil {
		return false, time.Time{}, fmt.Errorf("许可证数据解析失败: %w", err)
	}

	// 5. 检查有效期（必须使用互联网时间）
	ntpTime, err := getNTPTime()
	if err != nil {
		return false, time.Time{}, fmt.Errorf("获取网络时间失败（防止系统时间被篡改）: %w", err)
	}

	if ntpTime.After(licData.ExpiryDate) {
		return false, licData.ExpiryDate, errors.New("许可证已过期")
	}

	// 6. 检查硬件指纹
	if licData.HardwareFingerprint != "" {
		currentFingerprint := GetHardwareFingerprint()
		if currentFingerprint != licData.HardwareFingerprint {
			return false, licData.ExpiryDate, errors.New("硬件指纹不匹配")
		}
	}

	return true, licData.ExpiryDate, nil
}

// ValidateOnline 在线校验许可证（调用 server-active API）
func (m *Manager) ValidateOnline(licenseKey string) (bool, string, error) {
	fingerprint := GetHardwareFingerprint()
	
	req := map[string]string{
		"license_key":           licenseKey,
		"hardware_fingerprint":  fingerprint,
	}

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(m.serverURL+"/api/v1/licenses/validate", "application/json", nil)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if valid, ok := result["valid"].(bool); ok && valid {
		if licData, ok := result["license_data"].(string); ok {
			return true, licData, nil
		}
	}

	return false, "", errors.New(result["message"].(string))
}

// loadPublicKey 从文件加载RSA公钥
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("不是RSA公钥")
	}

	return rsaPub, nil
}

// getEmbeddedPublicKey 获取硬编码的公钥（生产环境）
func getEmbeddedPublicKey() *rsa.PublicKey {
	// 这里应该硬编码实际的公钥
	// 示例：从PEM字符串解析
	pemStr := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----`
	
	block, _ := pem.Decode([]byte(pemStr))
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	return pub.(*rsa.PublicKey)
}

// getNTPTime 从NTP服务器获取标准时间
func getNTPTime() (time.Time, error) {
	// 简化实现：调用公共时间API
	resp, err := http.Get("http://worldtimeapi.org/api/timezone/Etc/UTC")
	if err != nil {
		return time.Now(), err
	}
	defer resp.Body.Close()

	var result struct {
		Datetime time.Time `json:"datetime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Now(), err
	}

	return result.Datetime, nil
}
