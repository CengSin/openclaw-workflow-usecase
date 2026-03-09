package activity

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// VerifySources 对清洗后的摘要进行溯源校验，过滤幻觉来源（如未来日期、不可核实引用）。
func VerifySources(ctx context.Context, cleanedResearch string) (string, error) {
	if strings.TrimSpace(cleanedResearch) == "" {
		return "", errors.New("待校验的研报摘要不能为空")
	}

	prompt := fmt.Sprintf(
		`你是一名负责"溯源校验"的 agent，专门识别并过滤研报中的幻觉来源。

下面是已整理的结构化研报摘要：
%s

你的任务：
1. 逐条检查每个来源引用：
   - 发布日期是否合理（不能是未来日期，不能早于该机构成立时间）
   - 发布机构是否真实存在
   - 引用的核心数据是否与正文描述一致
2. 对每条来源标注校验结果：✅ 可信 / ⚠️ 存疑 / ❌ 疑似幻觉
3. 移除所有标注为 ❌ 的内容，对 ⚠️ 内容保留但加注说明。

输出要求：
- 输出"校验后的结构化研报摘要"，格式与输入保持一致。
- 在文末附上"溯源校验报告"，列出：
  1. 移除的幻觉来源列表（含原因）
  2. 存疑来源列表（含疑点说明）
  3. 校验通过的来源数量
- 如果所有来源均通过校验，直接输出原摘要并附简短校验报告。`,
		cleanedResearch,
	)

	return callAgent(ctx, verifyAgentID(), "verify_sources", prompt)
}
