package task

import (
	"asum/pkg/engine"
	"asum/pkg/errorx"
	"asum/pkg/utils"
	"errors"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(appSvc Service) *Handler {
	return &Handler{service: appSvc}
}

type CreateTaskReq struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

// CreateTask 创建新任务
// @Summary 创建新任务
// @Description 为当前登录用户创建一个新的任务。
// @Tags Task
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateTaskReq true "创建任务参数"
// @Success 200 {object} engine.Response "创建成功"
// @Failure 400 {object} engine.Response "参数错误 (如名称为空)"
// @Failure 401 {object} engine.Response "未授权 (Token 缺失或无效)"
// @Failure 403 {object} engine.Response "禁止操作 (如任务已存在)"
// @Router /app/task [post]
func (h *Handler) CreateTask(c *engine.Ctx) error {
	var req CreateTaskReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}
	if req.Name == "" {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	currentUserID := utils.GetUserID(c)
	if currentUserID == 0 {
		return errors.New("用户未登录")
	}

	err := h.service.CreateTask(c.StdCtx, &req, currentUserID)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(nil)
}

func (h *Handler) UpdateTask(c *engine.Ctx) error {
	// return
	return nil
}

func (h *Handler) GetTask(c *engine.Ctx) error {
	// return
	return nil
}

func (h *Handler) DeleteTask(c *engine.Ctx) error {
	// return
	return nil
}

func (h *Handler) ListTask(c *engine.Ctx) error {
	return nil
}
