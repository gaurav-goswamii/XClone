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

	comment.UserID = strings.TrimSpace(comment.UserID)
	if comment.UserID == "" {
		return nil, ErrUnauthorizedUser
	}

	comment.PostID = strings.TrimSpace(comment.PostID)
	if comment.PostID == "" {
		return nil, ErrPostNotFound
	}

	exists, err := s.postRepo.PostExists(comment.PostID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}

	created, err := s.commentRepo.Create(comment)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *CommentService) GetComments(
	postID string,
	limit, offset int,
) ([]*models.CommentResponse, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return nil, ErrPostNotFound
	}

	exists, err := s.postRepo.PostExists(postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}

	comments, err := s.commentRepo.ListByPost(postID, limit, offset)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *CommentService) DeleteComment(
	commentID, userID string,
) error {
	commentID = strings.TrimSpace(commentID)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrUnauthorizedUser
	}
	if commentID == "" {
		return ErrCommentNotFound
	}

	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(comment.UserID) != userID {
		return ErrNotCommentOwner
	}
	return s.commentRepo.Delete(comment.ID, comment.PostID, userID)
}
