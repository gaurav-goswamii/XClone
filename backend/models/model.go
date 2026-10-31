package models

import "time"

// User represents a user in the XClone platform
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

// Post represents a post/tweet in the XClone platform
type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url,omitempty"`
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

// PostMedia represents a single photo attached to a post, in display order
type PostMedia struct {
	// Ordinal has no omitempty on purpose: 0 is the first photo, not a missing value
	Ordinal   int    `json:"ordinal"`
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	MediaType string `json:"media_type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// PostResponse represents a post returned with author profile details
type PostResponse struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	AuthorName     string      `json:"author_name"`
	AuthorUsername string      `json:"author_username"`
	AuthorAvatar   string      `json:"author_avatar"`
	Content        string      `json:"content"`
	MediaURL       string      `json:"media_url,omitempty"`
	Media          []PostMedia `json:"media"`
	Likes          int         `json:"likes"`
	Retweets       int         `json:"retweets"`
	Replies        int         `json:"replies"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

