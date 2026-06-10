// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/contful/contful/admin/pkg/uid"
	"github.com/contful/contful/admin/internal/audit"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrWeakPassword      = errors.New("password must be at least 8 characters with uppercase, lowercase and numbers")
)

// isPasswordStrong 检查密码强度：至少8位，包含大小写字母与数字
func isPasswordStrong(pwd string) bool {
	if len(pwd) < 8 {
		return false
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	for _, c := range pwd {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	return hasLower && hasUpper && hasDigit
}

type UserService struct {
	userRepo        *repository.UserRepository
	storageProvider storage.StorageProvider
}

func NewUserService(userRepo *repository.UserRepository, storageProvider storage.StorageProvider) *UserService {
	return &UserService{userRepo: userRepo, storageProvider: storageProvider}
}

// Create 创建用户（管理员操作）
func (s *UserService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	// 检查邮箱是否已存在
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// 密码强度检查
	if !isPasswordStrong(req.Password) {
		return nil, ErrWeakPassword
	}

	// 密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &model.SystemUser{
		ID:                uid.New(),
		Email:             req.Email,
		PasswordHash:      string(hashed),
		Nickname:          req.Nickname,
		Status:            model.UserStatusActive,
		IsSuperAdmin:      req.IsSuperAdmin,
		CreatedTime:        now,
		UpdatedTime:        now,
		PasswordChangedTime: &now, // 创建用户时设置密码修改时间
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// Get 获取单个用户
func (s *UserService) Get(ctx context.Context, id uid.UID) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, id uid.UID, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.IsSuperAdmin != nil {
		user.IsSuperAdmin = *req.IsSuperAdmin
	}

	user.UpdatedTime = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// Delete 删除用户（支持软删除和永久删除）
func (s *UserService) Delete(ctx context.Context, id uid.UID, permanent bool) error {
	if permanent {
		return s.userRepo.PermanentDelete(ctx, id)
	}
	return s.userRepo.Delete(ctx, id)
}

// Restore 恢复软删除的用户
func (s *UserService) Restore(ctx context.Context, id uid.UID) error {
	return s.userRepo.Restore(ctx, id)
}

// List 分页列表（可包含已删除记录）
func (s *UserService) List(ctx context.Context, page, pageSize int, includeDeleted bool) (*model.PageResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	users, total, err := s.userRepo.List(ctx, page, pageSize, includeDeleted)
	if err != nil {
		return nil, err
	}

	items := make([]*model.UserResponse, len(users))
	for i := range users {
		items[i] = toUserResponse(&users[i])
	}

	return &model.PageResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateMe 用户更新自己的资料
func (s *UserService) UpdateMe(ctx context.Context, userID uid.UID, req *model.UpdateMeRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	user.UpdatedTime = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// UpdatePassword 用户修改密码
func (s *UserService) UpdatePassword(ctx context.Context, userID uid.UID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 新密码强度检查
	if !isPasswordStrong(newPassword) {
		return ErrWeakPassword
	}

	// 新密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user.PasswordHash = string(hashed)
	user.PasswordChangedTime = &now // 更新密码修改时间
	user.UpdatedTime = now
	return s.userRepo.Update(ctx, user)
}

// ResetPassword 管理员重置用户密码（不需要旧密码）
func (s *UserService) ResetPassword(ctx context.Context, id uid.UID, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 新密码强度检查
	if !isPasswordStrong(newPassword) {
		return ErrWeakPassword
	}

	// 新密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user.PasswordHash = string(hashed)
	user.PasswordChangedTime = &now // 更新密码修改时间
	user.UpdatedTime = now
	return s.userRepo.Update(ctx, user)
}

func toUserResponse(u *model.SystemUser) *model.UserResponse {
	resp := &model.UserResponse{
		ID:                u.ID,
		Email:             u.Email,
		Nickname:          u.Nickname,
		AvatarURL:         u.AvatarURL,
		Status:            u.Status,
		IsSuperAdmin:      u.IsSuperAdmin,
		MFAEnabled:        u.MFAEnabled,
		CreatedTime:       u.CreatedTime,
		PasswordChangedTime: u.PasswordChangedTime,
	}
	if u.DeletedAt.Valid {
		resp.DeletedAt = &u.DeletedAt.Time
	}
	return resp
}

// VerifyUserResult 用户验签结果
type VerifyUserResult struct {
	Valid     bool   `json:"valid"`
	Algorithm string `json:"algorithm"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
	Reason    string `json:"reason,omitempty"`
}

// SignUser 对用户数据重新签名
func (s *UserService) SignUser(ctx context.Context, id uid.UID) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	signer := audit.GetSigner(ctx)
	if signer == nil {
		return errors.New("数据签名器未启用")
	}

	payload := CanonicalSystemUserPayload(user.Email, user.PasswordHash, user.Nickname, string(user.Status), user.IsSuperAdmin)
	sig, err := signer.Sign("system_users", user.ID, payload)
	if err != nil {
		return fmt.Errorf("签名失败: %w", err)
	}

	user.DataSignature = sig
	return s.userRepo.Update(ctx, user)
}

// VerifyUser 验签用户数据
func (s *UserService) VerifyUser(ctx context.Context, id uid.UID) (*VerifyUserResult, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	signer := audit.GetSigner(ctx)
	if signer == nil {
		return &VerifyUserResult{Valid: false, Reason: "数据签名器未启用"}, nil
	}

	payload := CanonicalSystemUserPayload(user.Email, user.PasswordHash, user.Nickname, string(user.Status), user.IsSuperAdmin)

	valid, err := signer.Verify("system_users", user.ID, payload, user.DataSignature)
	if err != nil {
		return nil, fmt.Errorf("验签失败: %w", err)
	}

	result := &VerifyUserResult{
		Valid:     valid,
		Algorithm: signer.Algorithm(),
		Signature: user.DataSignature,
		Payload:   payload,
	}
	if !valid {
		result.Reason = "数据签名不匹配，数据可能已被篡改"
	}
	return result, nil
}

// UploadAvatar 上传用户头像（通过 StorageProvider，支持本地/对象存储）
func (s *UserService) UploadAvatar(ctx context.Context, userID uid.UID, file io.Reader, header *multipart.FileHeader) (string, error) {
	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read avatar file: %w", err)
	}

	// 构造存储 Key：avatars/{userID}.{ext}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	storageKey := filepath.Join("avatars", userID.String()+ext)

	// 通过 StorageProvider 上传
	_, err = s.storageProvider.Upload(ctx, storageKey, bytes.NewReader(data), int64(len(data)), &storage.WriteOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar: %w", err)
	}

	// 获取访问 URL
	avatarURL, err := s.storageProvider.URL(ctx, storageKey, 0)
	if err != nil {
		return "", fmt.Errorf("get avatar url: %w", err)
	}

	// 更新数据库中的 avatar_url
	if err := s.userRepo.UpdateAvatarURL(ctx, userID, avatarURL); err != nil {
		return "", fmt.Errorf("update avatar url: %w", err)
	}

	return avatarURL, nil
}

