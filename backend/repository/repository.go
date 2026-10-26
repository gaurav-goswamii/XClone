package repository

import (
	"errors"
	"backend/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPostNotFound      = errors.New("post not found")
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
	// Create persists the post and its photos atomically; media may be empty
	Create(post *models.Post, media []models.PostMedia) error
	GetByID(id string) (*models.PostResponse, error)
	ListFeed(limit, offset int) ([]*models.PostResponse, error)
	ListByUser(username string, limit, offset int) ([]*models.PostResponse, error)
	Update(id, userID, content string, media []models.PostMedia) error
	Delete(id, userID string) error
	GetMediaByPostID(postID string) ([]models.PostMedia, error)
}

