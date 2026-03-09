package main

import (
	"ai.openclaw.usecase/namespace"
	"ai.openclaw.usecase/workflow"
	"ai.openclaw.usecase/workflow/activity"
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	input := activity.ResearchReportInput{
		Topic:     "新能源行业过去一周研报",
		TimeRange: "过去一周",
		Industry:  "新能源",
		Style:     "脱水研报",
	}

	if len(os.Args) > 1 && os.Args[1] != "" {
		input.Topic = os.Args[1]
	}
	if len(os.Args) > 2 && os.Args[2] != "" {
		input.TimeRange = os.Args[2]
	}
	if len(os.Args) > 3 && os.Args[3] != "" {
		input.Industry = os.Args[3]
	}
	if len(os.Args) > 4 && os.Args[4] != "" {
		input.Style = os.Args[4]
	}
	if len(os.Args) > 5 && os.Args[5] != "" {
		if n, err := strconv.Atoi(os.Args[5]); err == nil {
			input.MaxSources = n
		}
	}

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("无法创建 Temporal Client", err)
	}
	defer c.Close()

	// 开发态下直接内嵌启动一个 worker，避免单独起 ./worker 时 main 卡住等待。
	we := worker.New(c, namespace.TaskQueueName, worker.Options{})
	we.RegisterWorkflow(workflow.ResearchReportWorkflow)
	we.RegisterActivity(activity.FetchResearchReports)
	we.RegisterActivity(activity.CleanResearchData)
	we.RegisterActivity(activity.VerifySources)
	we.RegisterActivity(activity.WriteCondensedResearchReport)
	if err := we.Start(); err != nil {
		log.Fatalln("无法启动内嵌 Worker", err)
	}
	defer we.Stop()

	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("research-report-%d", time.Now().UnixNano()),
		TaskQueue: namespace.TaskQueueName,
	}

	run, err := c.ExecuteWorkflow(
		context.Background(),
		options,
		workflow.ResearchReportWorkflow,
		input,
	)
	if err != nil {
		log.Fatalln("无法启动 Workflow", err)
	}

	var report string
	waitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := run.Get(waitCtx, &report); err != nil {
		log.Fatalln("Workflow 执行失败", err)
	}

	fmt.Println("==== 研报输出 ====")
	fmt.Println(report)
}
