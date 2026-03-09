package activity

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func WriteCondensedResearchReport(ctx context.Context, input ResearchReportInput, verifiedResearch string) (string, error) {
	if strings.TrimSpace(verifiedResearch) == "" {
		return "", errors.New("校验后的研报摘要不能为空")
	}

	prompt := fmt.Sprintf(
		`
下面是已经整理、清洗并经过溯源校验的结构化研报摘要：
%s

输出风格：%s

你的任务：
- 基于这些已经整理过的信息，写一篇中文研报。
- 保持专业、简洁、高信息密度。
- 不要重复原始噪声，不要罗列无关细节。
- 只引用溯源校验通过的来源，不引用已被标记为幻觉的内容。

输出要求：
- 输出结构建议：
  1. 标题
  2. 核心结论
  3. 关键催化与数据
  4. 风险与不确定性
  5. 参考来源（仅保留校验通过的来源）
- 文章要可直接阅读，不要输出过程说明。`,
		verifiedResearch,
		input.effectiveStyle(),
	)

	return callAgent(ctx, writeAgentID(), "write_condensed_report", prompt)
}
