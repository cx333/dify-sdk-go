// Dify 示例服务 —— 演示 SDK 各 API Client 的完整用法。
// 基于 Fiber v3 + dig.Container，注册 Chat / Workflow / Knowledge 演示路由。
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/config"
	"github.com/wgl/dify-sdk/di"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 构建 DI 容器
	container, err := di.BuildContainer(cfg)
	if err != nil {
		log.Fatalf("构建容器失败: %v", err)
	}

	// 3. 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ServerHeader: "dify-examples",
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 健康检查
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 4. 注入依赖并注册演示路由
	if err := container.Invoke(func(httpClient *client.HTTPClient) {
		chatClient := client.NewChatClient(httpClient, cfg.DefaultUser)
		wfClient := client.NewWorkflowClient(httpClient, cfg.DefaultUser)
		kbClient := client.NewKnowledgeClient(httpClient)

		v1 := app.Group("/api/v1")

		// Chat 示例：发送对话消息
		v1.Post("/chat", func(c fiber.Ctx) error {
			var req client.ChatRequest
			if err := c.Bind().JSON(&req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			resp, err := chatClient.SendMessage(context.Background(), req)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(resp)
		})

		// Workflow 示例：执行工作流
		v1.Post("/workflow", func(c fiber.Ctx) error {
			var req client.WorkflowRunRequest
			if err := c.Bind().JSON(&req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			resp, err := wfClient.Run(context.Background(), req)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(resp)
		})

		// Knowledge 示例：列出知识库
		v1.Get("/datasets", func(c fiber.Ctx) error {
			resp, err := kbClient.ListDatasets(context.Background(), 1, 20)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(resp)
		})

		// App 信息示例：获取应用基本信息
		v1.Get("/info", func(c fiber.Ctx) error {
			info, err := chatClient.GetAppInfo(context.Background())
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(info)
		})
	}); err != nil {
		log.Fatalf("容器注入失败: %v", err)
	}

	// 5. 启动服务
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("示例服务启动，监听端口 :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("服务异常: %v", err)
	}
}
