package controllers

import (
	"encoding/json"
	"errors"
	"io" "net/http" "strconv" "strings" "backend/models" "backend/service" "backend/utils"
)

const maxCommentBodyBytes = 16 << 10

type CommentController struct {
	commentService *service.CommentService
	jwtSecret string
}

func NewCommentController(commentService *service.CommentService, jwtSecret string) *CommentController {
	return &CommentController{
		commentService: commentService,
		jwtSecret: jwtSecret,
	}
}

func(c *CommentController) requireAuth(w http.ResponseWriter, r*http.Request) (*utils.JWTClaims, bool){
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

type DeleteCommentRequest struct {
	PostID string
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