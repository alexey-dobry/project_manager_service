package hasher

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/student-pm/auth-service/internal/domain"
)

// BcryptHasher реализует usecase.PasswordHasher.
type BcryptHasher struct {
	cost int
}

// New — cost ∈ [4..31], рекомендуется 10–12. Меньше — для тестов.
func New(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return domain.ErrInvalidCredentials
	}
	return nil
}
