package api

import (
	"dy-live-license/internal/license"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateLicenseRequest 生成许可证请求
type GenerateLicenseRequest struct {
	CustomerID     string                 `json:"customer_id"`
	SoftwareID     string                 `json:"software_id"`
	ExpiryDays     int                    `json:"expiry_days"`
	MaxActivations int                    `json:"max_activations"`
	LicenseType    string                 `json:"license_type"`
	Features       map[string]interface{} `json:"features"`
}

// ValidateLicenseRequest 校验许可证请求
type ValidateLicenseRequest struct {
	LicenseKey           string `json:"license_key"`
	HardwareFingerprint  string `json:"hardware_fingerprint"`
	IPAddress            string `json:"ip_address"`
	DeviceInfo           string `json:"device_info"`
}

// TransferLicenseRequest 转移许可证请求
type TransferLicenseRequest struct {
	LicenseKey         string `json:"license_key"`
	OldFingerprint     string `json:"old_fingerprint"`
	NewFingerprint     string `json:"new_fingerprint"`
}

// GenerateLicense 生成许可证
func GenerateLicense(c *gin.Context, manager *license.Manager) {
	var req GenerateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// 计算过期时间
	expiryDate := time.Now().AddDate(0, 0, req.ExpiryDays)

	// 生成许可证
	licenseData, err := manager.GenerateLicense(
		req.CustomerID,
		req.SoftwareID,
		expiryDate,
		req.MaxActivations,
		req.LicenseType,
		req.Features,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate license", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"license_key":  licenseData.LicenseKey,
		"license_data": licenseData.EncodedData,
		"expiry_date":  licenseData.ExpiryDate,
	})
}

// ValidateLicense 校验许可证
func ValidateLicense(c *gin.Context, manager *license.Manager) {
	var req ValidateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// 校验许可证
	valid, message, licenseInfo, err := manager.ValidateLicense(
		req.LicenseKey,
		req.HardwareFingerprint,
		req.IPAddress,
		req.DeviceInfo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": message,
		})
		return
	}

	// 计算剩余天数
	remainingDays := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)

	c.JSON(http.StatusOK, gin.H{
		"valid":          true,
		"message":        message,
		"expiry_date":    licenseInfo.ExpiryDate,
		"remaining_days": remainingDays,
		"license_data":   licenseInfo.EncodedData,
	})
}

// TransferLicense 转移许可证
func TransferLicense(c *gin.Context, manager *license.Manager) {
	var req TransferLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// 转移许可证
	if err := manager.TransferLicense(req.LicenseKey, req.OldFingerprint, req.NewFingerprint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License transferred successfully",
	})
}

// GetLicenseInfo 获取许可证信息
func GetLicenseInfo(c *gin.Context, manager *license.Manager) {
	licenseKey := c.Param("license_key")

	info, err := manager.GetLicenseInfo(licenseKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "License not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// RevokeLicense 撤销许可证
func RevokeLicense(c *gin.Context, manager *license.Manager) {
	licenseKey := c.Param("license_key")

	if err := manager.RevokeLicense(licenseKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Revoke failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License revoked successfully",
	})
}
