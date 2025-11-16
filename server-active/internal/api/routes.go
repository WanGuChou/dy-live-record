package api

import (
	"database/sql"
	"dy-live-license/internal/config"
	"dy-live-license/internal/license"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, db *sql.DB, cfg *config.Config) {
	licenseManager := license.NewManager(db, cfg)

	api := router.Group("/api/v1")
	{
		// 许可证生成 (管理后台使用)
		api.POST("/licenses/generate", func(c *gin.Context) {
			GenerateLicense(c, licenseManager)
		})

		// 许可证校验/激活 (客户端调用)
		api.POST("/licenses/validate", func(c *gin.Context) {
			ValidateLicense(c, licenseManager)
		})

		// 许可证转移/解绑 (管理后台使用)
		api.POST("/licenses/transfer", func(c *gin.Context) {
			TransferLicense(c, licenseManager)
		})

		// 查询许可证信息
		api.GET("/licenses/:license_key", func(c *gin.Context) {
			GetLicenseInfo(c, licenseManager)
		})

		// 撤销许可证
		api.POST("/licenses/:license_key/revoke", func(c *gin.Context) {
			RevokeLicense(c, licenseManager)
		})
	}
}
