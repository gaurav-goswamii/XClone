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

// ChatService provides business operations for 1:1 chats and their messages.
// userRepo is here because opening a chat starts from a username the person
// typed, and "does that person exist" is a business rule, not transport.
type ChatService struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

// NewChatService creates a new instance of ChatService
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

// StartChatWithUsername resolves a username to its user and opens the chat.
// The UI only ever knows the handle, so this is the entry point the controller
// calls; StartChat stays usable on its own for an ID we already hold.
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

// StartChat returns the chat between these two users, creating it if this is
// the first time they talk. It is idempotent on purpose: opening a chat twice
// must not leave two rows behind.
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

	// The pair is ordered before the insert because the table's
	// CHECK (user1_id < user2_id) is what guarantees one row per pair: without
	// ordering here, Alice→Bob and Bob→Alice would be two different rows.
	user1ID, user2ID := canonicalPair(userID, otherUserID)

	chat := &models.Chat{
		ID:      utils.GenerateChatID(),
		User1ID: user1ID,
		User2ID: user2ID,
	}

	if err := s.chatRepo.Create(chat); err != nil {
		// Two simultaneous "open chat" clicks race on the UNIQUE pair. The
		// loser re-reads instead of failing, so both callers get the same chat.
		if existing, getErr := s.chatRepo.GetByUsers(user1ID, user2ID); getErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return chat, nil
}

// GetChat loads a chat the caller participates in
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

// assertParticipant is the security boundary of this feature: every message
// read and every message write goes through it, so a chat ID guessed by a
// stranger is a 403 and never someone else's conversation.
func assertParticipant(chat *models.Chat, userID string) error {
	if chat.User1ID != userID && chat.User2ID != userID {
		return ErrUnauthorized
	}
	return nil
}

// canonicalPair orders two user IDs so that a pair has one representation
func canonicalPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}
