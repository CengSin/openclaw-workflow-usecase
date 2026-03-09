package workflow

import (
	"ai.openclaw.usecase/workflow/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func ResearchReportWorkflow(ctx workflow.Context, input activity.ResearchReportInput) (string, error) {
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
	var verifiedResearch string
	var finalReport string
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow: 开始执行研报编排",
		"topic", input.Topic,
		"timeRange", input.TimeRange,
		"industry", input.Industry,
		"style", input.Style,
		"maxSources", input.MaxSources,
	)

	if err := workflow.ExecuteActivity(ctx, activity.FetchResearchReports, input).Get(ctx, &rawResearch); err != nil {
		logger.Error("Workflow: 研报检索失败", "error", err)
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, activity.CleanResearchData, rawResearch).Get(ctx, &cleanedResearch); err != nil {
		logger.Error("Workflow: 数据整理清洗失败", "error", err)
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, activity.VerifySources, cleanedResearch).Get(ctx, &verifiedResearch); err != nil {
		logger.Error("Workflow: 溯源校验失败", "error", err)
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, activity.WriteCondensedResearchReport, input, verifiedResearch).Get(ctx, &finalReport); err != nil {
		logger.Error("Workflow: 研报撰写失败", "error", err)
		return "", err
	}

	logger.Info(
		"Workflow 完成",
		"rawLength", len(rawResearch),
		"cleanedLength", len(cleanedResearch),
		"verifiedLength", len(verifiedResearch),
		"reportLength", len(finalReport),
	)
	return finalReport, nil
}
