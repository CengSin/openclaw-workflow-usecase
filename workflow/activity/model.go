package activity

const (
	ChronicleSkillQVeris           = "qveris"
	ChronicleSkillResearchAIPicker = "research-ai-picker"
)

type ResearchRetrievalRequest struct {
	Topic          string   `json:"topic"`
	RewrittenQuery string   `json:"rewrittenQuery"`
	RequiredSkills []string `json:"requiredSkills"`
}

type ResearchRetrievalResult struct {
	Document   string   `json:"document"`
	Query      string   `json:"query"`
	SkillChain []string `json:"skillChain"`
}
