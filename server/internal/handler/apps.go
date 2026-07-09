package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/pkg/response"
	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/store"
)

// AppsHandler 应用列表相关请求处理器。
type AppsHandler struct {
	store *store.MemoryStore
	log   *logger.Logger
}

// NewAppsHandler 创建 AppsHandler。
func NewAppsHandler(store *store.MemoryStore, log *logger.Logger) *AppsHandler {
	return &AppsHandler{store: store, log: log}
}

// AppItem 应用列表中的单个应用项。
type AppItem struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Mode        string   `json:"mode"`
	ModeLabel   string   `json:"mode_label"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// List 返回所有已发现的应用列表。
// GET /api/v1/apps
func (h *AppsHandler) List(c fiber.Ctx) error {
	apps := h.store.All()
	items := make([]AppItem, 0, len(apps))
	for _, app := range apps {
		var idx int
		if app.ID != "" {
			idx, _ = strconv.Atoi(app.ID)
		}
		items = append(items, AppItem{
			Index:       idx,
			Name:        app.Name,
			Mode:        app.Mode,
			ModeLabel:   client.AppMode(app.Mode).Label(),
			Description: app.Description,
			Tags:        app.Tags,
		})
	}
	return response.OK(c, items)
}
