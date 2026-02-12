package Infrastructure

import "golang.org/x/crypto/bcrypt"

type PasswordService interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type bcryptService struct{}

func NewPasswordService() PasswordService { return &bcryptService{} }

func (b *bcryptService) Hash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func (b *bcryptService) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
