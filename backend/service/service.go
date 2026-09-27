package service

import (
	"backend/models"
	"backend/repository"
)

// UserService defines business logic operations for users
type UserService interface {
	GetUserProfile(id string) (*models.User, error)
	RegisterUser(user *models.User) error
}

// PostService defines business logic operations for posts
type PostService interface {
	GetPost(id string) (*models.Post, error)
	GetFeed(limit int) ([]*models.Post, error)
	CreatePost(post *models.Post) error
}

// ServiceContainer aggregates all application services
type ServiceContainer struct {
	UserRepo repository.UserRepository
	PostRepo repository.PostRepository
}
