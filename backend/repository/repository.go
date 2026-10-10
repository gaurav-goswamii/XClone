package repository

import (
	"time"

	"backend/models"
)

// UserRepository defines data access methods for users
type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	SaveVerificationCode(email, code string, ttl time.Duration) error
	VerifyCode(email, code string) (bool, error)
	Create(user *models.User) error
	UpdateProfile(id string, name string, avatarURL string, bio string) error
}

// PostRepository defines data access methods for posts
type PostRepository interface {
	GetByID(id string) (*models.Post, error)
	ListRecent(limit int) ([]*models.Post, error)
	Create(post *models.Post) error
}
