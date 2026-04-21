package model

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// SystemUser 系统用户
type SystemUser struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string      `json:"email" gorm:"type:varchar(255);unique_index;not null"`
	PasswordHash string      `json:"-" gorm:"type:varchar(255);not null"`
	Nickname     string      `json:"nickname" gorm:"type:varchar(100)"`
	AvatarURL    string      `json:"avatar_url" gorm:"type:text"`
	Status       UserStatus  `json:"status" gorm:"type:user_status;not null;default:'active'"`
	IsSuperAdmin bool        `json:"is_super_admin" gorm:"not null;default:false"`
	SiteID       *uuid.UUID  `json:"site_id" gorm:"type:uuid"`
	LastLoginTime *time.Time `json:"last_login_time" gorm:"type:timestamptz"`
	LastLoginIP  *string     `json:"last_login_ip" gorm:"type:inet"`
	CreatedTime  time.Time   `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime  time.Time   `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	DeletedTime  *time.Time  `json:"deleted_time" gorm:"type:timestamptz"`
}

func (SystemUser) TableName() string {
	return "system_users"
}

// SystemRole 系统角色
type SystemRole struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);unique_index;not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsSystem    bool      `json:"is_system" gorm:"not null;default:false"`
	Permissions []string  `json:"permissions" gorm:"type:jsonb;serializer:json"`
	CreatedTime time.Time `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime time.Time `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	DeletedTime *time.Time `json:"deleted_time" gorm:"type:timestamptz"`
}

func (SystemRole) TableName() string {
	return "system_roles"
}

// SiteUser 站点用户关联
type SiteUser struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID           uuid.UUID  `json:"site_id" gorm:"type:uuid;not null;index"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID           uuid.UUID  `json:"role_id" gorm:"type:uuid;not null;index"`
	Status           UserStatus `json:"status" gorm:"type:user_status;not null;default:'active'"`
	ExtraPermissions []string   `json:"extra_permissions" gorm:"type:jsonb;serializer:json"`
	CreatedTime      time.Time  `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime      time.Time  `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	DeletedTime      *time.Time `json:"deleted_time" gorm:"type:timestamptz"`
}

func (SiteUser) TableName() string {
	return "site_users"
}

// SiteRole 站点角色
type SiteRole struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID             uuid.UUID  `json:"site_id" gorm:"type:uuid;not null;index"`
	Name               string     `json:"name" gorm:"type:varchar(100);not null"`
	Description        string     `json:"description" gorm:"type:text"`
	IsSystem           bool       `json:"is_system" gorm:"not null;default:false"`
	Permissions        []string   `json:"permissions" gorm:"type:jsonb;serializer:json"`
	ContentPermissions []string   `json:"content_permissions" gorm:"type:jsonb;serializer:json"`
	ChannelPermissions []string   `json:"channel_permissions" gorm:"type:jsonb;serializer:json"`
	SortOrder          int        `json:"sort_order" gorm:"not null;default:0"`
	CreatedTime        time.Time  `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime        time.Time  `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	DeletedTime        *time.Time `json:"deleted_time" gorm:"type:timestamptz"`
}

func (SiteRole) TableName() string {
	return "site_roles"
}

// ============================================
// DTO / Request/Response
// ============================================

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

// RefreshResponse 刷新 Token 响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse 用户响应（脱敏）
type UserResponse struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Nickname     string     `json:"nickname"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	Status       UserStatus `json:"status"`
	IsSuperAdmin bool      `json:"is_super_admin"`
	CreatedTime  time.Time  `json:"created_time"`
}

// UserWithSites 用户及关联站点
type UserWithSites struct {
	UserResponse
	Sites []SiteBasic `json:"sites"`
}

// SiteBasic 站点基本信息
type SiteBasic struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// CreateUserRequest 管理员创建用户请求
type CreateUserRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Nickname    string `json:"nickname"`
	IsSuperAdmin bool  `json:"is_super_admin"`
}

// UpdateUserRequest 管理员更新用户请求
type UpdateUserRequest struct {
	Nickname      *string     `json:"nickname"`
	Status        *UserStatus `json:"status"`
	IsSuperAdmin  *bool       `json:"is_super_admin"`
}
