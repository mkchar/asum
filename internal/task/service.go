package task

import (
	"asum/internal/user"
	"asum/pkg/models"
	"asum/pkg/utils"
	"context"
	"strings"
)

type Service interface {
	CreateTask(c context.Context, req *CreateTaskReq, userID uint64) error
}

type service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) CreateTask(c context.Context, req *CreateTaskReq, userID uint64) error {
	if err := s.repo.Create(c, userID, &models.Task{
		Name:    strings.Trim(req.Name, " "),
		Remark:  strings.Trim(req.Remark, " "),
		Status:  models.StatusEnabled,
		TaskKey: utils.NewUUID(),
	}); err != nil {
		return err
	}
	_ = s.userRepo.AddLog(c, &models.UserLog{
		UserID:    userID,
		IP:        utils.GetRemoteIP(c),
		UserAgent: utils.GetUserAgent(c),
		Type:      models.LogTypeCreateTask,
	})

	return nil
}
