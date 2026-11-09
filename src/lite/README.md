# lite

量潮沟通云的微信小程序，基于 Taro 4、React 与 TypeScript，构建产物输出到 `dist/`，由微信开发者工具导入运行。定位是**只读展示**——字段口径与平台解耦的约定见下方文档，小程序不写任何东西。demo 数据取创始人圈定的 6 个活跃群（清单经 lark-cli 核对），群主等内部信息不展示。

## 页面

- 首页（沟通渠道卡片集）：一渠道一张卡，展示名称、用途描述与分类（全员群 / 业务群 / 职能群）；数据来自 `src/data/channels.ts`。

## 文档

- [介绍](docs/intro/index.md)：要解决的问题与使用场景。
- [用户指南](docs/user-guide/index.md)：卡片字段，以及渠道不绑定平台、先手动管理的口径。

## 命令

```bash
npm install          # 首次安装依赖
npm run dev:weapp    # 监听构建
npm run build:weapp  # 构建产物到 dist/
```

## 上架准备

1. 在微信公众平台注册小程序并取得 AppID，替换 `project.config.json` 中的 `appid`，当前占位值为 `touristappid`，替换后需重新执行构建。
2. 用微信开发者工具导入本目录，工具按 `project.config.json` 的 `miniprogramRoot` 指向 `dist/` 预览与调试。
3. 在小程序后台配置服务器域名；当前页面纯本地数据，无请求，真机不需合法域名。
4. 选定服务类目并填写《用户隐私保护指引》。
5. 上传体验版供内部验收，再提交审核；审核通过后发布，后续版本沿用同一 AppID 迭代。

## 发布（惯例，CI 待配）

按子包惯例打 tag：`lite/vX.Y.Z`；子包版本见本目录 [`CHANGELOG.md`](CHANGELOG.md)，仓库级变更才进根 CHANGELOG。
