package Domain

import (
	"time"
)

type SyncOperation string

const (
	SyncOperationCreate SyncOperation = "create"
	SyncOperationUpdate SyncOperation = "update"
	SyncOperationDelete SyncOperation = "delete"
)

type SyncItem struct {
	ID         string        `json:"id"`
	LocalID    string        `json:"local_id"`
	Operation  SyncOperation `json:"operation"`
	EntityType string        `json:"entity_type"` // sale, expense, product
	Data       interface{}   `json:"data"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type SyncBatch struct {
	BusinessID string     `json:"business_id" validate:"required"`
	Items      []SyncItem `json:"items" validate:"required"`
	DeviceID   string     `json:"device_id" validate:"required"`
	Timestamp  time.Time  `json:"timestamp" validate:"required"`
}

type SyncResponse struct {
	Success    []SyncResult `json:"success"`
	Failed     []SyncResult `json:"failed"`
	ServerTime time.Time    `json:"server_time"`
}

type SyncResult struct {
	LocalID   string    `json:"local_id"`
	ServerID  string    `json:"server_id,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type SyncStatus struct {
	LastSync    time.Time `json:"last_sync"`
	Pending     int       `json:"pending"`
	SyncedToday int       `json:"synced_today"`
	Total       int       `json:"total"`
}

type SyncRepository interface {
	LogSync(businessID, deviceID string, items []SyncItem, result SyncResponse) error
	GetSyncStatus(businessID string) (*SyncStatus, error)
	GetLastSync(businessID, deviceID string) (*time.Time, error)
}

// Helper function to get current time
func TimeNow() time.Time {
	return time.Now()
}
