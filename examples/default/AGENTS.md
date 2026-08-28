# AGENTS.md

## 子模块仓库初始化规则

当需要新建子模块仓库并将其纳入 `quanttide-connect` 时，遵循以下步骤：

### 命名约定

| 子目录 | GitHub 仓库名 | 命名模式 |
|--------|--------------|----------|
| `data/` | `quanttide-{noun}-of-communication-management` | `quanttide-report-of-communication-management`, `quanttide-intention-of-communication-management`, `quanttide-journal-of-communication-management` |
| `docs/` | `quanttide-tutorial-of-communication-management` |  |
| `apps/` | 独立项目名 | 如 `qtcloud-connect` |

### 初始化步骤

1. **创建 GitHub 仓库**

   ```
   gh repo create quanttide/quanttide-{name} --public --add-readme --license cc-by-4.0
   ```

   - 可见性：`--public`
   - 许可证：`--license cc-by-4.0`（CC BY 4.0）
   - README：`--add-readme`（自动生成）

2. **添加为子模块**

   ```
   git submodule add https://github.com/quanttide/quanttide-{name}.git {path}
   ```

3. **提交**

   ```
   git add .gitmodules {path}
   git commit -m "feat: add {name} as {path} submodule"
   ```

### 目录约定

- `data/` — 数据类项目（报告、意图、日志等结构化数据）
- `docs/` — 文档类项目（教程、指南等）
- `apps/` — 可部署应用
- `packages/` — 共享库/工具包
- `examples/` — 示例项目
