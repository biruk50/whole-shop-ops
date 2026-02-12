package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Business struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name         string             `bson:"name" json:"name" validate:"required"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	BusinessType string             `bson:"business_type" json:"business_type" validate:"required"`
	Currency     string             `bson:"currency" json:"currency" validate:"required"`
	Timezone     string             `bson:"timezone" json:"timezone"`
	Address      string             `bson:"address,omitempty" json:"address,omitempty"`
	City         string             `bson:"city,omitempty" json:"city,omitempty"`
	Country      string             `bson:"country,omitempty" json:"country,omitempty"`
	Phone        string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Email        string             `bson:"email,omitempty" json:"email,omitempty"`
	Status       BusinessStatus     `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type BusinessStatus string

const (
	BusinessStatusActive   BusinessStatus = "active"
	BusinessStatusInactive BusinessStatus = "inactive"
	BusinessStatusClosed   BusinessStatus = "closed"
)

type CreateBusinessRequest struct {
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description,omitempty"`
	BusinessType string `json:"business_type" validate:"required"`
	Currency     string `json:"currency" validate:"required"`
	Timezone     string `json:"timezone,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	Country      string `json:"country,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
}

type UpdateBusinessRequest struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	BusinessType string `json:"business_type,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Timezone     string `json:"timezone,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	Country      string `json:"country,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
}

type BusinessRepository interface {
	Create(business *Business) error
	FindByID(id string) (*Business, error)
	FindByUserID(userID string) ([]Business, error)
	Update(business *Business) error
	UpdateStatus(id string, status BusinessStatus) error
	Delete(id string) error
	FindByPhone(phone string) (*Business, error)
}
