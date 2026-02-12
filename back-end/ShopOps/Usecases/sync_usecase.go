package Usecases

import (
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
)

type SyncUseCase interface {
	ProcessBatch(batch Domain.SyncBatch) (*Domain.SyncResponse, error)
	GetSyncStatus(businessID string) (*Domain.SyncStatus, error)
	ValidateBatch(batch Domain.SyncBatch) error
	GetLastSync(businessID, deviceID string) (*time.Time, error)
}

type syncUseCase struct {
	syncService   Infrastructure.SyncService
	businessRepo  Domain.BusinessRepository
	salesRepo     Domain.SaleRepository
	expenseRepo   Domain.ExpenseRepository
	inventoryRepo Domain.ProductRepository
	syncRepo      Domain.SyncRepository
}

func NewSyncUseCase(
	syncService Infrastructure.SyncService,
	businessRepo Domain.BusinessRepository,
	salesRepo Domain.SaleRepository,
	expenseRepo Domain.ExpenseRepository,
	inventoryRepo Domain.ProductRepository,
	syncRepo Domain.SyncRepository,
) SyncUseCase {
	return &syncUseCase{
		syncService:   syncService,
		businessRepo:  businessRepo,
		salesRepo:     salesRepo,
		expenseRepo:   expenseRepo,
		inventoryRepo: inventoryRepo,
		syncRepo:      syncRepo,
	}
}

func (uc *syncUseCase) ProcessBatch(batch Domain.SyncBatch) (*Domain.SyncResponse, error) {
	// Validate batch
	if err := uc.ValidateBatch(batch); err != nil {
		return nil, fmt.Errorf("batch validation failed: %w", err)
	}

	// Process batch using sync service
	response, err := uc.syncService.ProcessBatch(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to process batch: %w", err)
	}

	return response, nil
}

func (uc *syncUseCase) GetSyncStatus(businessID string) (*Domain.SyncStatus, error) {
	// Validate business exists
	_, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}

	return uc.syncRepo.GetSyncStatus(businessID)
}

func (uc *syncUseCase) GetLastSync(businessID, deviceID string) (*time.Time, error) {
	return uc.syncRepo.GetLastSync(businessID, deviceID)
}

func (uc *syncUseCase) ValidateBatch(batch Domain.SyncBatch) error {
	// Validate business exists
	business, err := uc.businessRepo.FindByID(batch.BusinessID)
	if err != nil {
		return fmt.Errorf("invalid business ID: %w", err)
	}
	if business == nil {
		return fmt.Errorf("business not found")
	}

	// Validate device ID
	if batch.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}

	// Validate timestamp (not too far in future or past)
	now := time.Now()
	maxFuture := now.Add(24 * time.Hour)
	maxPast := now.Add(-7 * 24 * time.Hour) // Allow 7 days in past for offline data

	if batch.Timestamp.After(maxFuture) {
		return fmt.Errorf("timestamp is too far in the future")
	}

	if batch.Timestamp.Before(maxPast) {
		return fmt.Errorf("timestamp is too far in the past")
	}

	// Validate items
	if len(batch.Items) == 0 {
		return fmt.Errorf("no items to sync")
	}

	if len(batch.Items) > 1000 {
		return fmt.Errorf("batch too large. Maximum 1000 items per batch")
	}

	// Validate each item
	for _, item := range batch.Items {
		if err := uc.validateSyncItem(item); err != nil {
			return fmt.Errorf("invalid sync item: %w", err)
		}
	}

	return nil
}

func (uc *syncUseCase) validateSyncItem(item Domain.SyncItem) error {
	// Validate local ID
	if item.LocalID == "" {
		return fmt.Errorf("local ID is required")
	}

	// Validate operation
	switch item.Operation {
	case Domain.SyncOperationCreate, Domain.SyncOperationUpdate, Domain.SyncOperationDelete:
		// Valid operations
	default:
		return fmt.Errorf("invalid operation: %s", item.Operation)
	}

	// Validate entity type
	switch item.EntityType {
	case "sale", "expense", "product":
		// Valid entity types
	default:
		return fmt.Errorf("invalid entity type: %s", item.EntityType)
	}

	// Validate timestamps
	if item.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	if item.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}

	// Validate data based on operation
	if item.Operation != Domain.SyncOperationDelete && item.Data == nil {
		return fmt.Errorf("data is required for create/update operations")
	}

	return nil
}
