# qtcloud-connect Studio

Flutter 工作台，当前默认入口是共识追溯页面。

## 本地开发

```bash
flutter pub get
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

`CONNECT_PROVIDER_ENDPOINT` 默认值为 `http://localhost:8000/api`。生产环境部署到
`https://studio.connect.cloud.quanttide.com` 时，需要确保 Provider 的 CORS 白名单包含该来源。

## 验证

```bash
flutter analyze
flutter test
flutter build web --release --base-href /
```
