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

// SaveVerificationCode stores an OTP code for an email with expiration
func (r *PostgresUserRepository) SaveVerificationCode(email, code string, ttl time.Duration) error {
	normalized := strings.ToLower(strings.TrimSpace(email))
	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO verification_records (email, code, expires_at, attempts)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (email)
		DO UPDATE SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at, attempts = 0
	`
	_, err := r.db.Exec(query, normalized, strings.TrimSpace(code), expiresAt)
	return err
}

// VerifyCode verifies the OTP code against database record
func (r *PostgresUserRepository) VerifyCode(email, code string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT code, expires_at, attempts
		FROM verification_records
		WHERE email = $1
	`
	var storedCode string
	var expiresAt time.Time
	var attempts int

	err := r.db.QueryRow(query, normalized).Scan(&storedCode, &expiresAt, &attempts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrVerificationNotFound
		}
		return false, err
	}

	if time.Now().After(expiresAt) {
		_, _ = r.db.Exec(`DELETE FROM verification_records WHERE email = $1`, normalized)
		return false, ErrVerificationExpired
	}

	attempts++
	if attempts > 5 {
		_, _ = r.db.Exec(`DELETE FROM verification_records WHERE email = $1`, normalized)
		return false, errors.New("maximum verification attempts exceeded, please request a new code")
	}

	if storedCode != strings.TrimSpace(code) {
		_, _ = r.db.Exec(`UPDATE verification_records SET attempts = $1 WHERE email = $2`, attempts, normalized)
		return false, ErrInvalidVerification
	}

	// Code successfully verified: remove record
	_, _ = r.db.Exec(`DELETE FROM verification_records WHERE email = $1`, normalized)
	return true, nil
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

// NewPostgresPostRepository creates a new PostgresPostRepository
func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// GetByID retrieves a post by ID
func (r *PostgresPostRepository) GetByID(id string) (*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_url, likes, retweets, replies, created_at, updated_at
		FROM posts
		WHERE id = $1
	`
	var post models.Post
	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.UserID,
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

// ListRecent retrieves the most recent posts
func (r *PostgresPostRepository) ListRecent(limit int) ([]*models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, user_id, content, media_url, likes, retweets, replies, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
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
