package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultCost = 12
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type bcryptHasher struct {
	cost int
}

func NewHasher() Hasher {
	return &bcryptHasher{
		cost: DefaultCost,
	}
}

func NewHasherWithCost(cost int) Hasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultCost
	}
	return &bcryptHasher{
		cost: cost,
	}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *bcryptHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
