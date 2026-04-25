package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
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

	// 密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.SystemUser{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashed),
		Nickname:     req.Nickname,
		Status:       model.UserStatusActive,
		IsSuperAdmin: req.IsSuperAdmin,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// Get 获取单个用户
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// List 分页列表
func (s *UserService) List(ctx context.Context, page, pageSize int) (*model.PageResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	users, total, err := s.userRepo.List(ctx, page, pageSize)
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

// Update 更新用户
func (s *UserService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.UserResponse, error) {
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

// Delete 删除用户（软删除）
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

// UpdateMe 用户更新自己的资料
func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, req *model.UpdateMeRequest) (*model.UserResponse, error) {
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
func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 新密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashed)
	user.UpdatedTime = time.Now()
	return s.userRepo.Update(ctx, user)
}

func toUserResponse(u *model.SystemUser) *model.UserResponse {
	return &model.UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		Nickname:     u.Nickname,
		AvatarURL:    u.AvatarURL,
		Status:       u.Status,
		IsSuperAdmin: u.IsSuperAdmin,
		CreatedTime:  u.CreatedTime,
	}
}
