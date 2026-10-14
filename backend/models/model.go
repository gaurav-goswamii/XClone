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

// PostResponse represents a post returned with author profile details
type PostResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AuthorName     string    `json:"author_name"`
	AuthorUsername string    `json:"author_username"`
	AuthorAvatar   string    `json:"author_avatar"`
	Content        string    `json:"content"`
	MediaURL       string    `json:"media_url,omitempty"`
	Likes          int       `json:"likes"`
	Retweets       int       `json:"retweets"`
	Replies        int       `json:"replies"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

