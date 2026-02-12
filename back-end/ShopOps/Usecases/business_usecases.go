package Usecases

import (
	"fmt"

	Domain "ShopOps/Domain"
)

type BusinessUseCase interface {
	CreateBusiness(userID string, req Domain.CreateBusinessRequest) (*Domain.Business, error)
	GetBusinessByID(id string) (*Domain.Business, error)
	GetUserBusinesses(userID string) ([]Domain.Business, error)
	UpdateBusiness(id, userID string, req Domain.UpdateBusinessRequest) (*Domain.Business, error)
	UpdateBusinessStatus(id, userID string, status Domain.BusinessStatus) error
	DeleteBusiness(id, userID string) error
	ValidateBusinessAccess(businessID, userID string) (*Domain.Business, error)
}

type businessUseCase struct {
	businessRepo Domain.BusinessRepository
	userRepo     Domain.UserRepository
}

func NewBusinessUseCase(businessRepo Domain.BusinessRepository, userRepo Domain.UserRepository) BusinessUseCase {
	return &businessUseCase{
		businessRepo: businessRepo,
		userRepo:     userRepo,
	}
}

func (uc *businessUseCase) CreateBusiness(userID string, req Domain.CreateBusinessRequest) (*Domain.Business, error) {
	// Validate user exists
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Set default timezone if not provided
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	// Set default currency if not provided
	if req.Currency == "" {
		req.Currency = "USD"
	}

	objUserID, err := Domain.PrimitiveObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	business := &Domain.Business{
		UserID:       objUserID,
		Name:         req.Name,
		Description:  req.Description,
		BusinessType: req.BusinessType,
		Currency:     req.Currency,
		Timezone:     req.Timezone,
		Address:      req.Address,
		City:         req.City,
		Country:      req.Country,
		Phone:        req.Phone,
		Email:        req.Email,
	}

	if err := uc.businessRepo.Create(business); err != nil {
		return nil, fmt.Errorf("failed to create business: %w", err)
	}

	return business, nil
}

func (uc *businessUseCase) GetBusinessByID(id string) (*Domain.Business, error) {
	return uc.businessRepo.FindByID(id)
}

func (uc *businessUseCase) GetUserBusinesses(userID string) ([]Domain.Business, error) {
	return uc.businessRepo.FindByUserID(userID)
}

func (uc *businessUseCase) UpdateBusiness(id, userID string, req Domain.UpdateBusinessRequest) (*Domain.Business, error) {
	// Validate business exists and user has access
	business, err := uc.ValidateBusinessAccess(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		business.Name = req.Name
	}
	if req.Description != "" {
		business.Description = req.Description
	}
	if req.BusinessType != "" {
		business.BusinessType = req.BusinessType
	}
	if req.Currency != "" {
		business.Currency = req.Currency
	}
	if req.Timezone != "" {
		business.Timezone = req.Timezone
	}
	if req.Address != "" {
		business.Address = req.Address
	}
	if req.City != "" {
		business.City = req.City
	}
	if req.Country != "" {
		business.Country = req.Country
	}
	if req.Phone != "" {
		business.Phone = req.Phone
	}
	if req.Email != "" {
		business.Email = req.Email
	}

	if err := uc.businessRepo.Update(business); err != nil {
		return nil, fmt.Errorf("failed to update business: %w", err)
	}

	return business, nil
}

func (uc *businessUseCase) UpdateBusinessStatus(id, userID string, status Domain.BusinessStatus) error {
	// Validate business exists and user has access
	_, err := uc.ValidateBusinessAccess(id, userID)
	if err != nil {
		return err
	}

	return uc.businessRepo.UpdateStatus(id, status)
}

func (uc *businessUseCase) DeleteBusiness(id, userID string) error {
	// Validate business exists and user has access
	_, err := uc.ValidateBusinessAccess(id, userID)
	if err != nil {
		return err
	}

	return uc.businessRepo.Delete(id)
}

func (uc *businessUseCase) ValidateBusinessAccess(businessID, userID string) (*Domain.Business, error) {
	business, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}
	if business == nil {
		return nil, fmt.Errorf("business not found")
	}

	// Check if user owns this business
	if business.UserID.Hex() != userID {
		return nil, fmt.Errorf("access denied: user does not own this business")
	}

	return business, nil
}
