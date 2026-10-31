package service

import (
	"backend/models"
	"backend/utils"
	"errors"
	"strings"
	"unicode/utf8"
)

// MaxMessageRunes is the longest a message may be, counted in Unicode code
// points rather than bytes so an emoji costs one character, like MaxPostRunes
const MaxMessageRunes = 2000

var (
	ErrEmptyMessage    = errors.New("message cannot be empty")
	ErrMessageTooLong  = errors.New("message exceeds maximum length")
	ErrInvalidEncoding = errors.New("message contains invalid UTF-8")
)

// normalizeMessageContent trims a message and applies the length rule. Unlike
// a post, a message has nothing but its text, so empty is always a rejection
// here rather than something the caller decides.
func normalizeMessageContent(content string) (string, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return "", ErrEmptyMessage
	}

	// A body that is not valid UTF-8 would be stored and then handed back to
	// every reader of the chat, so it is refused at the edge of the service
	if !utf8.ValidString(content) {
		return "", ErrInvalidEncoding
	}

	if utf8.RuneCountInString(content) > MaxMessageRunes {
		return "", ErrMessageTooLong
	}

	return content, nil
}

// SendMessage appends a message to a chat the caller participates in
func (s *ChatService) SendMessage(
	userID string,
	chatID string,
	content string,
) (*models.Message, error) {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, err
	}

	if err := assertParticipant(chat, userID); err != nil {
		return nil, err
	}

	content, err = normalizeMessageContent(content)
	if err != nil {
		return nil, err
	}

	message := &models.Message{
		ID:       utils.GenerateMessageID(),
		ChatID:   chatID,
		SenderID: userID,
		Content:  content,
	}

	// Create fills in Seq and CreatedAt from the database, so the message
	// returned to the sender already carries the cursor the client polls with
	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

// ListMessages returns a chat's messages oldest-first for a participant
func (s *ChatService) ListMessages(
	userID string,
	chatID string,
) ([]models.Message, error) {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, err
	}

	if err := assertParticipant(chat, userID); err != nil {
		return nil, err
	}

	return s.messageRepo.ListByChat(chatID)
}
