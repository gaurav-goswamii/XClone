package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"backend/models"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrVerificationNotFound  = errors.New("verification record not found")
	ErrVerificationExpired   = errors.New("verification code expired")
	ErrInvalidVerification   = errors.New("invalid verification code")
)

// MemoryUserRepository implements UserRepository using an in-memory store
type MemoryUserRepository struct {
	mu            sync.RWMutex
	usersByID     map[string]*models.User
	usersByEmail  map[string]*models.User
	usersByHandle map[string]*models.User
	verifications map[string]*models.VerificationRecord
}

// NewMemoryUserRepository creates a new thread-safe in-memory user repository
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		usersByID:     make(map[string]*models.User),
		usersByEmail:  make(map[string]*models.User),
		usersByHandle: make(map[string]*models.User),
		verifications: make(map[string]*models.VerificationRecord),
	}
}

// GetByID finds a user by their unique ID
func (r *MemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByEmail finds a user by their lowercase email address
func (r *MemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	user, ok := r.usersByEmail[normalized]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByUsername finds a user by their lowercase username
func (r *MemoryUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	user, ok := r.usersByHandle[normalized]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// SaveVerificationCode stores a 6-digit OTP code for the given email with TTL
func (r *MemoryUserRepository) SaveVerificationCode(email, code string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	r.verifications[normalized] = &models.VerificationRecord{
		Email:     normalized,
		Code:      code,
		ExpiresAt: time.Now().Add(ttl),
		Attempts:  0,
	}
	return nil
}

// VerifyCode checks the OTP code for an email; returns true if valid
func (r *MemoryUserRepository) VerifyCode(email, code string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	record, ok := r.verifications[normalized]
	if !ok {
		return false, ErrVerificationNotFound
	}

	if time.Now().After(record.ExpiresAt) {
		delete(r.verifications, normalized)
		return false, ErrVerificationExpired
	}

	record.Attempts++
	if record.Attempts > 5 {
		delete(r.verifications, normalized)
		return false, errors.New("maximum verification attempts exceeded, please request a new code")
	}

	if record.Code != strings.TrimSpace(code) {
		return false, ErrInvalidVerification
	}

	return true, nil
}

// Create inserts a new user record
func (r *MemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailNorm := strings.ToLower(strings.TrimSpace(user.Email))
	handleNorm := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(user.Username), "@"))

	if _, exists := r.usersByEmail[emailNorm]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.usersByHandle[handleNorm]; exists {
		return ErrUserAlreadyExists
	}

	userCopy := *user
	userCopy.Email = emailNorm
	userCopy.Username = handleNorm
	if userCopy.CreatedAt.IsZero() {
		userCopy.CreatedAt = time.Now()
	}
	userCopy.UpdatedAt = time.Now()

	r.usersByID[userCopy.ID] = &userCopy
	r.usersByEmail[emailNorm] = &userCopy
	r.usersByHandle[handleNorm] = &userCopy

	return nil
}

// UpdateProfile updates the name, avatar_url, and bio for a user
func (r *MemoryUserRepository) UpdateProfile(id string, name string, avatarURL string, bio string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.usersByID[id]
	if !exists {
		return ErrUserNotFound
	}

	user.Name = strings.TrimSpace(name)
	user.AvatarURL = strings.TrimSpace(avatarURL)
	user.Bio = strings.TrimSpace(bio)
	user.UpdatedAt = time.Now()

	return nil
}
