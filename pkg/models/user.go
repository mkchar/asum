package models

import "time"

type Level int

const (
	LevelBasic Level = iota
	LevelPlus
	LevelPremium
	LevelTop
)

func (l Level) String() string {
	switch l {
	case LevelBasic:
		return "basic"
	case LevelPlus:
		return "plus"
	case LevelPremium:
		return "premium"
	case LevelTop:
		return "top"
	default:
		return "unknown"
	}
}

type ApiCache struct {
	UserLevel Level `json:"userLevel"`
	Quota     int   `json:"quota"`
}

type UserStatus int

const (
	StatusInactive UserStatus = iota
	StatusActive
	StatusBanned
)

type LogType int

const (
	LogTypeLogin LogType = iota
	LogTypeActive
	LogTypeSend
	LogTypeLogout
	LogTypeResetPassword
	LogTypeCreateTask
)

type User struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:100;not null" json:"name"`
	Email     string     `gorm:"size:255;uniqueIndex;not null;default:''" json:"email"`
	Password  string     `gorm:"size:255;not null" json:"-"`
	Level     Level      `gorm:"default:0" json:"level"`
	Quota     int        `gorm:"default:0" json:"quota"`
	Status    UserStatus `gorm:"default:0" json:"status"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
	LoginAt   *time.Time `json:"loginAt,omitempty"`
	Logs      []UserLog  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"logs,omitempty"`
	Tasks     []UserTask `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type UserLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"userId"`
	Type      LogType   `gorm:"not null" json:"type"`
	IP        string    `gorm:"size:45" json:"ip,omitempty"`
	UserAgent string    `gorm:"size:500" json:"userAgent,omitempty"`
	Extra     string    `gorm:"type:text" json:"extra,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserLog) TableName() string {
	return "user_logs"
}

type UserTask struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"userId"`
	TaskID    uint64    `gorm:"index;not null" json:"taskId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserTask) TableName() string {
	return "user_tasks"
}
