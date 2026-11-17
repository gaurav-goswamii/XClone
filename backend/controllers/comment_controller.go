package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/models"
	"backend/service"
	"backend/utils"
)

const maxCommentBodyBytes = 16 << 10

type CommentController struct {
	commentService *service.CommentService
	jwtSecret      string
}

func NewCommentController(commentService *service.CommentService, jwtSecret string) *CommentController {
	return &CommentController{
		commentService: commentService,
		jwtSecret:      jwtSecret,
	}
}

func (c *CommentController) requireAuth(w http.ResponseWriter, r *http.Request) (*utils.JWTClaims, bool) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return nil, false
	}

	claims, err := utils.ValidateJWT(cookie.Value, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired auth token")
		return nil, false
	}

	return claims, true
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func writeCommentError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrCommentNotFound):
		writeError(w, http.StatusNotFound, "comment not found")
	case errors.Is(err, service.ErrPostNotFound):
		writeError(w, http.StatusNotFound, "post not found")
	case errors.Is(err, service.ErrNotCommentOwner):
		writeError(w, http.StatusForbidden, "you can only delete your own comments")
	case errors.Is(err, service.ErrUnauthorizedUser):
		writeError(w, http.StatusUnauthorized, "unauthorized user")
	case errors.Is(err, service.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, service.ErrEmptyComment):
		writeError(w, http.StatusBadRequest, "comment cannot be empty")
	case errors.Is(err, service.ErrCommentTooLong):
		writeError(w, http.StatusBadRequest, "comment exceeds 280 characters")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

func (c *CommentController) PostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.ListCommentsHandler(w, r)
	case http.MethodPost:
		c.CreateCommentHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (c *CommentController) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	postID := strings.TrimSpace(r.PathValue("id"))
	if postID == "" {
		writeError(w, http.StatusBadRequest, "post id is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCommentBodyBytes)

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment := &models.Comment{
		ID:      utils.GenerateCommentID(),
		PostID:  postID,
		UserID:  claims.UserID,
		Content: req.Content,
	}

	created, err := c.commentService.CreateComment(comment)
	if err != nil {
		writeCommentError(w, err, "failed to create comment")
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "comment created successfully",
		Data: map[string]interface{}{
			"comment": created,
		},
	})
}

func (c *CommentController) ListCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	postID := strings.TrimSpace(r.PathValue("id"))
	if postID == "" {
		writeError(w, http.StatusBadRequest, "post id is required")
		return
	}

	limit := 20
	offset := 0

	query := r.URL.Query()
	if rawLimit := query.Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if rawOffset := query.Get("offset"); rawOffset != "" {
		if parsed, err := strconv.Atoi(rawOffset); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	comments, err := c.commentService.GetComments(postID, limit, offset)
	if err != nil {
		writeCommentError(w, err, "failed to fetch comments")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"comments": comments,
		},
	})
}

func (c *CommentController) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	commentID := strings.TrimSpace(r.PathValue("id"))
	if commentID == "" {
		writeError(w, http.StatusBadRequest, "comment id is required")
		return
	}

	if err := c.commentService.DeleteComment(commentID, claims.UserID); err != nil {
		writeCommentError(w, err, "failed to delete comment")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "comment deleted successfully",
	})
}
