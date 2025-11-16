# 项目文档结构

> **最后更新**: 2025-11-15  
> **文档版本**: v3.2.1  
> **文档总数**: 19 个（已清理 39 个过时文档）

---

## 📚 当前文档结构

```
dy-live-record/
├── 📖 核心文档（8 个）
│   ├── README.md                    # 项目主文档
│   ├── README_FYNE.md              # Fyne 版本详细文档
│   ├── README_ERRORS.md            # 错误排查指南
│   ├── CHANGELOG_v3.2.0.md         # v3.2.0 变更日志
│   ├── FYNE_MIGRATION.md           # WebView2 → Fyne 迁移指南
│   ├── UPGRADE_TO_FYNE.md          # 用户升级指南
│   ├── DEBUG_MODE.md               # 调试模式使用文档
│   └── INSTALL_GUIDE.md            # 依赖安装指南
│
├── 📁 子项目文档（8 个）
│   ├── server-go/
│   │   ├── README.md               # server-go 说明
│   │   └── proto/README.md         # Protobuf 定义文档
│   │
│   ├── server-active/
│   │   └── README.md               # License 服务说明
│   │
│   ├── server/                     # (Legacy Node.js 版本)
│   │   ├── README.md
│   │   ├── README_DOUYIN.md
│   │   └── README_LOGS.md
│   │
│   └── browser-monitor/
│       ├── README.md               # 插件说明
│       ├── QUICKSTART.md
│       ├── README_STARTUP_SCRIPTS.md
│       └── icons/README.md
│
├── 📋 管理文档（3 个）
│   ├── DOCS_CLEANUP.md             # 文档清理计划
│   └── DOCUMENTATION_STRUCTURE.md  # 本文档
│
└── 📦 归档（40 个）
    └── docs/archive/               # 过时文档归档
        └── README.md               # 归档说明

```

---

## 📖 文档分类详解

### 1. 核心用户文档

#### 🌟 README.md
**用途**: 项目总览和快速开始  
**受众**: 所有用户  
**内容**:
- 项目简介
- 快速开始
- 系统要求
- 核心功能
- 许可证信息

**何时阅读**: 首次接触项目

---

#### 🎨 README_FYNE.md
**用途**: Fyne GUI 版本完整文档  
**受众**: 使用 Fyne 版本的用户  
**内容**:
- Fyne 版本特点
- 详细安装步骤
- UI 功能说明
- 性能对比
- 常见问题

**何时阅读**: 使用 Fyne 版本时

---

#### 🐛 README_ERRORS.md
**用途**: 错误排查指南  
**受众**: 遇到问题的用户  
**内容**:
- 常见错误及解决方案
- 编译问题
- 运行时错误
- 调试技巧

**何时阅读**: 遇到错误时

---

### 2. 版本更新文档

#### 📝 CHANGELOG_v3.2.0.md
**用途**: v3.2.0 版本变更日志  
**受众**: 关注版本变化的用户  
**内容**:
- 新功能
- 改进项
- Bug 修复
- 性能提升
- 破坏性变更

**何时阅读**: 升级前或了解新版本时

---

#### 🔄 UPGRADE_TO_FYNE.md
**用途**: 从 WebView2 升级到 Fyne 的指南  
**受众**: 从旧版本升级的用户  
**内容**:
- 5 分钟快速升级
- 升级前后对比
- 数据兼容性
- 常见问题
- 回滚方法

**何时阅读**: 从 v3.1.x 升级到 v3.2.0 时

---

### 3. 技术文档

#### 🔬 FYNE_MIGRATION.md
**用途**: 技术迁移详细指南  
**受众**: 开发者、技术人员  
**内容**:
- 迁移前后对比
- 技术架构变更
- 性能测试数据
- 已知问题
- 未来计划

**何时阅读**: 了解技术细节时

---

#### 🐛 DEBUG_MODE.md
**用途**: 调试模式使用指南  
**受众**: 开发者、测试人员  
**内容**:
- 调试模式说明
- 配置选项
- 使用场景
- 安全注意事项
- 最佳实践

**何时阅读**: 本地开发测试时

---

#### 🔧 INSTALL_GUIDE.md
**用途**: 依赖安装详细指南  
**受众**: 首次安装的用户  
**内容**:
- Windows/Linux/macOS 安装
- MinGW-w64 安装
- Go 环境配置
- 常见安装问题

**何时阅读**: 首次编译前

---

### 4. 子项目文档

#### 🖥️ server-go/README.md
**用途**: Go 后端服务说明  
**受众**: 后端开发者  
**内容**:
- 架构设计
- 模块说明
- API 文档
- 配置说明

---

#### 📡 server-go/proto/README.md
**用途**: Protobuf 消息定义文档  
**受众**: 协议开发者  
**内容**:
- 消息类型定义
- 字段编号对照
- 使用说明
- 参考资料

---

#### 🔐 server-active/README.md
**用途**: License 授权服务说明  
**受众**: 系统管理员  
**内容**:
- 服务架构
- API 接口
- 数据库设计
- 部署说明

---

#### 🔌 browser-monitor/README.md
**用途**: 浏览器插件说明  
**受众**: 插件开发者、用户  
**内容**:
- 插件功能
- 安装方法
- 配置说明
- 开发指南

---

### 5. 管理文档

#### 📋 DOCS_CLEANUP.md
**用途**: 文档清理计划说明  
**受众**: 项目维护者  
**内容**:
- 清理原因
- 归档文件列表
- 保留文档说明
- 清理前后对比

---

#### 📚 DOCUMENTATION_STRUCTURE.md (本文档)
**用途**: 文档结构总览  
**受众**: 所有人  
**内容**:
- 文档目录
- 分类说明
- 阅读顺序
- 快速导航

---

## 🗺️ 文档阅读路径

### 新用户路径

```
1. README.md
   ↓
2. README_FYNE.md
   ↓
3. INSTALL_GUIDE.md
   ↓
4. 开始使用
```

### 升级用户路径

```
1. CHANGELOG_v3.2.0.md
   ↓
2. UPGRADE_TO_FYNE.md
   ↓
3. 执行升级
   ↓
4. README_FYNE.md (如需了解新功能)
```

### 开发者路径

```
1. README.md
   ↓
2. FYNE_MIGRATION.md
   ↓
3. DEBUG_MODE.md
   ↓
4. server-go/README.md
   ↓
5. server-go/proto/README.md
```

### 问题排查路径

```
1. README_ERRORS.md
   ↓
2. DEBUG_MODE.md
   ↓
3. INSTALL_GUIDE.md
   ↓
4. GitHub Issues
```

---

## 📊 文档统计

### 当前文档
- **总数**: 19 个
- **核心文档**: 8 个
- **子项目文档**: 8 个
- **管理文档**: 3 个

### 归档文档
- **总数**: 40 个
- **WebView2 相关**: 6 个
- **构建指南**: 5 个
- **测试文档**: 6 个
- **解析器修复**: 5 个
- **状态报告**: 6 个
- **旧版本**: 8 个
- **其他**: 4 个

### 清理效果
- **文档减少**: 71% (从 59 个到 19 个)
- **大小减少**: 约 70% (从 ~350KB 到 ~100KB)

---

## 🎯 文档维护原则

### ✅ 保留条件
1. 当前版本必需
2. 用户常用
3. 包含独特信息
4. 定期更新

### ❌ 归档条件
1. 功能已废弃
2. 被新文档替代
3. 开发过程文档
4. 版本特定且已过时

### 🔄 更新频率
- **README.md**: 主要版本更新
- **CHANGELOG**: 每个版本
- **技术文档**: 架构变更时
- **子项目文档**: 功能变更时

---

## 📞 文档贡献

### 提交新文档
1. 确保文档有明确用途
2. 避免与现有文档重复
3. 使用统一的格式
4. 添加到本文档的目录中

### 更新现有文档
1. 保持文档同步更新
2. 注明更新时间
3. 保留重要的历史信息
4. 使用清晰的版本标注

### 归档文档
1. 确认文档确实过时
2. 移动到 `docs/archive/`
3. 在归档 README 中说明
4. 从主文档中移除引用

---

## 🔗 快速链接

### 核心文档
- [项目主页](README.md)
- [Fyne 版本文档](README_FYNE.md)
- [错误排查](README_ERRORS.md)
- [变更日志](CHANGELOG_v3.2.0.md)

### 安装指南
- [依赖安装](INSTALL_GUIDE.md)
- [快速开始](README.md#快速开始)
- [升级指南](UPGRADE_TO_FYNE.md)

### 开发文档
- [调试模式](DEBUG_MODE.md)
- [技术迁移](FYNE_MIGRATION.md)
- [Proto 定义](server-go/proto/README.md)

### 子项目
- [server-go](server-go/README.md)
- [server-active](server-active/README.md)
- [browser-monitor](browser-monitor/README.md)

---

## 📝 文档更新日志

### v3.2.1 (2025-11-15)
- 清理 39 个过时文档
- 创建归档目录
- 添加文档结构说明
- 优化文档组织

### v3.2.0 (2025-11-15)
- 添加 Fyne 相关文档
- 更新主文档
- 添加调试模式文档

---

**维护者**: Cursor AI Assistant  
**联系方式**: GitHub Issues  
**最后审核**: 2025-11-15
