package Usecases

import (
	"fmt"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
)

type UserUseCase interface {
	Register(req Domain.RegisterRequest) (*Domain.User, error)
	Login(req Domain.LoginRequest) (*Domain.LoginResponse, error)
	RefreshToken(userID string) (string, error)
	GetUserByID(id string) (*Domain.User, error)
	GetCurrentUser(id string) (*Domain.User, error)
	UpdateUser(id string, req Domain.UpdateUserRequest) (*Domain.User, error)
	UpdateStatus(id string, status Domain.UserStatus) error
	DeleteUser(id string) error
}

type userUseCase struct {
	userRepo   Domain.UserRepository
	jwtService Infrastructure.JWTService
}

func NewUserUseCase(userRepo Domain.UserRepository, jwtService Infrastructure.JWTService) UserUseCase {
	return &userUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (uc *userUseCase) Register(req Domain.RegisterRequest) (*Domain.User, error) {
	// Check if user already exists with phone
	existingUser, err := uc.userRepo.FindByPhone(req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with phone %s already exists", req.Phone)
	}

	// Check if email already exists (if provided)
	if req.Email != "" {
		existingByEmail, err := uc.userRepo.FindByEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingByEmail != nil {
			return nil, fmt.Errorf("user with email %s already exists", req.Email)
		}
	}

	// Hash password
	hashedPassword, err := uc.jwtService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &Domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *userUseCase) Login(req Domain.LoginRequest) (*Domain.LoginResponse, error) {
	// Find user by phone
	user, err := uc.userRepo.FindByPhone(req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid phone or password")
	}

	// Check user status
	if user.Status != Domain.UserStatusActive {
		return nil, fmt.Errorf("account is not active")
	}

	// Check password
	if !uc.jwtService.CheckPasswordHash(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid phone or password")
	}

	// Generate token
	token, err := uc.jwtService.GenerateToken(user.ID.Hex(), user.Phone, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &Domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (uc *userUseCase) RefreshToken(userID string) (string, error) {
	return uc.jwtService.GenerateRefreshToken(userID)
}

func (uc *userUseCase) GetUserByID(id string) (*Domain.User, error) {
	return uc.userRepo.FindByID(id)
}

func (uc *userUseCase) GetCurrentUser(id string) (*Domain.User, error) {
	return uc.userRepo.FindByID(id)
}

func (uc *userUseCase) UpdateUser(id string, req Domain.UpdateUserRequest) (*Domain.User, error) {
	user, err := uc.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if email already exists
		if req.Email != user.Email {
			existing, err := uc.userRepo.FindByEmail(req.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing email: %w", err)
			}
			if existing != nil {
				return nil, fmt.Errorf("email already in use")
			}
			user.Email = req.Email
		}
	}
	if req.Phone != "" {
		// Check if phone already exists
		if req.Phone != user.Phone {
			existing, err := uc.userRepo.FindByPhone(req.Phone)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing phone: %w", err)
			}
			if existing != nil {
				return nil, fmt.Errorf("phone already in use")
			}
			user.Phone = req.Phone
		}
	}

	if err := uc.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (uc *userUseCase) UpdateStatus(id string, status Domain.UserStatus) error {
	return uc.userRepo.UpdateStatus(id, status)
}

func (uc *userUseCase) DeleteUser(id string) error {
	return uc.userRepo.Delete(id)
}
