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
					"调用 agent=%s 失败: %s 没有暴露 /v1/chat/completions，请检查 OPENCLAW_BASE_URL: %w",
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

func verifyAgentID() string {
	return envOrDefault("OPENCLAW_VERIFY_AGENT_ID", DefaultOpenClawAgentID)
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
