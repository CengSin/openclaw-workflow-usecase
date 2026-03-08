package workflow

import (
	"ai.openclaw.usecase/workflow/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func ChronicleResearchWorkflow(ctx workflow.Context, topic string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    20 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var rewrittenQuery string
	var result activity.ResearchRetrievalResult
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow: 先改写查询，再检索研报数据", "topic", topic)

	if err := workflow.ExecuteActivity(ctx, activity.RewriteResearchQueryWithChronicle, topic).Get(ctx, &rewrittenQuery); err != nil {
		logger.Error("Workflow: 查询改写失败", "error", err)
		return "", err
	}

	request := activity.ResearchRetrievalRequest{
		Topic:          topic,
		RewrittenQuery: rewrittenQuery,
		RequiredSkills: []string{
			activity.ChronicleSkillResearchAIPicker,
			activity.ChronicleSkillQVeris,
		},
	}

	if err := workflow.ExecuteActivity(ctx, activity.RetrieveResearchDataWithChronicle, request).Get(ctx, &result); err != nil {
		logger.Error("Workflow: 研报数据检索失败", "error", err)
		return "", err
	}

	logger.Info("Workflow 完成", "query", rewrittenQuery, "length", len(result.Document), "skills", result.SkillChain)
	return result.Document, nil
}
