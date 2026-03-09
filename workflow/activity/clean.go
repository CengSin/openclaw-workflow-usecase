package activity

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func CleanResearchData(ctx context.Context, rawResearch string) (string, error) {
	if strings.TrimSpace(rawResearch) == "" {
		return "", errors.New("原始检索资料不能为空")
	}

	prompt := fmt.Sprintf(
		`
下面是第一阶段得到的原始检索资料：
%s

你的任务：
- 去重、归并相似研报和重复信息。
- 提炼最重要的事实、观点、关键数字和潜在矛盾点。
- 清洗掉噪声信息、低价值描述和重复表述。
- 保留能够支撑最终脱水研报的核心证据。

输出要求：
- 输出一份"清洗后的结构化研报摘要"。
- 内容至少包含：
  1. 核心结论
  2. 重点研报与来源（保留原始发布机构、发布日期、URL）
  3. 关键数据点
  4. 重要观点与分歧
  5. 仍待确认的信息
- 不要生成最终文章，不要写成完整研报。`,
		rawResearch,
	)

	return callAgent(ctx, cleanAgentID(), "clean_research_data", prompt)
}
