# server-active - 许可证授权服务

## 概述

`server-active` 是一个基于 Go 语言开发的许可证授权服务，用于生成、校验和管理软件许可证。

## 主要功能

### 1. 许可证生成
- 生成唯一许可证密钥 (UUID)
- RSA 2048 位数字签名
- Base64 编码
- MySQL 持久化存储

### 2. 许可证校验
- RSA 签名验证
- 有效期检查
- 硬件指纹绑定
- 激活次数限制
- 在线/离线校验

### 3. 许可证管理
- 许可证转移（更换设备）
- 许可证撤销
- 查询许可证信息
- 激活历史记录

## 技术栈

- **语言**: Go 1.21+
- **数据库**: MySQL 8.0+
- **框架**: Gin (HTTP Router)
- **加密**: RSA 2048 + SHA-256
- **依赖**:
  - `github.com/gin-gonic/gin` - HTTP 框架
  - `github.com/go-sql-driver/mysql` - MySQL 驱动
  - `github.com/google/uuid` - UUID 生成
  - `github.com/beevik/ntp` - NTP 时间同步

## 快速开始

### 1. 生成 RSA 密钥对

```bash
# 创建 keys 目录
mkdir keys

# 生成私钥
openssl genrsa -out keys/private.pem 2048

# 生成公钥
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

### 2. 配置数据库

创建 `config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": "8080"
  },
  "database": {
    "host": "localhost",
    "port": "3306",
    "user": "root",
    "password": "your_password",
    "database": "dy_license"
  },
  "license": {
    "private_key_path": "./keys/private.pem",
    "public_key_path": "./keys/public.pem"
  }
}
```

### 3. 初始化数据库

```bash
mysql -u root -p
CREATE DATABASE dy_license CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. 编译运行

**Windows:**
```bash
build.bat
dy-live-license-server.exe
```

**Linux/Mac:**
```bash
go build -o dy-live-license-server .
./dy-live-license-server
```

## API 接口

### 1. 生成许可证

**请求:**
```http
POST /api/v1/licenses/generate
Content-Type: application/json

{
  "customer_id": "customer_001",
  "software_id": "dy-live-monitor",
  "expiry_days": 365,
  "max_activations": 1,
  "license_type": "full",
  "features": {
    "max_rooms": 10,
    "export_data": true
  }
}
```

**响应:**
```json
{
  "license_key": "550e8400-e29b-41d4-a716-446655440000",
  "license_data": "base64_encoded_license_string",
  "expiry_date": "2026-11-15T00:00:00Z"
}
```

### 2. 校验许可证

**请求:**
```http
POST /api/v1/licenses/validate
Content-Type: application/json

{
  "license_key": "550e8400-e29b-41d4-a716-446655440000",
  "hardware_fingerprint": "sha256_hardware_fingerprint",
  "ip_address": "192.168.1.100",
  "device_info": "Windows 10 Pro"
}
```

**响应（成功）:**
```json
{
  "valid": true,
  "message": "Validation successful",
  "expiry_date": "2026-11-15T00:00:00Z",
  "remaining_days": 365,
  "license_data": "base64_encoded_license_string"
}
```

**响应（失败）:**
```json
{
  "valid": false,
  "message": "License has expired"
}
```

### 3. 转移许可证

**请求:**
```http
POST /api/v1/licenses/transfer
Content-Type: application/json

{
  "license_key": "550e8400-e29b-41d4-a716-446655440000",
  "old_fingerprint": "old_device_fingerprint",
  "new_fingerprint": "new_device_fingerprint"
}
```

**响应:**
```json
{
  "success": true,
  "message": "License transferred successfully"
}
```

### 4. 查询许可证信息

**请求:**
```http
GET /api/v1/licenses/550e8400-e29b-41d4-a716-446655440000
```

**响应:**
```json
{
  "license_key": "550e8400-e29b-41d4-a716-446655440000",
  "software_id": "dy-live-monitor",
  "customer_id": "customer_001",
  "expiry_date": "2026-11-15T00:00:00Z",
  "activation_count": 1,
  "max_activations": 1,
  "license_type": "full",
  "status": "active"
}
```

### 5. 撤销许可证

**请求:**
```http
POST /api/v1/licenses/550e8400-e29b-41d4-a716-446655440000/revoke
```

**响应:**
```json
{
  "success": true,
  "message": "License revoked successfully"
}
```

## 数据库表结构

### licenses (许可证主表)

| 字段 | 类型 | 说明 |
|------|------|------|
| license_id | INT | 主键 |
| license_key | VARCHAR(255) | 许可证密钥 (UUID) |
| software_id | VARCHAR(100) | 软件产品ID |
| customer_id | VARCHAR(100) | 客户ID |
| expiry_date | TIMESTAMP | 过期时间 |
| hardware_fingerprint | VARCHAR(255) | 硬件指纹 |
| activation_count | INT | 当前激活次数 |
| max_activations | INT | 最大激活次数 |
| license_type | ENUM | 许可证类型 (trial/full) |
| features | JSON | 功能权限 |
| status | ENUM | 状态 (active/expired/revoked) |

### activation_records (激活记录表)

| 字段 | 类型 | 说明 |
|------|------|------|
| record_id | INT | 主键 |
| license_id | INT | 外键，关联 licenses 表 |
| hardware_fingerprint | VARCHAR(255) | 硬件指纹 |
| activation_time | TIMESTAMP | 激活时间 |
| ip_address | VARCHAR(45) | IP 地址 |
| device_info | TEXT | 设备信息 |

## 安全性

1. **RSA 2048 位加密**: 私钥签名，公钥验证
2. **SHA-256 哈希**: 数据完整性校验
3. **硬件指纹绑定**: 防止许可证转移
4. **激活次数限制**: 控制设备数量
5. **HTTPS**: API 通信加密（生产环境必须）

## 部署建议

1. **使用 HTTPS**: 所有 API 必须通过 HTTPS 访问
2. **数据库备份**: 定期备份 MySQL 数据库
3. **密钥保护**: RSA 私钥必须安全存储，不能泄露
4. **防火墙**: 限制数据库和 API 访问
5. **日志监控**: 记录所有许可证操作

## License

MIT License
