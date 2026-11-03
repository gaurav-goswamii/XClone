package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"backend/service"
	"backend/utils"
)

const maxChatBodyBytes = 16 << 10

type ChatController struct {
	chatService *service.ChatService
	jwtSecret   string
}

func NewChatController(chatService *service.ChatService, jwtSecret string) *ChatController {
	return &ChatController{
		chatService: chatService,
		jwtSecret:   jwtSecret,
	}
}

func (c *ChatController) requireAuth(w http.ResponseWriter, r *http.Request) (*utils.JWTClaims, bool) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return nil, false
	}

	claims, err := utils.ValidateJWT(cookie.Value, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return nil, false
	}

	return claims, true
}

type StartChatRequest struct {
	Username string `json:"username"`
}

func (c *ChatController) ChatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxChatBodyBytes)

	var req StartChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(req.Username, "@")))
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	chat, err := c.chatService.StartChatWithUsername(claims.UserID, username)
	if err != nil {
		writeChatError(w, err, "failed to open chat")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"chat": chat,
		},
	})
}

func writeChatError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrChatNotFound):
		writeError(w, http.StatusNotFound, "chat not found")
	case errors.Is(err, service.ErrRecipientNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, service.ErrCannotMessageSelf):
		writeError(w, http.StatusBadRequest, "you cannot message yourself")
	case errors.Is(err, service.ErrInvalidUser):
		writeError(w, http.StatusBadRequest, "invalid user")
	case errors.Is(err, service.ErrEmptyMessage):
		writeError(w, http.StatusBadRequest, "message cannot be empty")
	case errors.Is(err, service.ErrMessageTooLong):
		writeError(w, http.StatusBadRequest, "message exceeds 2000 characters")
	case errors.Is(err, service.ErrInvalidEncoding):
		writeError(w, http.StatusBadRequest, "message contains invalid characters")
	case errors.Is(err, service.ErrUnauthorized):
		writeError(w, http.StatusForbidden, "you are not a participant in this chat")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}
