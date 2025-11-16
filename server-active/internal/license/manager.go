package license

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"dy-live-license/internal/config"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// Manager 许可证管理器
type Manager struct {
	db         *sql.DB
	cfg        *config.Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// LicenseData 许可证数据
type LicenseData struct {
	LicenseKey           string                 `json:"license_key"`
	SoftwareID           string                 `json:"software_id"`
	CustomerID           string                 `json:"customer_id"`
	ExpiryDate           time.Time              `json:"expiry_date"`
	IssueDate            time.Time              `json:"issue_date"`
	HardwareFingerprint  string                 `json:"hardware_fingerprint"`
	ActivationTime       time.Time              `json:"activation_time"`
	MaxActivations       int                    `json:"max_activations"`
	CurrentActivations   int                    `json:"current_activations"`
	LicenseType          string                 `json:"license_type"`
	Features             map[string]interface{} `json:"features"`
	EncodedData          string                 `json:"-"`
}

// NewManager 创建许可证管理器
func NewManager(db *sql.DB, cfg *config.Config) *Manager {
	manager := &Manager{
		db:  db,
		cfg: cfg,
	}

	// 加载 RSA 密钥
	if err := manager.loadKeys(); err != nil {
		panic(fmt.Sprintf("Failed to load RSA keys: %v", err))
	}

	return manager
}

// loadKeys 加载 RSA 密钥
func (m *Manager) loadKeys() error {
	// 加载私钥
	privateKeyData, err := os.ReadFile(m.cfg.License.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil {
		return errors.New("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	m.privateKey = privateKey

	// 加载公钥
	publicKeyData, err := os.ReadFile(m.cfg.License.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	publicKeyBlock, _ := pem.Decode(publicKeyData)
	if publicKeyBlock == nil {
		return errors.New("failed to decode public key PEM")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}
	m.publicKey = publicKey

	return nil
}

// GenerateLicense 生成许可证
func (m *Manager) GenerateLicense(
	customerID, softwareID string,
	expiryDate time.Time,
	maxActivations int,
	licenseType string,
	features map[string]interface{},
) (*LicenseData, error) {
	// 生成唯一许可证密钥
	licenseKey := uuid.New().String()

	// 构造许可证数据
	licenseData := &LicenseData{
		LicenseKey:         licenseKey,
		SoftwareID:         softwareID,
		CustomerID:         customerID,
		ExpiryDate:         expiryDate,
		IssueDate:          time.Now(),
		MaxActivations:     maxActivations,
		CurrentActivations: 0,
		LicenseType:        licenseType,
		Features:           features,
	}

	// 序列化
	dataJSON, err := json.Marshal(licenseData)
	if err != nil {
		return nil, err
	}

	// 使用 RSA 私钥签名
	hashed := sha256.Sum256(dataJSON)
	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	// 组合数据和签名，并 Base64 编码
	combined := append(dataJSON, signature...)
	encodedData := base64.StdEncoding.EncodeToString(combined)
	licenseData.EncodedData = encodedData

	// 存入数据库
	featuresJSON, _ := json.Marshal(features)
	_, err = m.db.Exec(`
		INSERT INTO licenses (license_key, software_id, customer_id, expiry_date, max_activations, license_type, features)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, licenseKey, softwareID, customerID, expiryDate, maxActivations, licenseType, featuresJSON)
	if err != nil {
		return nil, err
	}

	return licenseData, nil
}

// ValidateLicense 校验许可证
func (m *Manager) ValidateLicense(
	licenseKey, hardwareFingerprint, ipAddress, deviceInfo string,
) (bool, string, *LicenseData, error) {
	// 从数据库查询许可证
	var licenseID int
	var softwareID, customerID, status, licenseType string
	var expiryDate time.Time
	var activationCount, maxActivations int
	var dbFingerprint sql.NullString

	err := m.db.QueryRow(`
		SELECT license_id, software_id, customer_id, expiry_date, hardware_fingerprint,
		       activation_count, max_activations, license_type, status
		FROM licenses
		WHERE license_key = ?
	`, licenseKey).Scan(
		&licenseID, &softwareID, &customerID, &expiryDate, &dbFingerprint,
		&activationCount, &maxActivations, &licenseType, &status,
	)

	if err == sql.ErrNoRows {
		return false, "License not found", nil, nil
	}
	if err != nil {
		return false, "Database error", nil, err
	}

	// 检查状态
	if status != "active" {
		return false, "License is not active (revoked or expired)", nil, nil
	}

	// 检查有效期
	if time.Now().After(expiryDate) {
		return false, "License has expired", nil, nil
	}

	// 检查硬件指纹
	if dbFingerprint.Valid && dbFingerprint.String != "" {
		if dbFingerprint.String != hardwareFingerprint {
			return false, "Hardware fingerprint mismatch", nil, nil
		}
	} else {
		// 首次激活，绑定硬件指纹
		_, err := m.db.Exec(`
			UPDATE licenses
			SET hardware_fingerprint = ?, activation_count = activation_count + 1
			WHERE license_id = ?
		`, hardwareFingerprint, licenseID)
		if err != nil {
			return false, "Failed to bind hardware", nil, err
		}

		// 记录激活
		_, err = m.db.Exec(`
			INSERT INTO activation_records (license_id, hardware_fingerprint, ip_address, device_info)
			VALUES (?, ?, ?, ?)
		`, licenseID, hardwareFingerprint, ipAddress, deviceInfo)
		if err != nil {
			return false, "Failed to record activation", nil, err
		}

		activationCount++
	}

	// 检查激活次数
	if activationCount > maxActivations {
		return false, "Activation limit exceeded", nil, nil
	}

	// 返回许可证数据
	licenseData := &LicenseData{
		LicenseKey:           licenseKey,
		SoftwareID:           softwareID,
		CustomerID:           customerID,
		ExpiryDate:           expiryDate,
		HardwareFingerprint:  hardwareFingerprint,
		MaxActivations:       maxActivations,
		CurrentActivations:   activationCount,
		LicenseType:          licenseType,
	}

	return true, "Validation successful", licenseData, nil
}

// TransferLicense 转移许可证
func (m *Manager) TransferLicense(licenseKey, oldFingerprint, newFingerprint string) error {
	// 验证旧指纹
	var dbFingerprint sql.NullString
	err := m.db.QueryRow(`
		SELECT hardware_fingerprint FROM licenses WHERE license_key = ?
	`, licenseKey).Scan(&dbFingerprint)

	if err == sql.ErrNoRows {
		return errors.New("license not found")
	}
	if err != nil {
		return err
	}

	if !dbFingerprint.Valid || dbFingerprint.String != oldFingerprint {
		return errors.New("old fingerprint mismatch")
	}

	// 更新指纹
	_, err = m.db.Exec(`
		UPDATE licenses SET hardware_fingerprint = ? WHERE license_key = ?
	`, newFingerprint, licenseKey)

	return err
}

// GetLicenseInfo 获取许可证信息
func (m *Manager) GetLicenseInfo(licenseKey string) (map[string]interface{}, error) {
	var softwareID, customerID, status, licenseType string
	var expiryDate time.Time
	var activationCount, maxActivations int

	err := m.db.QueryRow(`
		SELECT software_id, customer_id, expiry_date, activation_count, max_activations, license_type, status
		FROM licenses
		WHERE license_key = ?
	`, licenseKey).Scan(&softwareID, &customerID, &expiryDate, &activationCount, &maxActivations, &licenseType, &status)

	if err == sql.ErrNoRows {
		return nil, errors.New("license not found")
	}
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"license_key":       licenseKey,
		"software_id":       softwareID,
		"customer_id":       customerID,
		"expiry_date":       expiryDate,
		"activation_count":  activationCount,
		"max_activations":   maxActivations,
		"license_type":      licenseType,
		"status":            status,
	}, nil
}

// RevokeLicense 撤销许可证
func (m *Manager) RevokeLicense(licenseKey string) error {
	_, err := m.db.Exec(`
		UPDATE licenses SET status = 'revoked' WHERE license_key = ?
	`, licenseKey)
	return err
}

// ListAllLicenses 获取所有许可证列表
func (m *Manager) ListAllLicenses() ([]map[string]interface{}, error) {
	rows, err := m.db.Query(`
		SELECT license_key, software_id, customer_id, expiry_date, 
		       activation_count, max_activations, license_type, status, created_at
		FROM licenses
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	licenses := make([]map[string]interface{}, 0)
	for rows.Next() {
		var licenseKey, softwareID, customerID, licenseType, status string
		var expiryDate, createdAt time.Time
		var activationCount, maxActivations int

		err := rows.Scan(
			&licenseKey, &softwareID, &customerID, &expiryDate,
			&activationCount, &maxActivations, &licenseType, &status, &createdAt,
		)
		if err != nil {
			continue
		}

		licenses = append(licenses, map[string]interface{}{
			"license_key":      licenseKey,
			"software_id":      softwareID,
			"customer_id":      customerID,
			"expiry_date":      expiryDate,
			"activation_count": activationCount,
			"max_activations":  maxActivations,
			"license_type":     licenseType,
			"status":           status,
			"created_at":       createdAt,
		})
	}

	return licenses, nil
}
