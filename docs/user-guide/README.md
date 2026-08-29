# 用户指南

v0.1 的用户目标是把团队共识记录下来，并在 Studio 里查看共识追溯页面。

## 当前可用能力

- 使用 CLI 创建、查看、更新、确认和废弃共识。
- 使用 Studio 查看 Provider 中的共识列表和详情。
- 通过 Provider 持久化共识标题、描述、状态和时间戳。

## 快速开始

1. 启动 Provider：`cd src/provider && go run cmd/server/main.go`
2. 创建共识：`qtcloud-connect consensus create --title "共识标题" --description "共识描述"`
3. 打开 Studio：`cd src/studio && flutter run -d chrome`
4. 在共识追溯页面查看记录，并点击共识节点查看详情。

## 常用命令

```bash
qtcloud-connect consensus list
qtcloud-connect consensus show <consensus-id>
qtcloud-connect consensus update <consensus-id> --title "新标题" --description "新描述"
qtcloud-connect consensus confirm <consensus-id>
qtcloud-connect consensus deprecate <consensus-id>
```

## 注意事项

- 默认 Provider 地址是 `http://localhost:8000/api`，可通过 `--endpoint` 覆盖。
- Studio 默认读取同一个 Provider 地址，可通过 `--dart-define=CONNECT_PROVIDER_ENDPOINT=...` 覆盖。
- 公网环境的认证、HTTPS 和限流由部署网关提供；v0.1 Provider 不内置登录。

## 后续路线

- v0.2：建立 Message -> Consensus -> Memo 的共识加工管道。
- v0.3：接入 qtdata 等业务系统，加强客户沟通沉淀。
