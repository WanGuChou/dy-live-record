# 文档清理计划

## 📊 当前状态
- 总文件数: 46 个 md 文件
- 总大小: ~350KB

## 🗑️ 需要删除的文件（33 个）

### 1. WebView2 相关（已废弃 - 7 个）
- ❌ WEBVIEW2_FIX.md - Fyne 已替代
- ❌ BUILD_FIX_NOTES.md - 构建问题已解决
- ❌ BUILD_FIX_v2.md - 构建问题已解决
- ❌ BUILD_STATUS.md - 过时的状态
- ❌ MANUAL_BUILD_STEPS.md - 被简化
- ❌ SOLUTION_SUMMARY.md - 已过时
- ❌ FIX_CGO_PATHS.bat 相关问题

### 2. 构建指南（重复 - 3 个）
- ❌ BUILD_INSTRUCTIONS.md - 被 README_FYNE.md 包含
- ❌ QUICK_BUILD.md - 被 README_FYNE.md 包含
- ❌ QUICK_DEBUG.md - 被 DEBUG_MODE.md 替代

### 3. CDP/测试文档（开发过程 - 5 个）
- ❌ CDP_TEST.md - 开发测试文档
- ❌ CDP_SUMMARY.md - 开发总结
- ❌ CDP_USAGE.md - 被主文档包含
- ❌ DETAILED_TEST.md - 开发测试
- ❌ TEST_GUIDE.md - 已包含在主文档
- ❌ DEBUG_GUIDE.md - 被 DEBUG_MODE.md 替代

### 4. Douyin 解析器修复（开发过程 - 5 个）
- ❌ DOUYIN_FIELD_FIX.md
- ❌ DOUYIN_PARSER_FIX.md
- ❌ DOUYIN_PARSER_TECH.md
- ❌ DOUYIN_USER_FIX.md
- ❌ DOUYIN_QUICK_START.md

### 5. 状态报告（开发过程 - 6 个）
- ❌ COMPLETION_REPORT.md - 已完成
- ❌ CURRENT_STATUS.md - 过时
- ❌ FINAL_STATUS.md - 过时
- ❌ IMPLEMENTATION_STATUS.md - 过时
- ❌ PROJECT_SUMMARY.md - 过时
- ❌ FEATURE_SUMMARY.md - 过时

### 6. 旧版本文档（过时 - 7 个）
- ❌ CHANGELOG.md - 被 CHANGELOG_v3.2.0.md 替代
- ❌ UPGRADE_GUIDE.md - 被 UPGRADE_TO_FYNE.md 替代
- ❌ UPGRADE_TO_v1.0.1.md - 旧版本
- ❌ PROJECT_ROADMAP.md - 项目已完成
- ❌ PROJECT_STRUCTURE.md - 被主文档包含
- ❌ NEXT_STEPS.md - 已完成
- ❌ IMPROVEMENTS.md - 已实现
- ❌ RELEASE_NOTES.md - 旧版本

### 7. 重复/整合（可删除 - 5 个）
- ❌ QUICK_START.md - 被 README_FYNE.md 包含
- ❌ QUICK_START_WINDOWS.md - 被 README_FYNE.md 包含
- ❌ USAGE.md - 被主文档包含
- ❌ TROUBLESHOOTING.md - 被 README_ERRORS.md 包含
- ❌ HIDE_DEBUGGER_BANNER.md - 已不相关

## ✅ 保留的文件（13 个）

### 核心文档（6 个）
1. ✅ README.md - 项目主文档
2. ✅ README_FYNE.md - Fyne 版本详细文档
3. ✅ README_ERRORS.md - 错误排查（可选保留）
4. ✅ CHANGELOG_v3.2.0.md - 最新变更日志
5. ✅ FYNE_MIGRATION.md - 技术迁移指南
6. ✅ UPGRADE_TO_FYNE.md - 用户升级指南

### 功能文档（2 个）
7. ✅ DEBUG_MODE.md - 调试模式指南
8. ✅ INSTALL_GUIDE.md - 依赖安装指南

### 子项目文档（5 个）
9. ✅ server-go/proto/README.md - Proto 定义
10. ✅ server-go/README.md - server-go 说明
11. ✅ server-active/README.md - License 服务说明
12. ✅ browser-monitor/README.md - 插件说明
13. ✅ browser-monitor/icons/README.md - 图标说明

## 📈 清理后
- 文件数: 13 个（减少 71%）
- 估计大小: ~100KB（减少 70%）

## 🎯 推荐保留结构

```
/
├── README.md                    # 主文档
├── README_FYNE.md              # Fyne 详细文档
├── CHANGELOG_v3.2.0.md         # 变更日志
├── FYNE_MIGRATION.md           # 技术迁移
├── UPGRADE_TO_FYNE.md          # 升级指南
├── DEBUG_MODE.md               # 调试模式
├── INSTALL_GUIDE.md            # 安装指南
│
├── server-go/
│   ├── README.md
│   └── proto/README.md
│
├── server-active/
│   └── README.md
│
└── browser-monitor/
    ├── README.md
    └── icons/README.md
```

## 🔄 可选择操作

### 方案 A: 全部删除（推荐）
直接删除所有标记为 ❌ 的文件

### 方案 B: 归档
创建 `docs/archive/` 目录，移动而不是删除

### 方案 C: 整合
将有用信息整合到核心文档中，然后删除
