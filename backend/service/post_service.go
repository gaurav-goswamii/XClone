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

// CreatePost validates and persists a new text-only post
func (s *PostService) CreatePost(userID, content string) (*models.PostResponse, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil, ErrUnauthorizedPost
	}

	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, ErrEmptyContent
	}
     //counting characters(for emojis got to use this setup)
	if len([]rune(trimmedContent)) > 280 {
		return nil, ErrContentTooLong
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
