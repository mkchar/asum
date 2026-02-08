package user

import (
	"asum/pkg/db"
	"asum/pkg/models"
	"asum/pkg/rdb"
	"asum/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, u *models.User) error
	Update(ctx context.Context, u *models.User) error
	UserActiveAndInit(ctx context.Context, id uint64, level models.Level) error
	Delete(ctx context.Context, id uint64) error
	HardDelete(ctx context.Context, id uint64) error

	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	AddLog(ctx context.Context, log *models.UserLog) error
	GetLogs(ctx context.Context, userID uint64, limit int) ([]models.UserLog, error)
	AddTask(ctx context.Context, userTask *models.UserTask) error
	RemoveTask(ctx context.Context, userID, taskID uint64) error
	GetTasks(ctx context.Context, userID uint64) ([]models.UserTask, error)

	GetQuotaByKey(ctx context.Context, key string) int64
	UpdateLoginTime(ctx context.Context, id uint64) error
}

type repository struct {
	db  *db.DB
	rdb *rdb.Client
}

func NewRepository(db *db.DB, rdb *rdb.Client) Repository {
	if err := db.AutoMigrate(&models.User{}, &models.UserLog{}, &models.UserTask{}); err != nil {
		panic(err)
	}
	return &repository{db: db, rdb: rdb}
}

func (r *repository) GetQuotaByKey(ctx context.Context, key string) int64 {
	var quota int

	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.quota").
		Joins("INNER JOIN user_tasks ON user_tasks.user_id = users.id").
		Joins("INNER JOIN tasks ON tasks.id = user_tasks.task_id").
		Where("tasks.task_key = ?", key).
		Where("users.deleted_at IS NULL").
		Where("tasks.deleted_at IS NULL").
		Limit(1).
		Scan(&quota).Error

	if err != nil {
		return 0
	}

	return int64(quota)
}

func (r *repository) Create(ctx context.Context, u *models.User) error {
	exists, err := r.ExistsByEmail(ctx, u.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repository) UserActiveAndInit(ctx context.Context, id uint64, lev models.Level) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status": models.StatusActive,
		}).Error; err != nil {
			return err
		}
		taskKey := utils.NewUUID()
		newTask := models.Task{
			Name:    "默认空间",
			TaskKey: taskKey,
			Status:  models.StatusEnabled,
			Remark:  "用户激活自动初始化",
		}

		if err := tx.Create(&newTask).Error; err != nil {
			return err
		}
		newUserTask := models.UserTask{
			UserID: id,
			TaskID: newTask.ID,
		}

		if err := tx.Create(&newUserTask).Error; err != nil {
			return err
		}
		newLog := models.UserLog{
			UserID:    id,
			Type:      models.LogTypeActive,
			IP:        utils.GetRemoteIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Extra:     "User activated and default workspace created",
		}

		if err := tx.Create(&newLog).Error; err != nil {
			return err
		}
		cache, _ := json.Marshal(models.ApiCache{UserLevel: lev, Quota: 0})
		if err := r.rdb.Set(ctx, fmt.Sprintf("apiKey:%s", taskKey), string(cache), 0).Err(); err != nil {
			return err
		}
		return nil
	})
}

func (r *repository) Update(ctx context.Context, u *models.User) error {
	result := r.db.WithContext(ctx).
		Model(u).
		Select("name", "email", "password", "level", "status", "updated_at").
		Updates(u)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&models.User{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *repository) HardDelete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&models.UserLog{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&models.UserTask{}).Error; err != nil {
			return err
		}

		result := tx.Unscoped().Where("id = ?", id).Delete(&models.User{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrUserNotFound
		}

		return nil
	})
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email)

	var u models.User
	if err := query.First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error

	return count > 0, err
}

func (r *repository) AddLog(ctx context.Context, log *models.UserLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) GetLogs(ctx context.Context, userID uint64, limit int) ([]models.UserLog, error) {
	if limit <= 0 {
		limit = 10
	}

	var logs []models.UserLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}

func (r *repository) AddTask(ctx context.Context, userApp *models.UserTask) error {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserTask{}).
		Where("user_id = ? AND task_id = ?", userApp.UserID, userApp.TaskID).
		Count(&count).Error

	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(userApp).Error
}

func (r *repository) RemoveTask(ctx context.Context, userID, taskID uint64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND task_id = ?", userID, taskID).
		Delete(models.UserTask{}).Error
}

func (r *repository) GetTasks(ctx context.Context, userID uint64) ([]models.UserTask, error) {
	var tasks []models.UserTask
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error

	return tasks, err
}

func (r *repository) UpdateLoginTime(ctx context.Context, id uint64) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("login_at", now).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.db.WithContext(ctx).
		Select("level", "quota").
		First(&user, id).Error; err != nil {
		return err
	}
	var taskKeys []string
	if err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Joins("JOIN user_tasks ON user_tasks.task_id = tasks.id").
		Where("user_tasks.user_id = ?", id).
		Pluck("tasks.task_key", &taskKeys).Error; err != nil {
		return err
	}

	if len(taskKeys) == 0 {
		return nil
	}

	cacheData := models.ApiCache{
		UserLevel: models.Level(user.Level),
		Quota:     user.Quota,
	}
	cacheBytes, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	cacheStr := string(cacheBytes)

	pipe := r.rdb.Pipeline()
	for _, key := range taskKeys {
		redisKey := fmt.Sprintf("apiKey:%s", key)
		pipe.Set(ctx, redisKey, cacheStr, 0)
	}

	_, err = pipe.Exec(ctx)
	return err
}
