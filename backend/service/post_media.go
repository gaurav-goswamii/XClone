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

	// maxMediaDimension bounds the client-reported pixel sizes. They are only
	// used to reserve layout space, so an implausible value is clamped rather
	// than rejected.
	maxMediaDimension = 20000

	// mediaTokenPattern matches the random suffix the upload endpoint generates
	// with utils.GenerateRandomToken(8): 8 bytes rendered as 16 hex characters.
	mediaTokenPattern = "[0-9a-f]{16}"
)

var (
	ErrTooManyPhotos      = fmt.Errorf("a post can have at most %d photos", MaxPostPhotos)
	ErrInvalidMediaRef    = errors.New("photo reference is not a valid upload")
	ErrMediaNotConfigured = errors.New("media uploads are not configured")
)

// MediaRef is an untrusted, client-supplied reference to an already-uploaded
// asset. It deliberately has no URL field: the client names an asset, and the
// server decides what URL that name resolves to.
type MediaRef struct {
	PublicID string
	Width    int
	Height   int
}

// mediaResolver turns untrusted references into trusted, persistable rows. It
// holds the Cloudinary cloud and folder because "where our media lives" is the
// rule it validates against.
type mediaResolver struct {
	cloudName    string
	uploadFolder string
}

// publicIDPattern builds the only shape of public ID this user is allowed to
// claim, for example: ^(xclone/)?posts/usr_abc/[0-9a-f]{16}$
//
// regexp.QuoteMeta is load-bearing. The user ID is interpolated into the
// pattern, so without escaping, an ID containing a regex metacharacter would
// change what the pattern means instead of being matched literally.
//
// The folder prefix is optional because Cloudinary returns the public ID in
// two different shapes depending on the account. With classic fixed folders it
// prepends the folder ("xclone/posts/..."); with dynamic folders the folder is
// tracked separately and the public ID stays bare ("posts/..."). Accepting
// both keeps every guarantee below intact and means the endpoint works without
// knowing which mode the account is in.
func (m mediaResolver) publicIDPattern(userID string) (*regexp.Regexp, error) {
	optionalFolder := ""
	if m.uploadFolder != "" {
		optionalFolder = "(?:" + regexp.QuoteMeta(m.uploadFolder+"/") + ")?"
	}
	owned := regexp.QuoteMeta(fmt.Sprintf("posts/%s/", userID))
	return regexp.Compile("^" + optionalFolder + owned + mediaTokenPattern + "$")
}

// resolve validates every reference and returns rows ready to persist.
//
// This is the security boundary for post photos. POST /api/posts carries JSON,
// so a caller can put any string in a media field. Accepting only an opaque
// public ID and formatting the URL here means a caller cannot point a post at
// a host we do not control, nor at another user's upload.
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

		// The anchored pattern is what rejects a foreign host, a path escape
		// like ../, an avatar asset, and another user's photo, all at once.
		if !pattern.MatchString(publicID) {
			return nil, ErrInvalidMediaRef
		}
		// Refuse rather than silently drop: quietly posting 3 of the 4 photos
		// the user picked would be harder to understand than an error.
		if seen[publicID] {
			return nil, ErrInvalidMediaRef
		}
		seen[publicID] = true

		out = append(out, models.PostMedia{
			// Assigned from the position in the request, never taken from the
			// client, so gaps and duplicate slots cannot be submitted at all.
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

// deliveryURL builds the public URL for a stored asset. f_auto and q_auto let
// Cloudinary pick the best format and quality per browser at delivery time.
func (m mediaResolver) deliveryURL(publicID string) string {
	return fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/f_auto,q_auto/%s", m.cloudName, publicID)
}

// clampDimension keeps a client-reported pixel size within a sane range
func clampDimension(value int) int {
	if value < 0 {
		return 0
	}
	if value > maxMediaDimension {
		return maxMediaDimension
	}
	return value
}
