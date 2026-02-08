package task

import (
	"context"
	"errors"

	"asum/pkg/db"
	"asum/pkg/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, userID uint64, a *models.Task) error
	Update(ctx context.Context, a *models.Task) error
	Delete(ctx context.Context, id uint64) error
	HardDelete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*models.Task, error)
	// FindByTaskKey(ctx context.Context, apiKey string) (*models.Task, error)
	ExistsByTaskKey(ctx context.Context, taskKey string) (bool, error)

	// ValidateTaskKey(ctx context.Context, taskKey string) (*models.Task, error)

	AddUser(ctx context.Context, taskID, userID uint64, role string) error
	RemoveUser(ctx context.Context, taskID, userID uint64) error
	GetUsers(ctx context.Context, taskID uint64) ([]models.UserTask, error)
	GetUsersByTaskID(ctx context.Context, taskid int64) ([]models.UserTask, error)
}

type repository struct {
	db *db.DB
}

func NewRepository(db *db.DB) Repository {
	if err := db.AutoMigrate(&models.Task{}, &models.TaskItem{}); err != nil {
		panic(err)
	}
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, userID uint64, a *models.Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Task{}).Where("task_key = ?", a.TaskKey).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return models.ErrTaskAlreadyExists
		}

		if err := tx.Create(a).Error; err != nil {
			return err
		}

		userTask := models.UserTask{
			UserID: userID,
			TaskID: a.ID,
		}

		if err := tx.Create(&userTask).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repository) Update(ctx context.Context, a *models.Task) error {
	result := r.db.WithContext(ctx).
		Model(a).
		Select("name", "appkey", "status", "remark", "updated_at").
		Updates(a)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrTaskNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&models.Task{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrTaskNotFound
	}
	return nil
}

func (r *repository) HardDelete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", id).Delete(&models.UserTask{}).Error; err != nil {
			return err
		}

		result := tx.Unscoped().Where("id = ?", id).Delete(&models.Task{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return models.ErrTaskNotFound
		}

		return nil
	})
}

func (r *repository) FindByID(ctx context.Context, id uint64) (*models.Task, error) {
	query := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)

	var a models.Task
	if err := query.First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrTaskNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) FindByTaskKey(ctx context.Context, apiKey string) (*models.Task, error) {
	query := r.db.WithContext(ctx).Where("task_key = ? AND deleted_at IS NULL", apiKey)
	var a models.Task
	if err := query.First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrTaskNotFound
		}
		return nil, err
	}
	return &a, nil
}

// func (r *repository) FindByTaskid(ctx context.Context, taskid string, opts ...QueryOptions) (*models.Task, error) {
// 	opt := r.mergeOptions(opts)

// 	query := r.db.WithContext(ctx).Where("taskid = ? AND deleted_at IS NULL", taskid)
// 	query = r.applyPreloads(query, opt)

// 	var a models.Task
// 	if err := query.First(&a).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, models.ErrTaskNotFound
// 		}
// 		return nil, err
// 	}
// 	return &a, nil
// }

func (r *repository) ExistsByTaskKey(ctx context.Context, taskKey string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("task_key = ? AND deleted_at IS NULL", taskKey).
		Count(&count).Error

	return count > 0, err
}

func (r *repository) ValidateAppKey(ctx context.Context, taskKey string) (*models.Task, error) {
	var a models.Task
	err := r.db.WithContext(ctx).
		Where("task_key = ? AND status = ? AND deleted_at IS NULL",
			taskKey, models.StatusEnabled).
		First(&a).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrInvalidTaskKey
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) AddUser(ctx context.Context, taskID, userID uint64, role string) error {
	if role == "" {
		role = "member"
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserTask{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Count(&count).Error

	if err != nil {
		return err
	}
	if count > 0 {
		return r.db.WithContext(ctx).
			Model(&models.UserTask{}).
			Where("task_id = ? AND user_id = ?", taskID, userID).
			Update("role", role).Error
	}

	ua := &models.UserTask{
		TaskID: taskID,
		UserID: userID,
	}
	return r.db.WithContext(ctx).Create(ua).Error
}

func (r *repository) RemoveUser(ctx context.Context, taskID, userID uint64) error {
	return r.db.WithContext(ctx).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Delete(&models.UserTask{}).Error
}

func (r *repository) GetUsers(ctx context.Context, taskID uint64) ([]models.UserTask, error) {
	var users []models.UserTask
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Find(&users).Error

	return users, err
}

func (r *repository) GetUsersByTaskID(ctx context.Context, taskid int64) ([]models.UserTask, error) {
	app, err := r.FindByID(ctx, uint64(taskid))
	if err != nil {
		return nil, err
	}

	return r.GetUsers(ctx, app.ID)
}
