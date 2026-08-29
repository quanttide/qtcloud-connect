# 量潮沟通云文档

## 产品简介

量潮沟通云（qtcloud-connect）是一款沟通管理工具，当前 v0.1 聚焦把团队讨论沉淀为可追溯共识：CLI 手动记录，Provider 持久化，Studio 展示 Message -> Consensus -> Memo 的共识页面。

## 文档结构

本目录包含以下文档分类：

### [用户指南](user-guide/README.md)
面向最终用户的操作文档，帮助用户了解和使用系统功能。
- CLI 记录共识
- Studio 查看共识
- 常见问题

### [开发者指南](dev-guide/README.md)
面向开发人员的技术文档，帮助理解当前 Go Provider、Rust CLI、Flutter Studio 的开发和发布前检查。
- 本地开发命令
- 测试和构建命令
- 发布与部署注意事项

### [API 参考文档](api-references/README.md)
详细的 API 接口文档，供开发人员调用和集成。
- Provider API 接口文档
- 共识 API 接口文档
- CLI 共识命令示例
- Studio 对接方式

## 快速导航

### 按角色导航

#### 业务用户
如果您是使用系统的业务用户，建议从以下文档开始：
1. [用户指南](user-guide/README.md)
2. [用户指南 - 详细内容](user-guide/index.md)
3. [API 参考 - 共识 API](api-references/provider.md#共识-api)

#### 开发人员
如果您是参与开发的工程师，建议从以下文档开始：
1. [开发者指南](dev-guide/README.md)
2. [开发者指南 - 详细内容](dev-guide/index.md)
3. [API 参考文档 - Provider API](api-references/provider.md)

#### 第三方集成者
如果您需要与系统集成，建议从以下文档开始：
1. [API 参考文档 - Provider API](api-references/provider.md)
2. [开发者指南 - CLI 命令](dev-guide/index.md#cli)
3. [开发者指南 - Studio 运行](dev-guide/index.md#studio)

### 按功能导航

#### 共识追溯
- [API 参考 - 共识 API](api-references/provider.md#共识-api)
- [开发者指南 - CLI](dev-guide/index.md#cli)
- [开发者指南 - Studio](dev-guide/index.md#studio)

## 系统要求

### 用户端
- 现代浏览器（Chrome 90+、Firefox 88+、Safari 14+、Edge 90+）
- 稳定的网络连接
- 支持 JavaScript 和 Local Storage

### 开发端
- Go 1.26+
- Rust 1.70+
- Flutter / Dart SDK

## 获取帮助

### 文档问题
如果您发现文档中有错误或需要补充，请：
1. 在 GitHub 仓库中创建 Issue
2. 标签选择 `documentation`
3. 详细描述问题或建议

### 技术支持
如果您在使用过程中遇到问题，请联系技术支持：
- 邮箱：support@quanttide.com
- 内部工单系统：[链接]
- 紧急联系人：[姓名] [电话]

### 社区交流
- 内部论坛：[链接]
- 飞书群组：[群组名称]
- 周会时间：每周三 14:00-15:00

## 文档维护

### 贡献流程
1. Fork 项目仓库
2. 创建文档分支：`git checkout -b docs/your-feature`
3. 提交更改：`git commit -m "docs: add/fix your documentation"`
4. 推送分支：`git push origin docs/your-feature`
5. 创建 Pull Request

### 文档规范
- 使用 Markdown 格式
- 遵循 [量潮科技文档格式章程](../../.quanttide/docs-format.md)
- 中文为主，技术术语可保留英文
- 包含示例代码和截图

### 更新频率
- 用户指南：功能可用性变化后更新
- 开发者指南：组件技术栈、运行方式或发布流程变化后更新
- API 参考：每次 API 变更后更新

## 版本历史

### v1.0.0 (2026-08-28)
- 初始版本发布
- 完成基础文档结构
- 包含用户指南、开发者指南、API 参考

### 后续版本规划
- v1.1.0：添加视频教程和交互式示例
- v1.2.0：多语言支持（英文版本）
- v2.0.0：重构文档结构，优化搜索功能

## 相关资源

### 项目仓库
- 主仓库：[quanttide-connect](https://github.com/quanttide/quanttide-connect)
- CLI 工具：[qtcloud-connect](https://github.com/quanttide/qtcloud-connect)
- 产品档案：[quanttide-profile-of-product-development](https://github.com/quanttide/quanttide-profile-of-product-development)

### 外部文档
- Go 文档：[go.dev/doc](https://go.dev/doc/)
- Flutter 文档：[flutter.dev](https://flutter.dev/)
- Rust 文档：[doc.rust-lang.org](https://doc.rust-lang.org/)

### 设计资源
- Figma 设计稿：[链接]
- 原型演示：[链接]
- 图标库：[链接]

---

*最后更新时间：2026-08-29*
*文档版本：v0.1 draft*
