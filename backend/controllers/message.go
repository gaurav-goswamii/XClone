package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Content string `json:"content"`
}

// MessagesHandler handles GET and POST on /api/chats/{id}/messages
func (c *ChatController) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	chatID := strings.TrimSpace(r.PathValue("id"))
	if chatID == "" {
		writeError(w, http.StatusBadRequest, "chat id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		c.listMessages(w, r, chatID)
	case http.MethodPost:
		c.sendMessage(w, r, chatID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listMessages serves a chat's history to one of its two participants
func (c *ChatController) listMessages(w http.ResponseWriter, r *http.Request, chatID string) {
	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	messages, err := c.chatService.ListMessages(claims.UserID, chatID)
	if err != nil {
		writeChatError(w, err, "failed to fetch messages")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"messages": messages,
		},
	})
}

// sendMessage appends one message to a chat the caller participates in
func (c *ChatController) sendMessage(w http.ResponseWriter, r *http.Request, chatID string) {
	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxChatBodyBytes)

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	message, err := c.chatService.SendMessage(claims.UserID, chatID, req.Content)
	if err != nil {
		writeChatError(w, err, "failed to send message")
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "message sent successfully",
		Data: map[string]interface{}{
			"message": message,
		},
	})
}
