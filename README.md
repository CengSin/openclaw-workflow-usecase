# openclaw-workflow-usecase

> 用 **Temporal** 编排 **OpenClaw Agent**，构建可靠、可重试的 AI 自动化工作流。

本项目展示了如何将 OpenClaw 的本地 Agent 能力接入生产级工作流引擎，以"新能源行业脱水研报生成"为示例场景，实现：**研报检索 → 数据清洗 → 脱水成文** 的三阶段自动化 pipeline。

---

## 效果示例

输入一句话：

```
写一篇关于新能源行业过去一周有关的研报的脱水研报
```

自动完成三个阶段，最终输出：

```markdown
# 新能源行业脱水研报（2026.06.17-2026.06.24）

## 核心结论
1. 全球电力需求已进入加速增长期，新能源替代传统能源进程确定性推进。
2. 亚洲已成为全球新能源发展核心引擎，中国新能源装机规模全球第一。
3. AI 数据中心建设带来新增需求，驱动电力设备领域技术升级与出口突破。

## 关键数据
| 核心指标 | 数据 |
|---------|------|
| 2025 年底中国可再生能源装机 | 23.4 亿千瓦，历史性超越火电占比 |
| 2030 年前全球电力需求年均增速 | >3.5%，为整体能源需求增速的 2.5 倍 |
...
```

---

## 为什么要用 Temporal？

直接调用 LLM API 的问题：一旦中间某步失败，整个流程需要从头开始，没有可观测性，也没有办法单独重试某一步。

Temporal 解决了这些问题：

- **自动重试**：每个 Activity 失败后按策略自动重试，不影响其他步骤
- **幂等性**：同一 Workflow 实例不会重复执行已完成的步骤
- **可观测**：通过 Temporal Web UI 实时查看每个 Activity 的状态和输出
- **可扩展**：新增更多 Agent 步骤只需添加 Activity，Workflow 结构清晰

---

## 架构

```
用户输入 (topic)
     │
     ▼
┌─────────────────────────────────────────────┐
│          ResearchReportWorkflow              │
│                                              │
│  Activity 1: FetchResearchReports            │
│  └─ Chronicle Agent 检索相关研报与原始数据    │
│               │                             │
│               ▼                             │
│  Activity 2: CleanResearchData              │
│  └─ Chronicle Agent 去重、提炼、结构化输出   │
│               │                             │
│               ▼                             │
│  Activity 3: WriteCondensedResearchReport   │
│  └─ Chronicle Agent 撰写最终脱水研报         │
└─────────────────────────────────────────────┘
     │
     ▼
最终脱水研报（Markdown 格式）
```

每个 Activity 通过 `openclaw-go` SDK 调用本地 OpenClaw 的 `/v1/chat/completions` 接口，由 `x-openclaw-agent-id` 将请求路由到指定 Agent。

---

## 前置条件

- Go 1.22+
- [OpenClaw](https://openclaw.ai) 已在本地运行（默认地址 `http://localhost:18789`）
- Chronicle Agent 已在 OpenClaw 中配置并启用
- [Temporal Server](https://docs.temporal.io/self-hosted-guide) 已在本地运行（默认地址 `localhost:7233`）

快速启动本地 Temporal（需要 Docker）：

```bash
brew install temporal   # macOS
temporal server start-dev
```

---

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/CengSin/openclaw-workflow-usecase.git
cd openclaw-workflow-usecase

# 运行（一体化模式，自动内嵌启动 Worker）
go run . "写一篇关于新能源行业过去一周有关的研报的脱水研报"

# 或者传入自定义主题
go run . "写一篇关于大模型行业最新进展的脱水研报"
```

---

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `OPENCLAW_BASE_URL` | `http://localhost:18789` | OpenClaw 本地服务地址 |
| `OPENCLAW_TOKEN` | （空） | 鉴权 Token，本地使用时通常不需要 |
| `OPENCLAW_FETCH_AGENT_ID` | `chronicle` | 负责"研报检索"的 Agent |
| `OPENCLAW_CLEAN_AGENT_ID` | `chronicle` | 负责"数据清洗"的 Agent |
| `OPENCLAW_WRITE_AGENT_ID` | `chronicle` | 负责"撰写成文"的 Agent |

---

## 运行集成测试

```bash
OPENCLAW_INTEGRATION=1 \
OPENCLAW_TEST_TOPIC="写一篇关于新能源行业过去一周有关的研报的脱水研报" \
go test ./workflow -run TestResearchReportWorkflow_Integration -v
```

不设置 `OPENCLAW_INTEGRATION=1` 时，集成测试会自动跳过，`go test ./...` 可安全在 CI 中运行。

---

## 项目结构

```
.
├── main.go                              # 一体化入口：启动 Worker + 提交 Workflow
├── namespace/
│   └── namespace.go                     # Temporal Task Queue 名称定义
└── workflow/
    ├── workflow.go                      # 主 Workflow：编排三个 Activity
    ├── workflow_integration_test.go     # 本地集成测试
    └── activity/
        ├── activity.go                  # 三个 Activity 实现（调用 OpenClaw）
        └── model.go                     # 默认常量（Agent ID、Model 名称）
```

---

## 下一步计划

- [ ] 结构化 Workflow 输入参数（时间范围、行业范围、报告风格、目标读者）
- [ ] 增加研报溯源 Activity，为最终报告附带来源、机构、发布时间标注
- [ ] 支持为 fetch / clean / write 配置不同的专职 Agent
- [ ] 加入人工审核 Signal，在发布前触发人工确认节点
- [ ] 更精细的错误分类（鉴权错误标记为不可重试，网络超时允许重试）

---

## 依赖

- [openclaw-go](https://github.com/a3tai/openclaw-go) — OpenClaw 官方 Go SDK
- [Temporal Go SDK](https://github.com/temporalio/sdk-go) — 工作流引擎 SDK

---

## License

MIT