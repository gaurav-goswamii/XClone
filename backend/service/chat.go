package service

import (
	"backend/models"
	"backend/repository"
	"backend/utils"
	"errors"
)

var (
	ErrUnauthorized = errors.New("unauthorized user")
	ErrInvalidUser  = errors.New("invalid user")
	ErrChatNotFound = repository.ErrChatNotFound
	ErrRecipientNotFound = errors.New("recipient not found")
	ErrCannotMessageSelf = errors.New("you cannot message yourself")
)

type ChatService struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

func NewChatService(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

func (s *ChatService) StartChatWithUsername(
	userID string,
	username string,
) (*models.Chat, error) {
	other, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}

	return s.StartChat(userID, other.ID)
}

func (s *ChatService) StartChat(
	userID string,
	otherUserID string,
) (*models.Chat, error) {
	if userID == "" || otherUserID == "" {
		return nil, ErrInvalidUser
	}
	if userID == otherUserID {
		return nil, ErrCannotMessageSelf
	}

	existingChat, err := s.chatRepo.GetByUsers(userID, otherUserID)
	if err == nil {
		return existingChat, nil
	}
	if !errors.Is(err, ErrChatNotFound) {
		return nil, err
	}

	user1ID, user2ID := canonicalPair(userID, otherUserID)

	chat := &models.Chat{
		ID:      utils.GenerateChatID(),
		User1ID: user1ID,
		User2ID: user2ID,
	}

	if err := s.chatRepo.Create(chat); err != nil {
		if existing, getErr := s.chatRepo.GetByUsers(user1ID, user2ID); getErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return chat, nil
}

func (s *ChatService) GetChat(userID, chatID string) (*models.Chat, error) {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, err
	}

	if err := assertParticipant(chat, userID); err != nil {
		return nil, err
	}

	return chat, nil
}

func assertParticipant(chat *models.Chat, userID string) error {
	if chat.User1ID != userID && chat.User2ID != userID {
		return ErrUnauthorized
	}
	return nil
}

func canonicalPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}
