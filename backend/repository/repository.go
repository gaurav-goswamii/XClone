package repository

import (
	"errors"
	"backend/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPostNotFound      = errors.New("post not found")
	ErrChatNotFound      = errors.New("chat not found")
)

type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	UpdateProfile(id string, name string, avatarURL string, bio string) error
}

type PostRepository interface {
	Create(post *models.Post, media []models.PostMedia) error
	GetByID(id string) (*models.PostResponse, error)
	ListFeed(limit, offset int) ([]*models.PostResponse, error)
	ListByUser(username string, limit, offset int) ([]*models.PostResponse, error)
	Update(id, userID, content string, media []models.PostMedia) error
	Delete(id, userID string) error
	GetMediaByPostID(postID string) ([]models.PostMedia, error)
}

type ChatRepository interface {
	Create(chat *models.Chat) error
	GetByID(chatID string) (*models.Chat, error)
	GetByUsers(user1ID, user2ID string) (*models.Chat, error)
}

type MessageRepository interface {
	Create(message *models.Message) error
	ListByChat(chatID string) ([]models.Message, error)
}