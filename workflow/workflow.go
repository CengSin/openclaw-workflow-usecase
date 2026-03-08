package workflow

import (
	"ai.openclaw.usecase/workflow/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func ResearchReportWorkflow(ctx workflow.Context, topic string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var rawResearch string
	var cleanedResearch string
	var finalReport string
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow: 开始执行脱水研报编排", "topic", topic)

	if err := workflow.ExecuteActivity(ctx, activity.FetchResearchReports, topic).Get(ctx, &rawResearch); err != nil {
		logger.Error("Workflow: 研报检索失败", "error", err)
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, activity.CleanResearchData, rawResearch).Get(ctx, &cleanedResearch); err != nil {
		logger.Error("Workflow: 数据整理清洗失败", "error", err)
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, activity.WriteCondensedResearchReport, cleanedResearch).Get(ctx, &finalReport); err != nil {
		logger.Error("Workflow: 脱水研报撰写失败", "error", err)
		return "", err
	}

	logger.Info(
		"Workflow 完成",
		"rawLength", len(rawResearch),
		"cleanedLength", len(cleanedResearch),
		"reportLength", len(finalReport),
	)
	return finalReport, nil
}
