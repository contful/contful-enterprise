// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string      `json:"email" gorm:"type:varchar(255);unique_index;not null"`
	PasswordHash  string      `json:"-" gorm:"type:varchar(255);not null"`
	Nickname      string      `json:"nickname" gorm:"type:varchar(100)"`
	AvatarURL     string      `json:"avatar_url" gorm:"type:text"`
	Status        UserStatus  `json:"status" gorm:"type:user_status;not null;default:'active'"`
	IsSuperAdmin  bool        `json:"is_super_admin" gorm:"not null;default:false"`
	LastLoginTime *time.Time  `json:"last_login_time" gorm:"type:timestamptz"`
	LastLoginIP   *string     `json:"last_login_ip" gorm:"type:inet"`
	MFAEnabled    bool        `json:"mfa_enabled" gorm:"not null;default:false"`
	TOTPSecret    *string     `json:"-" gorm:"type:varchar(512)"` // AES-256-GCM 加密后存储
	RecoveryCodes *string     `json:"-" gorm:"type:text"`         // AES-256-GCM 加密后存储（JSON 数组）
	CreatedTime   time.Time   `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime   time.Time   `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	PasswordChangedTime *time.Time `json:"password_changed_time" gorm:"type:timestamptz"` // 密码最后修改时间（用于密码过期检查）
	DeletedAt   gorm.DeletedAt `json:"deleted_time" gorm:"column:deleted_time;index"` // 软删除时间戳
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
	CreatedTime time.Time     `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime time.Time     `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt   gorm.DeletedAt `json:"deleted_time" gorm:"column:deleted_time;index"`
}

func (SystemRole) TableName() string {
	return "system_roles"
}

// SystemUserRole 系统用户-角色关联（多对多）
type SystemUserRole struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_role"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_role"`
	CreatedTime time.Time `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
}

func (SystemUserRole) TableName() string {
	return "system_user_roles"
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
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"encrypted_password"` // RSA 加密后的密码（若提供则优先使用）
	TokenID           string `json:"token_id"`            // Anti-Replay Token ID
	RSAToken          string `json:"rsa_token"`           // Anti-Replay Token 值
}

// LoginResponse 登录响应
type LoginResponse struct {
	User              *UserResponse `json:"user"`
	AccessToken        string        `json:"access_token"`
	RefreshToken       string        `json:"refresh_token"`
	PasswordExpired    bool          `json:"password_expired,omitempty"`
	PasswordExpireDays *int          `json:"password_expire_days,omitempty"`
	MFASetupRequired   bool          `json:"mfa_setup_required,omitempty"` // MFA 强制开启但用户未设置
}

// MFARequiredResponse 登录步骤 1 — 需要 MFA 验证时的响应
type MFARequiredResponse struct {
	MFARequired bool   `json:"mfa_required"`
	MFAToken    string `json:"mfa_token"`
}

// RefreshResponse 刷新 Token 响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse 用户响应（脱敏）
type UserResponse struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Nickname          string     `json:"nickname"`
	AvatarURL         string     `json:"avatar_url,omitempty"`
	Status            UserStatus `json:"status"`
	IsSuperAdmin      bool       `json:"is_super_admin"`
	MFAEnabled        bool       `json:"mfa_enabled"`
	CreatedTime       time.Time  `json:"created_time"`
	PasswordChangedTime *time.Time `json:"password_changed_time,omitempty"` // 密码最后修改时间
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`              // 软删除时间（仅在包含已删除记录时返回）
}

// ============================================
// MFA DTO
// ============================================

// MFASetupResponse MFA 初始化响应（Setup 阶段）
type MFASetupResponse struct {
	TOTPSecret  string `json:"totp_secret"`  // Base32 明文 Secret（仅此次返回）
	OTPAuthURI  string `json:"otpauth_uri"`
	QRCodeURL   string `json:"qr_code_url"`
}

// MFAEnableRequest 启用 MFA 请求
type MFAEnableRequest struct {
	TOTPCode string `json:"totp_code" binding:"required,len=6"`
}

// MFAEnableResponse 启用 MFA 响应（含 Recovery Code）
type MFAEnableResponse struct {
	MFAEnabled    bool     `json:"mfa_enabled"`
	RecoveryCodes []string `json:"recovery_codes"` // 明文，仅此次可见
}

// MFADisableRequest 关闭 MFA 请求
type MFADisableRequest struct {
	TOTPCode string `json:"totp_code" binding:"required,len=6"`
}

// MFAVerifyRequest 登录 MFA 验证请求
type MFAVerifyRequest struct {
	MFAToken string `json:"mfa_token" binding:"required"`
	TOTPCode string `json:"totp_code" binding:"required,len=6"`
}

// MFARecoverRequest Recovery Code 恢复请求
type MFARecoverRequest struct {
	Email        string `json:"email" binding:"required,email"`
	RecoveryCode string `json:"recovery_code" binding:"required"`
}

// MFARecoverResponse Recovery Code 恢复响应
type MFARecoverResponse struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	RemainingCodes int    `json:"remaining_codes"`
}

// RecoveryCode 恢复码记录
type RecoveryCode struct {
	Code   string  `json:"code"`
	Used   bool    `json:"used"`
	UsedAt *string `json:"used_at"`
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

// UpdateMeRequest 用户更新自己的资料请求
type UpdateMeRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
}

// UpdatePasswordRequest 用户修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordRequest 管理员重置用户密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
