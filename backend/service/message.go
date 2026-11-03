package service

import (
	"backend/models"
	"backend/utils"
	"errors"
	"strings"
	"unicode/utf8"
)

const MaxMessageRunes = 2000

var (
	ErrEmptyMessage    = errors.New("message cannot be empty")
	ErrMessageTooLong  = errors.New("message exceeds maximum length")
	ErrInvalidEncoding = errors.New("message contains invalid UTF-8")
)

func normalizeMessageContent(content string) (string, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return "", ErrEmptyMessage
	}

	if !utf8.ValidString(content) {
		return "", ErrInvalidEncoding
	}

	if utf8.RuneCountInString(content) > MaxMessageRunes {
		return "", ErrMessageTooLong
	}

	return content, nil
}

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

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

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
