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

// ChatRepository defines data access for 1:1 chats. A chat row is keyed by the
// canonically ordered user pair, so Alice-opens-Bob and Bob-opens-Alice both
// land on the same row instead of creating two conversations.
type ChatRepository interface {
	Create(chat *models.Chat) error
	GetByID(chatID string) (*models.Chat, error)
	// GetByUsers returns ErrChatNotFound when this pair has never talked. The
	// pair is ordered inside the implementation, so either argument order works.
	GetByUsers(user1ID, user2ID string) (*models.Chat, error)
}

// MessageRepository defines data access for the messages inside one chat
type MessageRepository interface {
	// Create reads seq and created_at back onto the message: the database
	// assigns both, so the caller must not guess them
	Create(message *models.Message) error
	ListByChat(chatID string) ([]models.Message, error)
}