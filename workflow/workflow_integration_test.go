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

	topic := os.Getenv("OPENCLAW_TEST_TOPIC")
	if strings.TrimSpace(topic) == "" {
		topic = "写一篇关于新能源行业过去一周有关的研报的脱水研报"
	}

	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.SetTestTimeout(30 * time.Minute)
	env.RegisterWorkflow(ResearchReportWorkflow)
	env.RegisterActivity(activity.FetchResearchReports)
	env.RegisterActivity(activity.CleanResearchData)
	env.RegisterActivity(activity.WriteCondensedResearchReport)

	env.ExecuteWorkflow(ResearchReportWorkflow, topic)

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
