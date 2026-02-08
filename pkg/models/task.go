package models

import (
	"errors"
	"time"
)

type TaskStatus int

const (
	StatusDisabled TaskStatus = iota
	StatusEnabled
)

type Task struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:100;not null" json:"name"`
	TaskKey   string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	Status    TaskStatus `gorm:"default:1" json:"status"`
	Remark    string     `gorm:"size:500" json:"remark,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Users []UserTask `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"users,omitempty"`
	Items []TaskItem `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task_items,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}

type TaskItem struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	TaskID   int
	TaskType int `gorm:"default:1" json:"task_type"`
}

func (TaskItem) TableName() string {
	return "task_items"
}

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrInvalidTaskKey    = errors.New("invalid task key")
)
