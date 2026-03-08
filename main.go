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
	"time"
)

func main() {
	topic := "2026年AI Agent市场观察"
	if len(os.Args) > 1 && os.Args[1] != "" {
		topic = os.Args[1]
	}

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("无法创建 Temporal Client", err)
	}
	defer c.Close()

	// 开发态下直接内嵌启动一个 worker，避免单独起 ./worker 时 main 卡住等待。
	we := worker.New(c, namespace.TaskQueueName, worker.Options{})
	we.RegisterWorkflow(workflow.ChronicleResearchWorkflow)
	we.RegisterActivity(activity.RewriteResearchQueryWithChronicle)
	we.RegisterActivity(activity.RetrieveResearchDataWithChronicle)
	if err := we.Start(); err != nil {
		log.Fatalln("无法启动内嵌 Worker", err)
	}
	defer we.Stop()

	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("chronicle-research-%d", time.Now().UnixNano()),
		TaskQueue: namespace.TaskQueueName,
	}

	run, err := c.ExecuteWorkflow(
		context.Background(),
		options,
		workflow.ChronicleResearchWorkflow,
		topic,
	)
	if err != nil {
		log.Fatalln("无法启动 Workflow", err)
	}

	var article string
	waitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := run.Get(waitCtx, &article); err != nil {
		log.Fatalln("Workflow 执行失败", err)
	}

	fmt.Println("==== Chronicle 检索结果文档 ====")
	fmt.Println(article)
}
