# 用户指南详细内容

## 使用流程

### 1. 记录共识

团队形成阶段性结论后，用 CLI 写入 Provider：

```bash
qtcloud-connect consensus create \
  --title "Studio 优先上线共识页面" \
  --description "v0.1 先打通 CLI 写入、Provider 持久化、Studio 展示。"
```

命令会返回共识 ID、状态和时间戳。后续更新、确认、废弃都使用该 ID。

### 2. 查看共识

```bash
qtcloud-connect consensus list
qtcloud-connect consensus show <consensus-id>
```

Studio 默认打开“共识追溯图”，会显示 Provider 返回的共识图。点击节点可以查看标题、描述、状态、创建时间和更新时间；也可以新增共识、纳入已有共识、编辑提议中的共识，或为两个节点建立自定义类型的关联。

### 3. 更新状态

```bash
qtcloud-connect consensus update <consensus-id> --title "新标题" --description "新描述"
qtcloud-connect consensus confirm <consensus-id>
qtcloud-connect consensus deprecate <consensus-id>
```

状态含义：

| 状态 | 含义 |
|------|------|
| `proposed` | 提议中，尚未确认 |
| `confirmed` | 已确认，可作为当前团队共识 |
| `deprecated` | 已废弃，不再作为当前共识 |

## 常见问题

### Provider 未连接

确认 Provider 正在运行：

```bash
cd src/provider
go run cmd/server/main.go
```

如果端口不是 `8000`，CLI 增加 `--endpoint`，Studio 增加 `--dart-define=CONNECT_PROVIDER_ENDPOINT=...`。

### Studio 看不到新增记录

点击 Studio 右上角刷新按钮；如果仍不可见，先用 `qtcloud-connect consensus list` 确认 Provider 中是否存在记录。记录存在但不在图上时，使用“纳入已有共识”把它加入当前图。

### 公网部署是否可以直接开放 Provider

不可以。Provider 可用 `CONNECT_AUTH_TOKEN` 做服务级 Bearer 校验，但公网环境仍必须放在用户级认证、HTTPS、限流网关之后。

## 后续路线

- v0.2：从 Message 加工 Consensus，再生成 Memo。
- v0.3：接入 qtdata 等业务系统，沉淀客户沟通上下文。
