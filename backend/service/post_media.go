package service

import (
	"backend/models"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	MaxPostPhotos = 4
	maxMediaDimension = 20000
	mediaTokenPattern = "[0-9a-f]{16}"
)

var (
	ErrTooManyPhotos      = fmt.Errorf("a post can have at most %d photos", MaxPostPhotos)
	ErrInvalidMediaRef    = errors.New("photo reference is not a valid upload")
	ErrMediaNotConfigured = errors.New("media uploads are not configured")
)

type MediaRef struct {
	PublicID string
	Width    int
	Height   int
}

type mediaResolver struct {
	cloudName    string
	uploadFolder string
}

func (m mediaResolver) publicIDPattern(userID string) (*regexp.Regexp, error) {
	optionalFolder := ""
	if m.uploadFolder != "" {
		optionalFolder = "(?:" + regexp.QuoteMeta(m.uploadFolder+"/") + ")?"
	}
	owned := regexp.QuoteMeta(fmt.Sprintf("posts/%s/", userID))
	return regexp.Compile("^" + optionalFolder + owned + mediaTokenPattern + "$")
}

func (m mediaResolver) resolve(userID string, refs []MediaRef) ([]models.PostMedia, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	if len(refs) > MaxPostPhotos {
		return nil, ErrTooManyPhotos
	}
	if m.cloudName == "" {
		return nil, ErrMediaNotConfigured
	}

	pattern, err := m.publicIDPattern(strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(refs))
	out := make([]models.PostMedia, 0, len(refs))

	for i, ref := range refs {
		publicID := strings.TrimSpace(ref.PublicID)

		if !pattern.MatchString(publicID) {
			return nil, ErrInvalidMediaRef
		}
		if seen[publicID] {
			return nil, ErrInvalidMediaRef
		}
		seen[publicID] = true

		out = append(out, models.PostMedia{
			Ordinal:   i,
			URL:       m.deliveryURL(publicID),
			PublicID:  publicID,
			MediaType: "image",
			Width:     clampDimension(ref.Width),
			Height:    clampDimension(ref.Height),
		})
	}
	return out, nil
}

func (m mediaResolver) deliveryURL(publicID string) string {
	return fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/f_auto,q_auto/%s", m.cloudName, publicID)
}

func clampDimension(value int) int {
	if value < 0 {
		return 0
	}
	if value > maxMediaDimension {
		return maxMediaDimension
	}
	return value
}
