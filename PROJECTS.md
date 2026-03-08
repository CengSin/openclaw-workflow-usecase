# OpenClaw Workflow Usecase

## 项目目标

使用 Temporal 对 OpenClaw Agent 能力进行工作流编排，当前先实现一个最小可运行链路：

核心目标：
- 由 Workflow 触发本地 OpenClaw 的 `Chronicle` agent。
- 把“文章主题”交给 agent，先完成行业研报数据检索，再进入后续的信息筛选与成文整合。
- 工作流执行失败时自动重试（基础 RetryPolicy）。
- 提供可直接运行的本地测试入口。

---

## 当前代码结构

```text
.
├── main.go
├── namespace/
│   └── namespace.go
├── worker/
│   └── worker.go
└── workflow/
    ├── workflow.go
    ├── workflow_integration_test.go
    └── activity/
        ├── activity.go
        └── model.go
```

说明：
- `namespace/namespace.go`：定义 Temporal Task Queue 名称。
- `main.go`：开发态一体化入口，会先启动内嵌 Worker，再提交 Workflow 并等待检索结果文档。
- `worker/worker.go`：独立 Worker 启动入口，适合与 Workflow 提交端分离部署。
- `workflow/workflow.go`：Temporal 工作流（先执行“查询改写” Activity，再执行第一阶段“研报数据检索” Activity，并携带重试策略）。
- `workflow/activity/activity.go`：通过 `openclaw-go/chatcompletions` 调本地 OpenClaw。
- `workflow/workflow_integration_test.go`：本地集成测试（可选开关执行，避免 CI/本地无服务时误失败）。

---

## OpenClaw 文档对齐结论

- 使用 `chatcompletions.Client` 调用 `POST /v1/chat/completions`。
- 通过 `AgentID: "Chronicle"` 自动带 `x-openclaw-agent-id`，实现定向到本地 Chronicle agent。
- `Model` 使用 `openclaw:chronicle`。
- 当前工作流会先让 Chronicle 把用户主题改写成更适合 `research-ai-picker` 的检索句，再执行检索阶段。
- 支持 `OPENCLAW_BASE_URL` 与 `OPENCLAW_TOKEN` 环境变量。

---

## 当前实现状态（已完成）

- ✅ 修复原先无法编译的问题（未定义类型/函数、注册不一致、空入口）。
- ✅ 新建最小 Workflow：`ChronicleResearchWorkflow(topic string) -> document string`。
- ✅ 新增前置 Activity：`RewriteResearchQueryWithChronicle(topic)`，将用户主题改写为适合后续 skill 执行的检索指令。
- ✅ 新建第一阶段 Activity：`RetrieveResearchDataWithChronicle(request)` 调用本地 OpenClaw，并强制先后使用 `research-ai-picker`、`qveris` 两个 Skills。
- ✅ Worker 已注册新 Workflow 与 Activity。
- ✅ 根目录入口已支持内嵌启动 Worker，`go run . "<主题>"` 可直接本地联调。
- ✅ 新增集成测试：`workflow/workflow_integration_test.go`。
- ✅ `go test ./...` 可通过（默认跳过真正的 OpenClaw 集成调用）。
- ✅ Chronicle Agent已支持行业研报检索能力，生成内容自动关联最新研报数据、行业动态与权威分析。
- ✅ 研报检索依赖的两个核心Skill：
  - `~/.openclaw/skills/qveris/` - QVeris金融数据Skill，支持研报、行情、财务数据检索
  - `~/.openclaw/skills/research-ai-picker/` - Research AI Picker研报筛选Skill
- ✅ 第一个 Activity 已显式要求 Chronicle 按 `research-ai-picker -> qveris` 的顺序完成研报筛选与数据补强，并输出“检索结果文档”供后续 Activity 使用。

### 当前三阶段流程定义

0. 前置 Activity：改写查询
   - 将用户主题改写成适合 `research-ai-picker` 执行的一句话检索指令
   - 例如：`调用 research-ai-picker 查询过去一周新能源行业的相关研报`
1. 第一阶段 Activity：检索研报数据
   - 检索候选研报
   - 补齐金融/行业/财务数据
   - 输出原始检索文档
2. 第二阶段 Activity：筛选重要信息
   - 从原始检索文档中提取重点事实、核心观点、关键数字
3. 第三阶段 Activity：整合成文
   - 将筛选后的重点信息整理成最终文章/周报

---

## 本地运行说明

### 1) 直接本地运行（一体化模式）

```bash
# 先执行第一阶段研报检索
go run . "半导体行业周报"
# 先执行第一阶段研报检索
go run . "2026年Q1消费电子行业研报总结"
```

### 2) 分离运行（可选）

```bash
go run ./worker
go run . "半导体行业周报"
```

### 3) 执行 OpenClaw 集成测试（可选）

```bash
OPENCLAW_INTEGRATION=1 OPENCLAW_TEST_TOPIC="消费电子行业观察" go test ./workflow -run TestChronicleResearchWorkflow_Integration -v
```

可选环境变量：
- `OPENCLAW_BASE_URL`（默认 `http://localhost:18789`）
- `OPENCLAW_TOKEN`（本地网关若需要鉴权时填写）

## 下一步建议

- 将前置改写结果也结构化保留（时间范围、行业范围、研报类型、筛选条件）。
- 将第一阶段检索结果进一步结构化输出（候选研报、关键数据点、原始摘录、来源列表）。
- 增加 Workflow 输入参数对象（主题、字数、语气、目标读者、研报时间范围、数据源筛选规则）。
- 在 Activity 中补充超时与错误分类（鉴权错误/网络错误/模型错误/研报检索无结果错误）。
- 新增研报溯源能力，为后续第二、第三阶段提供来源、机构、发布时间等引用标注。
- 下一阶段再接入人工审核 signal 与 RAG 入库。

