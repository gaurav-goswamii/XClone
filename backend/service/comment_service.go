package service

import (
	"backend/models"
	"backend/repository"
	"errors"
	"strings"
	"unicode/utf8"
)

const MaxCommentRunes = 280

var (
	ErrEmptyComment     = errors.New("comment cannot be empty")
	ErrCommentTooLong   = errors.New("comment exceeds 280 characters")
	ErrNotCommentOwner  = errors.New("you can only delete your own comments")
	ErrUnauthorizedUser = errors.New("unauthorized user")

	ErrCommentNotFound = repository.ErrCommentNotFound
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (s *CommentService) CreateComment(
	comment *models.Comment,
) (*models.CommentResponse, error) {
	if comment == nil {
		return nil, ErrEmptyComment
	}
	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Content == "" {
		return nil, ErrEmptyComment
	}
	if utf8.RuneCountInString(comment.Content) > MaxCommentRunes {
		return nil, ErrCommentTooLong
	}
	if strings.TrimSpace(comment.UserID) == "" {
		return nil, ErrUnauthorized
	}
	if _, err := s.postRepo.GetByID(comment.PostID); err != nil {
		return nil, err
	}
	return s.commentRepo.Create(comment)
}

func (s *CommentService) GetComments(
	postID string,
	limit, offset int,
) ([]*models.CommentResponse, error) {
	if strings.TrimSpace(postID) == "" {
		return nil, ErrPostNotFound
	}
	if _, err := s.postRepo.GetByID(postID); err != nil {
		return nil, err
	}
	return s.commentRepo.ListByPost(postID, limit, offset)
}

func (s *CommentService) DeleteComment(
	commentID, userID string,
) error {
	if strings.TrimSpace(userID) == "" {
		return ErrUnauthorizedUser
	}
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	if comment.UserID != userID {
		return ErrNotCommentOwner
	}
	return s.commentRepo.Delete(comment.ID, comment.PostID, userID)
}
