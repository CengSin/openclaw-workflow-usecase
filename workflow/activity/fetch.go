package activity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func FetchResearchReports(ctx context.Context, input ResearchReportInput) (string, error) {
	if strings.TrimSpace(input.Topic) == "" {
		return "", errors.New("Topic 不能为空")
	}

	var constraints strings.Builder
	if input.TimeRange != "" {
		fmt.Fprintf(&constraints, "- 时间范围：%s\n", input.TimeRange)
	}
	if input.Industry != "" {
		fmt.Fprintf(&constraints, "- 行业范围：%s\n", input.Industry)
	}
	if input.MaxSources > 0 {
		fmt.Fprintf(&constraints, "- 最多引用来源数：%d\n", input.MaxSources)
	}

	prompt := fmt.Sprintf(
		`请根据用户需求进行"研报检索"
研究主题：%s
%s
当前时间：%s

你的任务：
- 理解用户真正需要的研报范围、时间范围、行业范围和重点信息。
- 自主使用你可用的 tools / skills / data sources 去检索最相关的研报、行业资料和原始数据。
- 当前阶段只负责"抓取资料"，不要做最终成文。
- 对第一步找到的每个 URL，调用 web_fetch 读取完整正文：
	- 提取标题、发布机构、发布日期、核心观点、关键数据
	- 对无法访问的链接标注原因（付费墙/需登录/链接失效）
	- 对疑似幻觉的来源（未来日期、域名与机构不符等）标注并移除

输出要求：
- 输出一份"原始检索资料文档"。
- 内容至少包含：
  1. 检索目标理解
  2. 命中的候选研报列表（含发布机构、发布日期、标题）
  3. 每份研报的核心观点摘录
  4. 关键行业/财务/市场数据
  5. 来源信息（URL 或可核实的引用）
- 如果资料不足，请明确指出缺口，不要编造内容。`,
		input.Topic,
		constraints.String(),
		time.Now().Format(time.DateTime),
	)

	return callAgent(ctx, fetchAgentID(), "fetch_research_reports", prompt)
}
