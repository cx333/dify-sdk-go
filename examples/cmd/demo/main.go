// Dify SDK CLI 演示 —— 直接调用本地 Dify 实例的各个 API。
// 运行：cd examples && go run ./cmd/demo/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// 加载 .env 配置
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("加载配置失败: %v\n提示: 确保 examples/.env 文件存在且配置正确", err)
	}

	// 使用第一个 API Key 创建 HTTPClient
	key := cfg.APIKeys[0]
	httpClient := client.NewHTTPClient(cfg.BaseURL, key, cfg.Timeout, client.DefaultRetryConfig())

	ctx := context.Background()
	cmd := os.Args[1]

	switch cmd {
	case "info":
		// 获取应用基本信息
		c := client.NewChatClient(httpClient)
		info, err := c.GetAppInfo(ctx)
		if err != nil {
			log.Fatalf("获取应用信息失败: %v", err)
		}
		fmt.Printf("应用名称: %s\n", info.Name)
		fmt.Printf("应用模式: %s\n", info.Mode)
		fmt.Printf("描述: %s\n", info.Description)
		fmt.Printf("标签: %s\n", strings.Join(info.Tags, ", "))

	case "chat":
		// 发送对话消息
		if len(os.Args) < 4 {
			fmt.Println("用法: demo chat <user> <message>")
			return
		}
		user := os.Args[2]
		query := strings.Join(os.Args[3:], " ")

		c := client.NewChatClient(httpClient)
		resp, err := c.SendMessage(ctx, client.ChatRequest{
			Query:  query,
			User:   user,
			Inputs: map[string]interface{}{},
		})
		if err != nil {
			log.Fatalf("发送消息失败: %v", err)
		}
		fmt.Printf("回复: %s\n", resp.Answer)
		fmt.Printf("会话ID: %s\n", resp.ConversationID)
		fmt.Printf("消息ID: %s\n", resp.MessageID)

	case "conversations":
		// 获取会话列表
		if len(os.Args) < 3 {
			fmt.Println("用法: demo conversations <user>")
			return
		}
		user := os.Args[2]
		c := client.NewChatClient(httpClient)
		resp, err := c.GetConversations(ctx, user, "", 20)
		if err != nil {
			log.Fatalf("获取会话列表失败: %v", err)
		}
		fmt.Printf("共 %d 个会话:\n", len(resp.Data))
		for _, conv := range resp.Data {
			fmt.Printf("  [%s] %s (更新于 %d)\n", conv.ID[:8], conv.Name, conv.UpdatedAt)
		}

	case "workflow":
		// 执行工作流
		if len(os.Args) < 3 {
			fmt.Println("用法: demo workflow <user> [key=value ...]")
			return
		}
		user := os.Args[2]
		inputs := map[string]interface{}{}
		for _, arg := range os.Args[3:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				inputs[parts[0]] = parts[1]
			}
		}

		c := client.NewWorkflowClient(httpClient)
		resp, err := c.Run(ctx, client.WorkflowRunRequest{
			Inputs: inputs,
			User:   user,
		})
		if err != nil {
			log.Fatalf("执行工作流失败: %v", err)
		}
		fmt.Printf("状态: %s\n", resp.Data.Status)
		fmt.Printf("运行ID: %s\n", resp.WorkflowRunID)
		fmt.Printf("耗时: %.2fs\n", resp.Data.ElapsedTime)
		if resp.Data.Outputs != nil {
			for k, v := range resp.Data.Outputs {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}

	case "datasets":
		// 列出知识库
		c := client.NewKnowledgeClient(httpClient)
		resp, err := c.ListDatasets(ctx, 1, 20)
		if err != nil {
			log.Fatalf("获取知识库列表失败: %v", err)
		}
		fmt.Printf("共 %d 个知识库:\n", len(resp.Data))
		for _, ds := range resp.Data {
			fmt.Printf("  [%s] %s (文档: %d, 字数: %d)\n",
				ds.ID[:8], ds.Name, ds.DocumentCount, ds.WordCount)
		}

	case "search":
		// 知识库检索
		if len(os.Args) < 4 {
			fmt.Println("用法: demo search <dataset_id> <query>")
			return
		}
		datasetID := os.Args[2]
		query := strings.Join(os.Args[3:], " ")

		c := client.NewKnowledgeClient(httpClient)
		resp, err := c.RetrieveSegments(ctx, datasetID, client.RetrieveRequest{
			Query: query,
		})
		if err != nil {
			log.Fatalf("检索失败: %v", err)
		}
		fmt.Printf("查询: %s → 返回 %d 条结果:\n", query, len(resp.Records))
		for i, r := range resp.Records {
			fmt.Printf("  %d. [%.4f] %s\n", i+1, r.Score, truncate(r.Segment.Content, 100))
		}

	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Dify SDK 演示 CLI

用法:
  demo info                          获取应用基本信息
  demo chat <user> <message>         发送对话消息
  demo conversations <user>          获取会话列表
  demo workflow <user> [k=v ...]     执行工作流
  demo datasets                      列出知识库
  demo search <dataset_id> <query>   知识库检索

环境要求:
  确保 examples/.env 文件已配置 DIFY_BASE_URL 和 DIFY_API_KEYS`)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n]) + "..."
	}
	return s
}
