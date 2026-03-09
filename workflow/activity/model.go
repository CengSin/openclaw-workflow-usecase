package activity

const (
	DefaultOpenClawAgentID = "chronicle"
	DefaultOpenClawModel   = "openclaw:chronicle"
)

// ResearchReportInput 是 ResearchReportWorkflow 的结构化输入。
// 零值字段表示"不限制"，workflow 内部会用合理默认值填充。
type ResearchReportInput struct {
	Topic      string // 必填，研究主题
	TimeRange  string // 时间范围，如"过去一周"
	Industry   string // 行业范围，如"新能源"
	Style      string // 输出风格："脱水研报" / "深度分析"，默认"脱水研报"
	MaxSources int    // 最多引用来源数，0 表示不限
}

// effectiveStyle 返回有效的输出风格，空值时回退到默认值。
func (r ResearchReportInput) effectiveStyle() string {
	if r.Style == "" {
		return "脱水研报"
	}
	return r.Style
}
