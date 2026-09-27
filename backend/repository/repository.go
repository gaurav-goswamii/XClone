package repository

import (
	"backend/models"
)

// UserRepository defines data access methods for users
type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
}

// PostRepository defines data access methods for posts
type PostRepository interface {
	GetByID(id string) (*models.Post, error)
	ListRecent(limit int) ([]*models.Post, error)
	Create(post *models.Post) error
}
