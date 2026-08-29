# API 参考文档

本目录记录 qtcloud-connect 当前可用的接口。v0.1 的重点是打通
`Provider -> CLI -> Studio` 的共识追溯闭环。

## 文档结构

| 文档 | 说明 |
|------|------|
| [Provider API](provider.md) | Go Provider 的消息和共识 API |

## Provider 概览

- **基础 URL**：`http://localhost:8000/api`
- **实现**：Go + SQLite
- **认证**：v0.1 Provider 尚未内置认证；公网部署必须放在认证、HTTPS、限流网关之后
- **CORS**：默认允许 `https://studio.connect.cloud.quanttide.com` 和本地开发端口

### 消息管理

```http
GET /api/messages
GET /api/messages/{id}
```

### 共识管理

```http
POST /api/consensuses
GET  /api/consensuses?page=1&page_size=20
GET  /api/consensuses/{id}
PUT  /api/consensuses/{id}
POST /api/consensuses/confirm
POST /api/consensuses/deprecate
```

### 健康检查

```http
GET /healthz
```

## CLI 示例

```bash
qtcloud-connect consensus create --title "共识标题" --description "共识描述"
qtcloud-connect consensus list
qtcloud-connect consensus show <consensus-id>
qtcloud-connect consensus update <consensus-id> --title "新标题" --description "新描述"
qtcloud-connect consensus confirm <consensus-id>
qtcloud-connect consensus deprecate <consensus-id>
```

`--endpoint` 可覆盖默认 Provider 地址 `http://localhost:8000/api`。

## 后续 API 路线

- v0.2：建立 Message -> Consensus -> Memo 共识加工管道。
- v0.3：把共识生产线接入 qtdata 等业务系统。
