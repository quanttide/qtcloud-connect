# 开发者指南

## 平台定位

qtcloud-connect 是沟通域的共识平台：把团队沟通沉淀为可追溯、可确认、可产出的共识。目标形态是不直接对接飞书：平台适配在 qtcloud-feishu，连接从适配层取镜像与行数据。v0.2 起按这条边界重组。

判断一条能力属于谁，只问一句：它认不认飞书的字段、命令与身份。认，归适配层；不认，归这里。

这样分层的目的有三个：换平台时领域模型不重建；平台原始数据留在镜像里可回溯；共识与产物成为可长期使用的资产。

## 平台边界

领域层持有 Message、Consensus、Relation、ConsensusGraph 与将来的 Memo，负责概念、状态、规则、API 与界面，不感知平台凭据与返回结构。需要写回平台时，领域层只产出意图与内容，由适配层执行，人工确认后再发。

适配层持有动作登记、定义、流程、原样镜像与游标，负责把平台响应拍平；不做岗位分类、话术与共识判断。

业务层持有岗位规则、评审口径与报告文案；qtrecurit 与 qtdata 属于这一层，消费领域层与适配层的数据。

依据见[平台意图](../../../../data/intention/index.md)：适配层做厚，领域层做薄，流程写在数据里。

## 方向调整

v0.1 的 CLI 同时承担共识操作与飞书发送：`mail`、`notice` 直接调 lark-cli，邮件拉取、报告与知识提取也在库里自建。这带来三个问题：与 qtcloud-feishu 重复，把平台细节带进领域应用的发布物，命令面随平台变化。

v0.2 起做四件事：

1. 飞书能力退出（库与示例先退，命令面按替代入口就绪情况退出），映射与步骤见[飞书功能废弃与移除](feishu-removal.md)。
2. 输入改为适配层镜像：消息先落 `raw/`，再按摄取契约进入 Provider，connect 不保存平台凭据。
3. CLI 收敛为共识操作（摄取若放 CLI，只做批处理），Provider 补齐消息写入与溯源接口。
4. 主线转到 Message → Consensus → Memo，报告与知识提取回业务域。

## 架构

- qtcloud-feishu：平台动作、定义、流程、镜像与游标。
- 摄取与映射（待建）：读镜像或行数据，按来源 id 幂等写入 Message，保留平台引用；先做成批处理工具，可挂 CLI，不进入 Provider 在线路径。
- Provider：Message、Consensus、Relation、ConsensusGraph 的持久化与 HTTP API，是事实源。
- CLI：共识创建、查询、确认与废弃，兼作批处理与脚本入口。
- Studio：共识图、节点编辑与人工确认。
- 业务层：qtrecurit 等消费共识与消息，产出报告、话术与推荐。

数据流一句话：适配层取数，摄取向领域模型，Provider 持久化，CLI 与 Studio 操作确认，业务层产出。

## 当前目标

v0.2 的目标按依赖排：

1. 完成库与示例的移除；`mail` 与 `notice` 的命令面等 qtcloud-feishu 入口就绪后退出，CLI 最终不再有飞书命令。
2. 定摄取契约：来源、来源 id、发生时间、正文、类型与元数据；先接 qtcloud-feishu 的 `raw/im` 与 `raw/mail`。
3. 打通溯源：Message 与 Consensus 的多对多关系要有写入与查询接口。
4. 给 Memo 留出模型与接口边界，生成留到 v0.3。

### 摄取契约草案

Message 增加来源字段 `source`、`source_id`、`occurred_at` 与 `metadata`，以 `(source, source_id)` 唯一约束保证重复导入不产生重复记录。

Provider 增加 `POST /api/messages`（幂等）、`POST /api/relations` 与 `GET /api/consensuses/{id}/messages`；现有 `GET /api/messages` 保持只读兼容。

适配层的 `raw/` 是原样留底，摄取只读它，不改写镜像格式。摄取失败不回写平台，重跑从游标继续。

## 组件

### Provider

Go、SQLite 与标准库 `net/http`。现有 API 覆盖消息只读、共识增删改查与状态、共识关系与 DAG 图。

不变量沿用：Message 正文不可变；共识关系拒绝自环与成环；从图里移出节点或边不删除全局对象；图是视图，共识是事实。

### CLI

Rust、clap 与 ureq。移除飞书能力后保留 `consensus`；若摄取放这里，另加 `ingest`，仍不碰平台凭据。通过 `--endpoint` 与 `CONNECT_AUTH_TOKEN` 访问 Provider。

### Studio

Flutter。共识追溯页面与白板，通过 `CONNECT_PROVIDER_ENDPOINT` 取数。静态产物不携带服务令牌，公网部署由网关完成用户认证。

## 开发

```bash
# Provider
cd src/provider
go test ./...
go vet ./...
go run cmd/server/main.go

# CLI
cd ../cli
cargo fmt --check
cargo test --locked
cargo clippy --locked -- -A warnings
cargo run -- consensus list

# Studio
cd ../studio
flutter pub get
flutter analyze
flutter test
flutter run -d chrome --dart-define=CONNECT_PROVIDER_ENDPOINT=http://localhost:8000/api
```

## 验收

`tests/verify-v0.1.ps1` 覆盖 Provider、CLI 与 Studio 的共识闭环，移除飞书能力后应原样通过。摄取与溯源在 v0.2 另加集成测试，用固定镜像夹具，不联网。

## 发布

CLI 通过 `cli/*` 标签发布，移除飞书能力属于破坏性变更，走 v0.2.0。Provider 先通过测试与构建再部署，Studio 产出静态资源后部署到目标域名。三者合并完成后再更新根仓子模块指针。

## 路线

- v0.1：共识追溯闭环，已完成。
- v0.1.x：冻结并标记飞书能力，替代在 qtcloud-feishu 落地。
- v0.2：移除飞书库与示例，建立消息摄取与溯源；命令面按替代入口就绪情况退出。
- v0.3：共识生产，从消息草拟共识并生成 Memo。
- v0.4：多平台与业务系统，企业微信接入不改领域模型，接 qtdata。

## 文档

- [群与渠道管理（lite）](lite.md)：群资产卡片化、名称与职责描述、群成员与群间关系，渠道先手动管理、适配层后接。
- [飞书功能废弃与移除](feishu-removal.md)
- [开发者指南概览](README.md)
- [Provider API](../api-references/provider.md)
- [用户指南](../user-guide/index.md)
