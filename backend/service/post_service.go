package service

import (
	"backend/models"
	"backend/repository"
	"backend/utils"
	"context"
	"errors"
	"log"
	"strings"
	"time"
)

// MaxPostRunes is the longest a post's text may be, counted in Unicode code
// points rather than bytes so that emoji cost one character each
const MaxPostRunes = 280

var (
	ErrEmptyContent     = errors.New("post content cannot be empty")
	ErrEmptyPost        = errors.New("post must have text or at least one photo")
	ErrContentTooLong   = errors.New("post content exceeds 280 characters")
	ErrUnauthorizedPost = errors.New("unauthorized user")
	ErrNotPostOwner     = errors.New("you can only modify your own posts")
	// ErrPostNotFound re-exports the repository sentinel so controllers depend on the service alone
	ErrPostNotFound = repository.ErrPostNotFound
)

// PostService provides business operations for posts
type PostService struct {
	postRepo repository.PostRepository
	// media validates client-supplied photo references. It lives here because
	// "where our media is allowed to come from" is a business rule: putting it
	// in the controller would make the transport layer the security boundary
	media mediaResolver
	photoDeleter PhotoDeleter
}

type PhotoDeleter interface {
	DeletePhoto(ctx context.Context, publicID string) error
}



// NewPostService creates a new instance of PostService
func NewPostService(postRepo repository.PostRepository, cloudName, uploadFolder string, photoDeleter PhotoDeleter,) *PostService {
	return &PostService{
		postRepo: postRepo,
		media:    mediaResolver{cloudName: cloudName, uploadFolder: uploadFolder},
		photoDeleter: photoDeleter,
	}
}

// normalizeContent trims post content and applies the length rule. It permits
// empty text: a post may now consist of photos alone, so whether "empty" is
// acceptable depends on the photos and is decided by the caller.
func normalizeContent(content string) (string, error) {
	trimmedContent := strings.TrimSpace(content)
	//counting characters(for emojis got to use this setup)
	if len([]rune(trimmedContent)) > MaxPostRunes {
		return "", ErrContentTooLong
	}
	return trimmedContent, nil
}

// CreatePost validates and persists a new post with up to MaxPostPhotos photos
func (s *PostService) CreatePost(userID, content string, refs []MediaRef) (*models.PostResponse, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil, ErrUnauthorizedPost
	}

	// Photos are resolved first so that "five photos and no text" reports the
	// specific photo-count problem rather than the generic empty-post one
	media, err := s.media.resolve(trimmedUserID, refs)
	if err != nil {
		return nil, err
	}

	trimmedContent, err := normalizeContent(content)
	if err != nil {
		return nil, err
	}

	// Text is no longer the only way to have a post, but a post still has to
	// carry something
	if trimmedContent == "" && len(media) == 0 {
		return nil, ErrEmptyPost
	}

	postID := utils.GeneratePostID()
	now := time.Now()

	post := &models.Post{
		ID:        postID,
		UserID:    trimmedUserID,
		Content:   trimmedContent,
		Likes:     0,
		Retweets:  0,
		Replies:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	// media_url is the legacy single-image column. Mirroring the first photo
	// into it keeps anything that still reads that field working.
	if len(media) > 0 {
		post.MediaURL = media[0].URL
	}

	if err := s.postRepo.Create(post, media); err != nil {
		return nil, err
	}

	// The re-read is what fills in the author details and the Media slice
	return s.postRepo.GetByID(postID)
}

// GetFeed retrieves timeline posts with pagination
func (s *PostService) GetFeed(limit, offset int) ([]*models.PostResponse, error) {
	return s.postRepo.ListFeed(limit, offset)
}

// GetUserPosts retrieves posts for a specific user
func (s *PostService) GetUserPosts(username string, limit, offset int) ([]*models.PostResponse, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return []*models.PostResponse{}, nil
	}
	return s.postRepo.ListByUser(cleanUsername, limit, offset)
}

// GetPost retrieves a single post by its ID
func (s *PostService) GetPost(postID string) (*models.PostResponse, error) {
	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return nil, ErrPostNotFound
	}
	return s.postRepo.GetByID(trimmedPostID)
}

// UpdatePost rewrites a post the user owns: its caption, its photos, or both.
// refs carries the photo list the caller wants the post to end up with. A nil
// slice means the caller did not mention photos at all and the existing ones
// are kept; an empty, non-nil slice is an explicit "remove them all".
func (s *PostService) UpdatePost(
	userID, postID, content string,
	refs []MediaRef,
) (*models.PostResponse, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil, ErrUnauthorizedPost
	}

	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return nil, ErrPostNotFound
	}

	trimmedContent, err := normalizeContent(content)
	if err != nil {
		return nil, err
	}

	existing, err := s.postRepo.GetByID(trimmedPostID)
	if err != nil {
		return nil, err
	}

	if existing.UserID != trimmedUserID {
		return nil, ErrNotPostOwner
	}

	// Editing only the caption is the common case, so an absent photo list has
	// to mean "leave them alone". Re-resolving the post's own public IDs rather
	// than copying its rows keeps a single path into the database: whatever is
	// written went through the same validation as a brand new photo.
	if refs == nil {
		refs = make([]MediaRef, 0, len(existing.Media))
		for _, item := range existing.Media {
			refs = append(refs, MediaRef{
				PublicID: item.PublicID,
				Width:    item.Width,
				Height:   item.Height,
			})
		}
	}

	media, err := s.media.resolve(trimmedUserID, refs)
	if err != nil {
		return nil, err
	}

	// A post must still carry something after the edit
	if trimmedContent == "" && len(media) == 0 {
		return nil, ErrEmptyPost
	}

	// Captured before the write, because afterwards the old rows are gone
	dropped := droppedPublicIDs(existing.Media, media)

	if err := s.postRepo.Update(
		trimmedPostID,
		trimmedUserID,
		trimmedContent,
		media,
	); err != nil {
		return nil, err
	}

	// A photo the user removed is no longer reachable from any page, so the
	// stored file should go too
	s.releasePhotos(context.Background(), dropped)

	return s.postRepo.GetByID(trimmedPostID)
}
//  Comparing by public ID rather than by position means
// reordering photos drops nothing.
func droppedPublicIDs(before, after []models.PostMedia) []string {
	kept := make(map[string]bool, len(after))
	for _, item := range after {
		kept[item.PublicID] = true
	}

	var dropped []string
	for _, item := range before {
		if !kept[item.PublicID] {
			dropped = append(dropped, item.PublicID)
		}
	}
	return dropped
}

// releasePhotos destroys stored files that no post references any more.
//
// Call it only after the database change has committed: the reference count
// below reads post_media as it now stands, so a public ID with no rows left is
// genuinely an orphan, while one another post still uses is left untouched.

// Failures are logged, not returned. By this point the user's edit or deletion
// has already succeeded, so surfacing an error here would report a failure for
// work that was done; the real cost of giving up is an unused file left in
// Cloudinary.
func (s *PostService) releasePhotos(ctx context.Context, publicIDs []string) {
	if s.photoDeleter == nil {
		return
	}

	for _, publicID := range publicIDs {
		if strings.TrimSpace(publicID) == "" {
			continue
		}

		if err := s.photoDeleter.DeletePhoto(ctx, publicID); err != nil {
			log.Printf("photo cleanup: deleting %q from Cloudinary: %v", publicID, err)
		}
	}
}

// DeletePost permanently removes a post the user owns
func (s *PostService) DeletePost(userID, postID string) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return ErrUnauthorizedPost
	}

	trimmedPostID := strings.TrimSpace(postID)
	if trimmedPostID == "" {
		return ErrPostNotFound
	}

	existing, err := s.postRepo.GetByID(trimmedPostID)
	if err != nil {
		return err
	}
	if existing.UserID != trimmedUserID {
		return ErrNotPostOwner
	}

	// Read the photo list while the rows still exist: deleting the post
	// cascades post_media away, taking the public IDs with it.
	media, err := s.postRepo.GetMediaByPostID(trimmedPostID)
	if err != nil {
		return err
	}

	if err := s.postRepo.Delete(trimmedPostID, trimmedUserID); err != nil {
		return err
	}

	publicIDs := make([]string, 0, len(media))
	for _, item := range media {
		publicIDs = append(publicIDs, item.PublicID)
	}
	s.releasePhotos(context.Background(), publicIDs)

	return nil
}
