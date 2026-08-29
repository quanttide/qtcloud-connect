# qtcloud-connect Studio

Flutter 工作台，当前默认入口是共识追溯页面。

共识追溯页面不是固定示例图，而是从 Provider 加载 `ConsensusGraph`。工具栏可添加和编辑共识、纳入已有共识、建立自定义类型关联；画布支持缩放、平移和节点拖动，右侧面板展示节点详情以及入边/出边。

## 本地开发

```bash
flutter pub get
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

`CONNECT_PROVIDER_ENDPOINT` 默认值为 `http://localhost:8000/api`。生产环境部署到
`https://studio.connect.cloud.quanttide.com` 时，需要确保 Provider 的 CORS 白名单包含该来源。Studio 是静态 Web 产物，不能将 `CONNECT_AUTH_TOKEN` 编入浏览器；公网用户级认证和 Provider 服务令牌应由网关或反向代理处理。

## 验证

```bash
flutter analyze
flutter test
flutter build web --release --base-href /
```
