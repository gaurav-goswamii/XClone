package repository

import (
	"errors"
	"backend/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository defines data access methods for users
type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	UpdateProfile(id string, name string, avatarURL string, bio string) error
}

// PostRepository defines data access methods for posts
type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id string) (*models.PostResponse, error)
	ListFeed(limit, offset int) ([]*models.PostResponse, error)
	ListByUser(username string, limit, offset int) ([]*models.PostResponse, error)
}

