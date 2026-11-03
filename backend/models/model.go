package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	BirthDate    string    `json:"birth_date"`
	AvatarURL    string    `json:"avatar_url"`
	Bio          string    `json:"bio"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Likes     int       `json:"likes"`
	Retweets  int       `json:"retweets"`
	Replies   int       `json:"replies"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Chat struct {
	    ID string `json:"id"`
		User1ID string `json:"user1_id"`
		User2ID string `json:"user2_id"`
		CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID string `json:"id"`
	ChatID string `json:"chat_id"`
	SenderID string `json:"sender_id"`
	Content string `json:"content"`
	Seq int64 `json:"seq"`
	CreatedAt time.Time `json:"created_at"`
}

type PostMedia struct {
	Ordinal   int    `json:"ordinal"`
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	MediaType string `json:"media_type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

type PostResponse struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	AuthorName     string      `json:"author_name"`
	AuthorUsername string      `json:"author_username"`
	AuthorAvatar   string      `json:"author_avatar"`
	Content        string      `json:"content"`
	Media          []PostMedia `json:"media"`
	Likes          int         `json:"likes"`
	Retweets       int         `json:"retweets"`
	Replies        int         `json:"replies"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

