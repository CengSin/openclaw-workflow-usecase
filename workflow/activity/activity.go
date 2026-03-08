package activity

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/a3tai/openclaw-go/chatcompletions"
)

func FetchResearchReports(ctx context.Context, topic string) (string, error) {
	if strings.TrimSpace(topic) == "" {
		return "", errors.New("用户需求不能为空")
	}

	prompt := fmt.Sprintf(
		`请根据用户需求进行“研报检索”
用户需求：
%s

你的任务：
- 理解用户真正需要的研报范围、时间范围、行业范围和重点信息。
- 自主使用你可用的 tools / skills / data sources 去检索最相关的研报、行业资料和原始数据。
- 当前阶段只负责“抓取资料”，不要做最终成文。

输出要求：
- 输出一份“原始检索资料文档”。
- 内容至少包含：
  1. 检索目标理解
  2. 命中的候选研报列表
  3. 每份研报的核心观点摘录
  4. 关键行业/财务/市场数据
  5. 来源信息
- 如果资料不足，请明确指出缺口，不要编造内容。`,
		topic,
	)

	return callAgent(ctx, fetchAgentID(), "fetch_research_reports", prompt)
}

func CleanResearchData(ctx context.Context, rawResearch string) (string, error) {
	if strings.TrimSpace(rawResearch) == "" {
		return "", errors.New("原始检索资料不能为空")
	}

	prompt := fmt.Sprintf(
		`你是一名负责“数据整理与清洗”的 agent。

下面是第一阶段得到的原始检索资料：
%s

你的任务：
- 去重、归并相似研报和重复信息。
- 提炼最重要的事实、观点、关键数字和潜在矛盾点。
- 清洗掉噪声信息、低价值描述和重复表述。
- 保留能够支撑最终脱水研报的核心证据。

输出要求：
- 输出一份“清洗后的结构化研报摘要”。
- 内容至少包含：
  1. 核心结论
  2. 重点研报与来源
  3. 关键数据点
  4. 重要观点与分歧
  5. 仍待确认的信息
- 不要生成最终文章，不要写成完整研报。`,
		rawResearch,
	)

	return callAgent(ctx, cleanAgentID(), "clean_research_data", prompt)
}

func WriteCondensedResearchReport(ctx context.Context, cleanedResearch string) (string, error) {
	if strings.TrimSpace(cleanedResearch) == "" {
		return "", errors.New("清洗后的研报摘要不能为空")
	}

	prompt := fmt.Sprintf(
		`你是一名负责“撰写脱水研报”的 agent。

下面是已经整理清洗完成的结构化研报摘要：
%s

你的任务：
- 基于这些已经整理过的信息，写一篇中文“脱水研报”。
- 保持专业、简洁、高信息密度。
- 不要重复原始噪声，不要罗列无关细节。

输出要求：
- 输出结构建议：
  1. 标题
  2. 核心结论
  3. 关键催化与数据
  4. 风险与不确定性
  5. 参考来源
- 文章要可直接阅读，不要输出过程说明。`,
		cleanedResearch,
	)

	return callAgent(ctx, writeAgentID(), "write_condensed_report", prompt)
}

func callAgent(ctx context.Context, agentID string, stage string, prompt string) (string, error) {
	baseURL := os.Getenv("OPENCLAW_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:18789"
	}

	model := os.Getenv("OPENCLAW_MODEL")
	if model == "" {
		model = DefaultOpenClawModel
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	client := &chatcompletions.Client{
		BaseURL: baseURL,
		Token:   os.Getenv("OPENCLAW_TOKEN"),
		AgentID: agentID,
	}

	logOpenClawPrompt(stage, agentID, model, prompt)

	resp, err := client.Create(reqCtx, chatcompletions.Request{
		// Agent route is selected by x-openclaw-agent-id. Model keeps the HTTP API request compatible.
		Model: model,
		Messages: []chatcompletions.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		var httpErr *chatcompletions.HTTPError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 401:
				return "", fmt.Errorf(
					"调用 agent=%s 失败: OpenClaw 需要鉴权，请设置 OPENCLAW_TOKEN（baseURL=%s）: %w",
					agentID,
					baseURL,
					err,
				)
			case 404:
				return "", fmt.Errorf(
					"调用 agent=%s 失败: %s 没有暴露 /v1/chat/completions，当前地址更像 OpenClaw Control UI，请检查 OPENCLAW_BASE_URL 或本地 API 能力: %w",
					agentID,
					baseURL,
					err,
				)
			}
		}
		return "", fmt.Errorf("调用 agent=%s 失败: %w", agentID, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("调用 agent=%s 失败: 返回为空（choices=0）", agentID)
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("调用 agent=%s 失败: 返回内容为空", agentID)
	}

	return content, nil
}

func logOpenClawPrompt(stage string, agentID string, model string, prompt string) {
	log.Printf(
		"=== OpenClaw prompt [%s] begin ===\nagent=%s\nmodel=%s\n%s\n=== OpenClaw prompt [%s] end ===",
		stage,
		agentID,
		model,
		prompt,
		stage,
	)
}

func fetchAgentID() string {
	return envOrDefault("OPENCLAW_FETCH_AGENT_ID", DefaultOpenClawAgentID)
}

func cleanAgentID() string {
	return envOrDefault("OPENCLAW_CLEAN_AGENT_ID", DefaultOpenClawAgentID)
}

func writeAgentID() string {
	return envOrDefault("OPENCLAW_WRITE_AGENT_ID", DefaultOpenClawAgentID)
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
