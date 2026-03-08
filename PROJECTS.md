# OpenClaw Workflow Usecase

## 项目目标

使用 Temporal 对 OpenClaw Agent 能力进行工作流编排，当前先实现一个最小可运行链路：

核心目标：
- 由 Workflow 触发本地 OpenClaw 的 `Chronicle` agent。
- 把用户需求交给 3 个不同职责的 agent，依次完成研报抓取、数据整理清洗、脱水研报生成。
- 工作流执行失败时自动重试（基础 RetryPolicy）。
- 提供可直接运行的本地测试入口。

---

## 当前代码结构

```text
.
├── main.go
├── namespace/
│   └── namespace.go
└── workflow/
    ├── workflow.go
    ├── workflow_integration_test.go
    └── activity/
        ├── activity.go
        └── model.go
```

说明：
- `namespace/namespace.go`：定义 Temporal Task Queue 名称。
- `main.go`：开发态一体化入口，会先启动内嵌 Worker，再提交 Workflow 并等待最终脱水研报结果。
- `workflow/workflow.go`：Temporal 工作流，按“抓研报 -> 清洗整理 -> 生成脱水研报”顺序编排 3 个 Activity。
- `workflow/activity/activity.go`：通过 `openclaw-go/chatcompletions` 调本地 OpenClaw。
- `workflow/workflow_integration_test.go`：本地集成测试（可选开关执行，避免 CI/本地无服务时误失败）。

---

## OpenClaw 文档对齐结论

- 使用 `chatcompletions.Client` 调用 `POST /v1/chat/completions`。
- 通过 `AgentID` 自动带 `x-openclaw-agent-id`，由 OpenClaw 负责把请求路由到对应 agent。
- `Model` 仅作为 OpenAI-compatible 请求体字段保留，不再承担 agent 路由语义。
- 默认可把 3 个 Activity 都路由到 `Chronicle`，也支持通过环境变量切换为不同 agent。
- 支持 `OPENCLAW_BASE_URL`、`OPENCLAW_TOKEN`、`OPENCLAW_FETCH_AGENT_ID`、`OPENCLAW_CLEAN_AGENT_ID`、`OPENCLAW_WRITE_AGENT_ID`。

---

## 当前实现状态（已完成）

- ✅ 修复原先无法编译的问题（未定义类型/函数、注册不一致、空入口）。
- ✅ 新建主 Workflow：`ResearchReportWorkflow(topic string) -> report string`。
- ✅ 新建 3 个职责型 Activity：
  - `FetchResearchReports`
  - `CleanResearchData`
  - `WriteCondensedResearchReport`
- ✅ 根目录入口已注册新 Workflow 与 Activity。
- ✅ 根目录入口已支持内嵌启动 Worker，`go run . "<主题>"` 可直接本地联调。
- ✅ 新增集成测试：`workflow/workflow_integration_test.go`。
- ✅ `go test ./...` 可通过（默认跳过真正的 OpenClaw 集成调用）。
- ✅ 每个 Activity 只约束“职责边界”，不再强制 agent 内部必须按某个 skills 链执行。
- ✅ 每次发给 OpenClaw 的 prompt 都会在日志中完整打印，方便调试。

### 当前三阶段流程定义

1. 第一阶段 Activity：抓取相关研报
   - 理解用户需求
   - 自主调用可用 skills / tools 检索候选研报、原始资料与相关数据
   - 输出原始检索资料文档
2. 第二阶段 Activity：筛选重要信息
   - 对原始资料做去重、归并、清洗和重点提炼
   - 输出结构化研报摘要
3. 第三阶段 Activity：整合成文
   - 根据结构化研报摘要生成最终脱水研报

---

## 本地运行说明

### 1) 直接本地运行（一体化模式）

```bash
go run . "写一篇关于新能源行业过去一周有关的研报的脱水研报"
```

### 2) 执行 OpenClaw 集成测试（可选）

```bash
OPENCLAW_INTEGRATION=1 OPENCLAW_TEST_TOPIC="写一篇关于新能源行业过去一周有关的研报的脱水研报" go test ./workflow -run TestResearchReportWorkflow_Integration -v
```

可选环境变量：
- `OPENCLAW_BASE_URL`（默认 `http://localhost:18789`）
- `OPENCLAW_TOKEN`（本地网关若需要鉴权时填写）
- `OPENCLAW_FETCH_AGENT_ID`（默认 `Chronicle`）
- `OPENCLAW_CLEAN_AGENT_ID`（默认 `Chronicle`）
- `OPENCLAW_WRITE_AGENT_ID`（默认 `Chronicle`）

## 下一步建议

- 将第一阶段检索结果进一步结构化输出（候选研报、关键数据点、原始摘录、来源列表）。
- 将第二阶段清洗结果进一步结构化输出（核心观点、关键数字、风险点、引用来源）。
- 增加 Workflow 输入参数对象（主题、时间范围、行业范围、报告风格、目标读者）。
- 在 Activity 中补充超时与错误分类（鉴权错误/网络错误/模型错误/检索无结果错误）。
- 新增研报溯源能力，为最终脱水研报附带来源、机构、发布时间等引用标注。
- 下一阶段再接入人工审核 signal 与 RAG 入库。

