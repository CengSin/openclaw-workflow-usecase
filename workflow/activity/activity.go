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

// RewriteResearchQueryWithChronicle 将用户主题改写成适合 research-ai-picker 执行的检索句。
func RewriteResearchQueryWithChronicle(ctx context.Context, topic string) (string, error) {
	baseURL := os.Getenv("OPENCLAW_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:18789"
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if strings.TrimSpace(topic) == "" {
		return "", errors.New("查询主题不能为空")
	}

	client := &chatcompletions.Client{
		BaseURL: baseURL,
		Token:   os.Getenv("OPENCLAW_TOKEN"),
		AgentID: "chronicle",
	}

	prompt := fmt.Sprintf(
		`请把下面的用户主题改写成一句适合后续 skill "%s" 执行的检索指令。
用户主题：%s
改写要求：
- 输出一句自然中文指令即可，不要输出解释。
- 指令中应明确检索对象、时间范围、行业/公司范围、目标信息类型。
- 语气要像真实可执行的查询，例如：“调用 %s 查询过去一周新能源行业的相关研报”。
- 如果原主题缺少时间范围，可默认补成“近期”或“过去一周”，但不要编造具体事实。
- 不要添加 markdown、序号、引号或额外说明。`,
		ChronicleSkillResearchAIPicker,
		topic,
		ChronicleSkillResearchAIPicker,
	)

	logChroniclePrompt("rewrite_query", prompt)

	resp, err := client.Create(reqCtx, chatcompletions.Request{
		Model: "openclaw:chronicle",
		Messages: []chatcompletions.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		if httpErr, ok := errors.AsType[*chatcompletions.HTTPError](err); ok {
			switch httpErr.StatusCode {
			case 401:
				return "", fmt.Errorf(
					"调用 Chronicle 失败: OpenClaw 需要鉴权，请设置 OPENCLAW_TOKEN（baseURL=%s）: %w",
					baseURL,
					err,
				)
			case 404:
				return "", fmt.Errorf(
					"调用 Chronicle 失败: %s 没有暴露 /v1/chat/completions，当前地址更像 OpenClaw Control UI，请检查 OPENCLAW_BASE_URL 或本地 API 能力: %w",
					baseURL,
					err,
				)
			}
		}
		return "", fmt.Errorf("调用 Chronicle 失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("Chronicle 返回为空（choices=0）")
	}

	rewritten := strings.TrimSpace(resp.Choices[0].Message.Content)
	if rewritten == "" {
		return "", errors.New("Chronicle 返回重写查询为空")
	}

	return rewritten, nil
}

// RetrieveResearchDataWithChronicle 调用 Chronicle 完成第一阶段: 研报筛选与数据检索。
func RetrieveResearchDataWithChronicle(ctx context.Context, req ResearchRetrievalRequest) (ResearchRetrievalResult, error) {
	baseURL := os.Getenv("OPENCLAW_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:18789"
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	if strings.TrimSpace(req.Topic) == "" {
		return ResearchRetrievalResult{}, errors.New("检索主题不能为空")
	}

	requiredSkills := req.RequiredSkills
	if len(requiredSkills) == 0 {
		requiredSkills = []string{
			ChronicleSkillResearchAIPicker,
			ChronicleSkillQVeris,
		}
	}

	query := strings.TrimSpace(req.RewrittenQuery)
	if query == "" {
		query = req.Topic
	}

	client := &chatcompletions.Client{
		BaseURL: baseURL,
		Token:   os.Getenv("OPENCLAW_TOKEN"),
		AgentID: "chronicle",
	}

	logChroniclePrompt("retrieve_research", query)

	resp, err := client.Create(reqCtx, chatcompletions.Request{
		Model: "openclaw:chronicle",
		Messages: []chatcompletions.Message{
			{Role: "user", Content: query},
		},
	})
	if err != nil {
		if httpErr, ok := errors.AsType[*chatcompletions.HTTPError](err); ok {
			switch httpErr.StatusCode {
			case 401:
				return ResearchRetrievalResult{}, fmt.Errorf(
					"调用 Chronicle 失败: OpenClaw 需要鉴权，请设置 OPENCLAW_TOKEN（baseURL=%s）: %w",
					baseURL,
					err,
				)
			case 404:
				return ResearchRetrievalResult{}, fmt.Errorf(
					"调用 Chronicle 失败: %s 没有暴露 /v1/chat/completions，当前地址更像 OpenClaw Control UI，请检查 OPENCLAW_BASE_URL 或本地 API 能力: %w",
					baseURL,
					err,
				)
			}
		}
		return ResearchRetrievalResult{}, fmt.Errorf("调用 Chronicle 失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ResearchRetrievalResult{}, errors.New("Chronicle 返回为空（choices=0）")
	}

	document := strings.TrimSpace(resp.Choices[0].Message.Content)
	if document == "" {
		return ResearchRetrievalResult{}, errors.New("Chronicle 返回内容为空")
	}

	return ResearchRetrievalResult{
		Document:   document,
		Query:      query,
		SkillChain: requiredSkills,
	}, nil
}

func logChroniclePrompt(stage string, prompt string) {
	log.Printf("=== Chronicle prompt [%s] begin ===\n%s\n=== Chronicle prompt [%s] end ===", stage, prompt, stage)
}
