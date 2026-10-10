package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"backend/models"
	"backend/repository"
	"backend/utils"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,15}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// SignupRequest represents user registration inputs for the single signup endpoint
type SignupRequest struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	BirthDate string `json:"birth_date"`
}

// LoginRequest represents inputs for email-based login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthService handles authentication and account onboarding logic
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// ValidateAge checks if the birthdate satisfies the minimum age of 13 years
func ValidateAge(birthDateStr string) error {
	t, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		return errors.New("invalid date of birth format, expected YYYY-MM-DD")
	}

	now := time.Now()
	// Calculate age
	years := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		years--
	}

	if years < 13 {
		return errors.New("you must be at least 13 years old to create an account")
	}

	if years > 120 {
		return errors.New("please provide a valid year of birth")
	}

	return nil
}

// ValidatePassword ensures password is at least 8 characters with letters and numbers
func ValidatePassword(pass string) error {
	if len(pass) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasLetter, hasNumber bool
	for _, c := range pass {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		} else if c >= '0' && c <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		return errors.New("password must contain both letters and numbers")
	}

	return nil
}

// ValidateUsername checks X handle guidelines (3-15 chars, alphanumeric or underscores)
func ValidateUsername(username string) error {
	clean := strings.TrimPrefix(strings.TrimSpace(username), "@")
	if !usernameRegex.MatchString(clean) {
		return errors.New("username must be 3-15 characters and contain only letters, numbers, and underscores")
	}
	return nil
}

// CheckUsernameAvailability checks if the desired username is valid and available
func (s *AuthService) CheckUsernameAvailability(username string) (bool, string, error) {
	clean := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	if err := ValidateUsername(clean); err != nil {
		return false, err.Error(), nil
	}

	existing, _ := s.userRepo.GetByUsername(clean)
	if existing != nil {
		return false, "That username has been taken. Please choose another.", nil
	}

	return true, "Username is available!", nil
}

// Signup performs validation, checks uniqueness, hashes password, and saves the user
func (s *AuthService) Signup(req SignupRequest) (*models.User, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) == 0 || len(name) > 50 {
		return nil, errors.New("name must be between 1 and 50 characters")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !emailRegex.MatchString(email) {
		return nil, errors.New("please provide a valid email address")
	}

	if err := ValidateAge(req.BirthDate); err != nil {
		return nil, err
	}

	username := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(req.Username), "@"))
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if email already registered
	existingEmail, _ := s.userRepo.GetByEmail(email)
	if existingEmail != nil {
		return nil, errors.New("an account with this email already exists")
	}

	// Check if username already taken
	existingUsername, _ := s.userRepo.GetByUsername(username)
	if existingUsername != nil {
		return nil, errors.New("username is already taken")
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := utils.GenerateUserID()
	user := &models.User{
		ID:           userID,
		Username:     username,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		BirthDate:    req.BirthDate,
		AvatarURL:    "",
		Bio:          "",
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	return user, nil
}

// GetUserByID fetches user details by ID
func (s *AuthService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// Login authenticates a user using their email and password
func (s *AuthService) Login(req LoginRequest) (*models.User, string, error) {
	//lower and clean the input email, checks if empty or invalid format
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !emailRegex.MatchString(email) {
		return nil, "", errors.New("please provide a valid email address")
	}

	if strings.TrimSpace(req.Password) == "" {
		return nil, "", errors.New("password is required")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, "", errors.New("invalid email or password")
	}

	// Issue JWT token (valid for 30 days)
	token, err := utils.GenerateJWT(user.ID, user.Username, s.jwtSecret, 30*24*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return user, token, nil
}

// GetUserByUsername fetches user details by their handle/username
func (s *AuthService) GetUserByUsername(username string) (*models.User, error) {
	cleanHandle := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	if cleanHandle == "" {
		return nil, errors.New("username cannot be empty")
	}
	return s.userRepo.GetByUsername(cleanHandle)
}

// UpdateUserProfile updates user's display name, avatar URL, and bio
func (s *AuthService) UpdateUserProfile(userID, name, avatarURL, bio string) (*models.User, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, errors.New("name cannot be empty")
	}
	if len(trimmedName) > 50 {
		return nil, errors.New("name cannot exceed 50 characters")
	}

	trimmedBio := strings.TrimSpace(bio)
	if len(trimmedBio) > 160 {
		return nil, errors.New("bio cannot exceed 160 characters")
	}

	trimmedAvatar := strings.TrimSpace(avatarURL)

	if err := s.userRepo.UpdateProfile(userID, trimmedName, trimmedAvatar, trimmedBio); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return s.userRepo.GetByID(userID)
}
