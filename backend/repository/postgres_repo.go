package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"backend/models"
	"github.com/lib/pq"
)

// PostgresUserRepository implements UserRepository using a PostgreSQL database
type PostgresUserRepository struct {
	db *sql.DB
}	

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// GetByID finds a user by their unique ID
func (r *PostgresUserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, strings.TrimSpace(id)).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail finds a user by their email (case-insensitive)
func (r *PostgresUserRepository) GetByEmail(email string) (*models.User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE LOWER(email) = $1
	`
	var user models.User
	err := r.db.QueryRow(query, normalized).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername finds a user by their username/handle (case-insensitive)
func (r *PostgresUserRepository) GetByUsername(username string) (*models.User, error) {
	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE LOWER(username) = $1
	`
	var user models.User
	err := r.db.QueryRow(query, normalized).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user record into the PostgreSQL users table
func (r *PostgresUserRepository) Create(user *models.User) error {
	emailNorm := strings.ToLower(strings.TrimSpace(user.Email))
	handleNorm := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(user.Username), "@"))

	now := time.Now()
	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := now

	query := `
		INSERT INTO users (id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(query,
		user.ID,
		handleNorm,
		strings.TrimSpace(user.Name),
		emailNorm,
		user.PasswordHash,
		user.BirthDate,
		user.AvatarURL,
		user.Bio,
		user.IsVerified,
		createdAt,
		updatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// 23505 is unique_violation
			if pqErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
		}
		return err
	}

	user.Username = handleNorm
	user.Email = emailNorm
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return nil
}

// UpdateProfile updates the name, avatar_url, and bio of a user
func (r *PostgresUserRepository) UpdateProfile(id string, name string, avatarURL string, bio string) error {
	query := `
		UPDATE users
		SET name = $1, avatar_url = $2, bio = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.Exec(query, strings.TrimSpace(name), strings.TrimSpace(avatarURL), strings.TrimSpace(bio), time.Now(), strings.TrimSpace(id))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// PostgresPostRepository implements PostRepository using PostgreSQL
type PostgresPostRepository struct {
	db *sql.DB
}

// NewPostgresPostRepository creates a new PostgresPostRepository/constructor
func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// Create inserts a new post record into the database
func (r *PostgresPostRepository) Create(post *models.Post) error {
	now := time.Now()
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	post.UpdatedAt = now

	query := `
		INSERT INTO posts (id, user_id, content, media_url, likes, retweets, replies, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(query,
		post.ID,
		post.UserID,
		post.Content,
		post.MediaURL,
		post.Likes,
		post.Retweets,
		post.Replies,
		post.CreatedAt,
		post.UpdatedAt,
	)
	return err
}

// GetByID retrieves a post by ID including author profile details
func (r *PostgresPostRepository) GetByID(id string) (*models.PostResponse, error) {
	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, COALESCE(p.media_url, ''), p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`
	var post models.PostResponse
	err := r.db.QueryRow(query, strings.TrimSpace(id)).Scan(
		&post.ID,
		&post.UserID,
		&post.AuthorName,
		&post.AuthorUsername,
		&post.AuthorAvatar,
		&post.Content,
		&post.MediaURL,
		&post.Likes,
		&post.Retweets,
		&post.Replies,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// ListFeed retrieves the most recent posts joined with author user data
func (r *PostgresPostRepository) ListFeed(limit, offset int) ([]*models.PostResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, COALESCE(p.media_url, ''), p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.PostResponse, 0)
	for rows.Next() {
		var post models.PostResponse
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorUsername,
			&post.AuthorAvatar,
			&post.Content,
			&post.MediaURL,
			&post.Likes,
			&post.Retweets,
			&post.Replies,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, rows.Err()
}

// ListByUser retrieves posts created by a specific user (identified by username)
func (r *PostgresPostRepository) ListByUser(username string, limit, offset int) ([]*models.PostResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, COALESCE(p.media_url, ''), p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE LOWER(u.username) = LOWER($1)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, strings.TrimSpace(username), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.PostResponse, 0)
	for rows.Next() {
		var post models.PostResponse
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorUsername,
			&post.AuthorAvatar,
			&post.Content,
			&post.MediaURL,
			&post.Likes,
			&post.Retweets,
			&post.Replies,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, rows.Err()
}

