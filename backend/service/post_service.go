package service

import (
	"errors"
	"strings"
	"time"
	"backend/models"
	"backend/repository"
	"backend/utils"
)

var (
	ErrEmptyContent     = errors.New("post content cannot be empty")
	ErrContentTooLong   = errors.New("post content exceeds 280 characters")
	ErrUnauthorizedPost = errors.New("unauthorized user")
	ErrNotPostOwner     = errors.New("you can only modify your own posts")
	// ErrPostNotFound re-exports the repository sentinel so controllers depend on the service alone
	ErrPostNotFound = repository.ErrPostNotFound
)

// PostService provides business operations for posts
type PostService struct {
	postRepo repository.PostRepository
}

// NewPostService creates a new instance of PostService
func NewPostService(postRepo repository.PostRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
	}
}

// validateContent trims post content and applies the shared length rules
func validateContent(content string) (string, error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return "", ErrEmptyContent
	}
	//counting characters(for emojis got to use this setup)
	if len([]rune(trimmedContent)) > 280 {
		return "", ErrContentTooLong
	}
	return trimmedContent, nil
}

// CreatePost validates and persists a new text-only post
func (s *PostService) CreatePost(userID, content string) (*models.PostResponse, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil, ErrUnauthorizedPost
	}

	trimmedContent, err := validateContent(content)
	if err != nil {
		return nil, err
	}

	postID := utils.GeneratePostID()
	now := time.Now()

	post := &models.Post{
		ID:        postID,
		UserID:    trimmedUserID,
		Content:   trimmedContent,
		MediaURL:  "",
		Likes:     0,
		Retweets:  0,
		Replies:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	return s.postRepo.GetByID(postID)
}

// GetFeed retrieves timeline posts with pagination
func (s *PostService) GetFeed(limit, offset int) ([]*models.PostResponse, error) {
	return s.postRepo.ListFeed(limit, offset)
}

// GetUserPosts retrieves posts for a specific user
func (s *PostService) GetUserPosts(username string, limit, offset int) ([]*models.PostResponse, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return []*models.PostResponse{}, nil
	}
	return s.postRepo.ListByUser(cleanUsername, limit, offset)
}

// GetPost retrieves a single post by its ID
func (s *PostService) GetPost(postID string) (*models.PostResponse, error) {
	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return nil, ErrPostNotFound
	}
	return s.postRepo.GetByID(trimmedPostID)
}

// UpdatePost rewrites the content of a post the user owns
func (s *PostService) UpdatePost(userID, postID, content string) (*models.PostResponse, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil, ErrUnauthorizedPost
	}

	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return nil, ErrPostNotFound
	}

	trimmedContent, err := validateContent(content)
	if err != nil {
		return nil, err
	}

	existing, err := s.postRepo.GetByID(trimmedPostID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != trimmedUserID {
		return nil, ErrNotPostOwner
	}

	if err := s.postRepo.Update(trimmedPostID, trimmedUserID, trimmedContent); err != nil {
		return nil, err
	}

	return s.postRepo.GetByID(trimmedPostID)
}

// DeletePost permanently removes a post the user owns
func (s *PostService) DeletePost(userID, postID string) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return ErrUnauthorizedPost
	}

	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return ErrPostNotFound
	}

	existing, err := s.postRepo.GetByID(trimmedPostID)
	if err != nil {
		return err
	}
	if existing.UserID != trimmedUserID {
		return ErrNotPostOwner
	}

	return s.postRepo.Delete(trimmedPostID, trimmedUserID)
}
