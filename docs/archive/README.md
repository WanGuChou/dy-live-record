# 文档归档目录

本目录包含项目开发过程中创建的历史文档，这些文档已经过时或被新文档替代。

## 📁 归档原因

这些文档被归档是因为：
- ✅ 已被新文档替代
- ✅ 记录的功能已废弃（如 WebView2）
- ✅ 开发过程中的临时文档
- ✅ 版本更新后不再适用

## 🗂️ 归档分类

### WebView2 相关（已废弃）
- WEBVIEW2_FIX.md
- BUILD_FIX_NOTES.md
- BUILD_FIX_v2.md
- BUILD_STATUS.md
- MANUAL_BUILD_STEPS.md
- SOLUTION_SUMMARY.md

项目已从 WebView2 迁移到 Fyne，这些文档不再适用。

### 构建指南（已整合）
- BUILD_INSTRUCTIONS.md
- QUICK_BUILD.md
- QUICK_DEBUG.md
- QUICK_START.md
- QUICK_START_WINDOWS.md

内容已整合到 `README_FYNE.md` 和 `DEBUG_MODE.md`。

### CDP/测试文档（开发过程）
- CDP_TEST.md
- CDP_SUMMARY.md
- CDP_USAGE.md
- DETAILED_TEST.md
- TEST_GUIDE.md
- DEBUG_GUIDE.md

开发测试过程文档，功能已稳定。

### Douyin 解析器（开发过程）
- DOUYIN_FIELD_FIX.md
- DOUYIN_PARSER_FIX.md
- DOUYIN_PARSER_TECH.md
- DOUYIN_USER_FIX.md
- DOUYIN_QUICK_START.md

解析器修复过程文档，问题已解决。

### 状态报告（已完成）
- COMPLETION_REPORT.md
- CURRENT_STATUS.md
- FINAL_STATUS.md
- IMPLEMENTATION_STATUS.md
- PROJECT_SUMMARY.md
- FEATURE_SUMMARY.md

项目状态报告，功能已全部完成。

### 旧版本文档
- CHANGELOG.md → 被 CHANGELOG_v3.2.0.md 替代
- UPGRADE_GUIDE.md → 被 UPGRADE_TO_FYNE.md 替代
- UPGRADE_TO_v1.0.1.md → 旧版本升级指南
- RELEASE_NOTES.md → 旧版本发布说明

### 项目管理（已完成）
- PROJECT_ROADMAP.md
- PROJECT_STRUCTURE.md
- NEXT_STEPS.md
- IMPROVEMENTS.md

项目规划文档，目标已达成。

### 其他
- USAGE.md → 内容已整合到主文档
- TROUBLESHOOTING.md → 被 README_ERRORS.md 包含
- HIDE_DEBUGGER_BANNER.md → CDP 相关，已不需要

## 📚 当前有效文档

请参考项目根目录的以下文档：

### 核心文档
- `README.md` - 项目主文档
- `README_FYNE.md` - Fyne 版本详细说明
- `CHANGELOG_v3.2.0.md` - 最新变更日志

### 使用指南
- `UPGRADE_TO_FYNE.md` - 升级指南
- `FYNE_MIGRATION.md` - 技术迁移指南
- `DEBUG_MODE.md` - 调试模式使用
- `INSTALL_GUIDE.md` - 依赖安装指南

### 子项目文档
- `server-go/README.md`
- `server-go/proto/README.md`
- `server-active/README.md`
- `browser-monitor/README.md`

## ⚠️ 注意事项

这些归档文档：
- 仅供参考，可能包含过时信息
- 不保证与当前版本兼容
- 建议优先查阅最新文档

## 🗑️ 清理历史

**归档时间**: 2025-11-15  
**归档版本**: v3.2.1  
**归档原因**: Fyne 迁移完成，清理冗余文档

如需恢复任何文档，请从 Git 历史记录中恢复。
