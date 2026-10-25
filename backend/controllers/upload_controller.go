package controllers

import (
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Shared limits for every image upload endpoint
const (
	// maxImageUploadBytes is the real per-file cap. It is enforced with
	// http.MaxBytesReader and re-checked against the multipart part header.
	maxImageUploadBytes = 5 * 1024 * 1024

	// multipartEnvelopeSlack leaves room for the boundaries and part headers
	// that wrap the file, so a file exactly at the cap is not rejected.
	multipartEnvelopeSlack = 1 << 20

	// multipartMemoryBudget is how much of the form is held in memory before
	// the remainder spills to temporary files. It is NOT a size limit.
	multipartMemoryBudget = 1 << 20

	// cloudinaryUploadTimeout bounds a single upload to Cloudinary
	cloudinaryUploadTimeout = 25 * time.Second

	// postMediaTag makes post photos findable in one Cloudinary Admin API call,
	// which is what a future cleanup of abandoned uploads needs.
	postMediaTag = "xclone_post_media"
)

// errCloudinaryInit separates "we could not build a client" (our own
// misconfiguration) from "the upload failed" (an upstream problem).
var errCloudinaryInit = errors.New("cloudinary init error")

// allowedImageTypes lists the types accepted by every image upload. The value
// compared against it is sniffed from the file, never taken from the client.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// a structure to upload image
type UploadController struct {
	cloudName    string
	apiKey       string
	apiSecret    string
	uploadFolder string
	jwtSecret    string
}

// NewUploadController creates a new instance of UploadController/ a consutructor function
func NewUploadController(cloudName, apiKey, apiSecret, uploadFolder, jwtSecret string) *UploadController {
	return &UploadController{
		cloudName:    cloudName,
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		uploadFolder: uploadFolder,
		jwtSecret:    jwtSecret,
	}
}

// requireAuth reads the auth cookie and validates the JWT. It writes the error
// response itself, so callers only need to check the bool.
func (c *UploadController) requireAuth(w http.ResponseWriter, r *http.Request) (*utils.JWTClaims, bool) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return nil, false
	}

	claims, err := utils.ValidateJWT(cookie.Value, c.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return nil, false
	}
	return claims, true
}

// requireCloudinary reports whether Cloudinary credentials are configured
func (c *UploadController) requireCloudinary(w http.ResponseWriter) bool {
	if c.cloudName == "" || c.apiKey == "" || c.apiSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "Cloudinary credentials are not configured")
		return false
	}
	return true
}

// enforceUploadLimit caps the request body before any of it is buffered, then
// parses the multipart form. It writes the error response itself.
//
// The order matters. ParseMultipartForm's argument is only an in-memory budget:
// anything larger spills to temporary files on disk and the call still
// succeeds, so it can never be a size limit on its own. http.MaxBytesReader is
// the actual cap, and it has to wrap the body before parsing begins.
func enforceUploadLimit(w http.ResponseWriter, r *http.Request, what string) bool {
	tooLarge := fmt.Sprintf("%s must be %dMB or smaller", what, maxImageUploadBytes/(1024*1024))

	// A declared length above the cap can be refused without reading anything.
	// Content-Length is client-supplied and is -1 for chunked bodies, so this
	// is an optimisation and never the guard itself.
	if r.ContentLength > maxImageUploadBytes+multipartEnvelopeSlack {
		writeError(w, http.StatusRequestEntityTooLarge, tooLarge)
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadBytes+multipartEnvelopeSlack)

	if err := r.ParseMultipartForm(multipartMemoryBudget); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, tooLarge)
			return false
		}
		writeError(w, http.StatusBadRequest, "invalid multipart form data")
		return false
	}
	return true
}

// removeMultipartTempFiles deletes any parts that spilled to disk instead of
// leaving them for the runtime to collect at some later point
func removeMultipartTempFiles(r *http.Request) {
	if r.MultipartForm != nil {
		_ = r.MultipartForm.RemoveAll()
	}
}

// openImageUpload pulls the named file out of an already-parsed multipart form,
// enforces the size cap on the part itself, and verifies the image type by
// sniffing the content. The returned file is rewound to its first byte and the
// caller is responsible for closing it.
func openImageUpload(w http.ResponseWriter, r *http.Request, field, what string) (multipart.File, *multipart.FileHeader, bool) {
	file, header, err := r.FormFile(field)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("%s file is required", what))
		return nil, nil, false
	}

	// The envelope slack above deliberately allows a little more than the cap,
	// so the part's own declared size is checked as well.
	if header.Size > maxImageUploadBytes {
		_ = file.Close()
		writeError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("%s must be %dMB or smaller", what, maxImageUploadBytes/(1024*1024)))
		return nil, nil, false
	}

	// Identify the type from the first bytes of the file. The client-supplied
	// Content-Type is ignored, so a text file renamed .jpg is still rejected.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		_ = file.Close()
		writeError(w, http.StatusBadRequest, "failed to read uploaded file")
		return nil, nil, false
	}

	if !allowedImageTypes[http.DetectContentType(buf[:n])] {
		_ = file.Close()
		writeError(w, http.StatusBadRequest, "invalid image type: must be JPEG, PNG, WEBP, or GIF")
		return nil, nil, false
	}

	// Rewind so the whole file can be streamed to Cloudinary
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		writeError(w, http.StatusInternalServerError, "failed to reset uploaded file")
		return nil, nil, false
	}
	return file, header, true
}

// uploadImage streams an image to Cloudinary. The folder always comes from
// config; everything else is the caller's choice, so the two endpoints keep
// their own public IDs and transformations.
func (c *UploadController) uploadImage(ctx context.Context, src io.Reader, params uploader.UploadParams) (*uploader.UploadResult, error) {
	cld, err := cloudinary.NewFromParams(c.cloudName, c.apiKey, c.apiSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCloudinaryInit, err)
	}

	params.Folder = c.uploadFolder
	// src is an io.Reader, so the file is streamed straight through instead of
	// being read into memory first
	result, err := cld.Upload.Upload(ctx, src, params)
	if err != nil {
		return nil, err
	}

	// The SDK reports an API-level rejection in the response body rather than
	// as a returned error, so a nil error is not enough to conclude success.
	// Without these two checks a refused upload looks like a 200 carrying an
	// empty URL, and the caller stores "" as the image.
	if result.Error.Message != "" {
		return nil, fmt.Errorf("cloudinary rejected the upload: %s", result.Error.Message)
	}
	if result.SecureURL == "" || result.PublicID == "" {
		return nil, errors.New("cloudinary returned no asset for the upload")
	}
	return result, nil
}

func (c *UploadController) deleteImage(
	ctx context.Context,
	publicID string,
) error {
	cld, err := cloudinary.NewFromParams(
		c.cloudName,
		c.apiKey,
		c.apiSecret,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", errCloudinaryInit, err)
	}

	result, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return err
	}
	if result.Error.Message != "" {
		return fmt.Errorf(
			"Cloudinary rejected deletion: %s",
			result.Error.Message,
		)
	}
	return nil
}

// writeUploadError maps an upload failure onto a status code
func writeUploadError(w http.ResponseWriter, err error) {
	if errors.Is(err, errCloudinaryInit) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeError(w, http.StatusBadGateway, fmt.Sprintf("cloudinary upload error: %v", err))
}

// UploadAvatarHandler handles POST /api/upload/avatar
func (c *UploadController) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	//Only post req allowed
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	if !c.requireCloudinary(w) {
		return
	}

	if !enforceUploadLimit(w, r, "avatar") {
		return
	}
	defer removeMultipartTempFiles(r)

	file, header, ok := openImageUpload(w, r, "avatar", "avatar")
	if !ok {
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(r.Context(), cloudinaryUploadTimeout)
	defer cancel()

	uploadResult, err := c.uploadImage(ctx, file, uploader.UploadParams{
		//create a unique filename for cloudinary public id
		PublicID:       fmt.Sprintf("avatar_%s_%d", claims.UserID, time.Now().Unix()),
		Overwrite:      boolPtr(true),
		Transformation: "c_thumb,g_face,w_400,h_400/f_auto,q_auto",
	})
	if err != nil {
		writeUploadError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Avatar uploaded successfully",
		Data: map[string]interface{}{
			"avatar_url": uploadResult.SecureURL,
			"filename":   header.Filename,
			"size":       header.Size,
		},
	})
}

// UploadPostMediaHandler handles POST /api/upload/post-media
func (c *UploadController) UploadPostMediaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	if !c.requireCloudinary(w) {
		return
	}

	if !enforceUploadLimit(w, r, "photo") {
		return
	}
	defer removeMultipartTempFiles(r)

	file, header, ok := openImageUpload(w, r, "photo", "photo")
	if !ok {
		return
	}
	defer file.Close()

	// A random suffix rather than a timestamp: four photos picked at once share
	// the same second, and the avatar scheme would make them overwrite each
	// other instead of becoming four separate assets.
	token, err := utils.GenerateRandomToken(8)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate a photo identifier")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), cloudinaryUploadTimeout)
	defer cancel()

	uploadResult, err := c.uploadImage(ctx, file, uploader.UploadParams{
		// The uploader's ID is part of the public ID, which is how the create
		// endpoint later verifies that the caller uploaded this photo itself.
		PublicID:  fmt.Sprintf("posts/%s/%s", claims.UserID, token),
		Overwrite: boolPtr(false),
		// c_limit only scales down, and only when the image is larger than the
		// box, so the aspect ratio of a landscape or portrait photo survives.
		Transformation: "c_limit,w_2048,h_2048/f_auto,q_auto",
		// A second line of defence behind the content sniffer
		AllowedFormats: api.CldAPIArray{"jpg", "jpeg", "png", "webp", "gif"},
		Tags:           api.CldAPIArray{postMediaTag},
	})
	if err != nil {
		writeUploadError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Photo uploaded successfully",
		Data: map[string]interface{}{
			// public_id is the only field POST /api/posts accepts. url is here
			// for debugging and verification; the composer previews the local
			// file instead.
			"public_id": uploadResult.PublicID,
			"width":     uploadResult.Width,
			"height":    uploadResult.Height,
			"url":       uploadResult.SecureURL,
			"filename":  header.Filename,
			"size":      header.Size,
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}
