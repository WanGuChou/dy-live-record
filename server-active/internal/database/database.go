package database

import (
	"database/sql"
	"dy-live-license/internal/config"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// InitDB 初始化数据库
func InitDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

// createTables 创建数据库表
func createTables(db *sql.DB) error {
	schemas := []string{
		// 许可证主表
		`CREATE TABLE IF NOT EXISTS licenses (
			license_id INT AUTO_INCREMENT PRIMARY KEY,
			license_key VARCHAR(255) NOT NULL UNIQUE COMMENT '许可证唯一Key',
			software_id VARCHAR(100) NOT NULL COMMENT '软件产品ID',
			customer_id VARCHAR(100) COMMENT '客户ID',
			expiry_date TIMESTAMP NULL COMMENT '过期时间',
			hardware_fingerprint VARCHAR(255) NULL COMMENT '绑定的硬件指纹',
			activation_count INT DEFAULT 0 NOT NULL COMMENT '当前激活次数',
			max_activations INT DEFAULT 1 NOT NULL COMMENT '最大激活次数',
			license_type ENUM('trial', 'full') DEFAULT 'full' COMMENT '许可证类型',
			features JSON NULL COMMENT '功能权限JSON',
			status ENUM('active', 'expired', 'revoked') DEFAULT 'active' NOT NULL COMMENT '许可证状态',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			KEY idx_software (software_id),
			KEY idx_customer (customer_id)
		) COMMENT '软件许可证主表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 激活记录表
		`CREATE TABLE IF NOT EXISTS activation_records (
			record_id INT AUTO_INCREMENT PRIMARY KEY,
			license_id INT NOT NULL COMMENT '外键, 关联licenses表',
			hardware_fingerprint VARCHAR(255) NOT NULL COMMENT '激活设备的硬件指纹',
			activation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '激活时间',
			ip_address VARCHAR(45) COMMENT '激活时的IP地址',
			device_info TEXT COMMENT '设备其他信息',
			FOREIGN KEY (license_id) REFERENCES licenses(license_id)
		) COMMENT '许可证激活历史记录表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	}

	return nil
}
