package workflow

import (
	"ai.openclaw.usecase/workflow/activity"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/testsuite"
)

func TestResearchReportWorkflow_Integration(t *testing.T) {
	if os.Getenv("OPENCLAW_INTEGRATION") != "1" {
		t.Skip("设置 OPENCLAW_INTEGRATION=1 后执行本地 OpenClaw 集成测试")
	}

	input := activity.ResearchReportInput{
		Topic:     "新能源行业过去一周研报",
		TimeRange: "过去一周",
		Industry:  "新能源",
		Style:     "脱水研报",
	}

	if topic := os.Getenv("OPENCLAW_TEST_TOPIC"); strings.TrimSpace(topic) != "" {
		input.Topic = topic
	}
	if tr := os.Getenv("OPENCLAW_TEST_TIME_RANGE"); strings.TrimSpace(tr) != "" {
		input.TimeRange = tr
	}
	if ind := os.Getenv("OPENCLAW_TEST_INDUSTRY"); strings.TrimSpace(ind) != "" {
		input.Industry = ind
	}
	if style := os.Getenv("OPENCLAW_TEST_STYLE"); strings.TrimSpace(style) != "" {
		input.Style = style
	}

	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.SetTestTimeout(30 * time.Minute)
	env.RegisterWorkflow(ResearchReportWorkflow)
	env.RegisterActivity(activity.FetchResearchReports)
	env.RegisterActivity(activity.CleanResearchData)
	env.RegisterActivity(activity.VerifySources)
	env.RegisterActivity(activity.WriteCondensedResearchReport)

	env.ExecuteWorkflow(ResearchReportWorkflow, input)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow 未完成")
	}

	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow 失败: %v", err)
	}

	var report string
	if err := env.GetWorkflowResult(&report); err != nil {
		t.Fatalf("读取 workflow 结果失败: %v", err)
	}

	if strings.TrimSpace(report) == "" {
		t.Fatal("workflow 返回空报告")
	}

	fmt.Println(report)
}
