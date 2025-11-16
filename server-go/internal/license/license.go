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
	"log"
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
	var pubKey *rsa.PublicKey
	var err error
	
	// 尝试从文件加载公钥
	if publicKeyPath != "" {
		pubKey, err = loadPublicKey(publicKeyPath)
		if err != nil {
			log.Printf("⚠️  公钥文件加载失败: %v", err)
		}
	}
	
	// 如果文件加载失败，尝试使用嵌入的公钥
	if pubKey == nil {
		pubKey = getEmbeddedPublicKey()
	}
	
	// 如果仍然没有公钥，创建一个临时的（仅用于调试）
	if pubKey == nil {
		log.Println("⚠️  警告：未找到有效公钥，License 验证将无法工作")
		log.Println("⚠️  请配置 publicKeyPath 或启用调试模式")
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
	// 检查公钥是否可用
	if m.publicKey == nil {
		return false, time.Time{}, errors.New("公钥未配置，无法验证许可证")
	}
	
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

	_, _ = json.Marshal(req)
	// TODO: 实现在线校验（发送 reqBody 到服务器）
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
	// TODO: 生产环境需要硬编码实际的公钥
	// 示例（需要替换为真实公钥）:
	// pemStr := `-----BEGIN PUBLIC KEY-----
	// MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...真实公钥内容...
	// -----END PUBLIC KEY-----`
	//
	// block, _ := pem.Decode([]byte(pemStr))
	// if block == nil {
	//     log.Println("❌ 公钥PEM解析失败")
	//     return nil
	// }
	// pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	// if err != nil {
	//     log.Printf("❌ 公钥解析失败: %v", err)
	//     return nil
	// }
	// return pub.(*rsa.PublicKey)
	
	log.Println("⚠️  警告：未配置嵌入公钥，请在生产环境配置 publicKeyPath")
	return nil
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
