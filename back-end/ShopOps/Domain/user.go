package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email"`
	Phone     string             `bson:"phone" json:"phone" validate:"required"`
	Password  string             `bson:"password" json:"-" validate:"required,min=6"`
	Role      UserRole           `bson:"role" json:"role"`
	Status    UserStatus         `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserRole string

const (
	RoleBusinessOwner UserRole = "business_owner"
	RoleStaff         UserRole = "staff"
	RoleAdmin         UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty"`
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone" validate:"omitempty"`
}

type UserRepository interface {
	Create(user *User) error
	FindByID(id string) (*User, error)
	FindByPhone(phone string) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(user *User) error
	UpdateStatus(id string, status UserStatus) error
	Delete(id string) error
}

func PrimitiveObjectIDFromHex(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}
