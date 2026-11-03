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

const (
	maxImageUploadBytes = 5 * 1024 * 1024
	multipartEnvelopeSlack = 1 << 20
	multipartMemoryBudget = 1 << 20
	cloudinaryUploadTimeout = 25 * time.Second
	postMediaTag = "xclone_post_media"
)

var errCloudinaryInit = errors.New("cloudinary init error")

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

type UploadController struct {
	cloudName    string
	apiKey       string
	apiSecret    string
	uploadFolder string
	jwtSecret    string
}

func NewUploadController(cloudName, apiKey, apiSecret, uploadFolder, jwtSecret string) *UploadController {
	return &UploadController{
		cloudName:    cloudName,
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		uploadFolder: uploadFolder,
		jwtSecret:    jwtSecret,
	}
}

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

func (c *UploadController) requireCloudinary(w http.ResponseWriter) bool {
	if c.cloudName == "" || c.apiKey == "" || c.apiSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "Cloudinary credentials are not configured")
		return false
	}
	return true
}

func enforceUploadLimit(w http.ResponseWriter, r *http.Request, what string) bool {
	tooLarge := fmt.Sprintf("%s must be %dMB or smaller", what, maxImageUploadBytes/(1024*1024))

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

func removeMultipartTempFiles(r *http.Request) {
	if r.MultipartForm != nil {
		_ = r.MultipartForm.RemoveAll()
	}
}

func openImageUpload(w http.ResponseWriter, r *http.Request, field, what string) (multipart.File, *multipart.FileHeader, bool) {
	file, header, err := r.FormFile(field)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("%s file is required", what))
		return nil, nil, false
	}

	if header.Size > maxImageUploadBytes {
		_ = file.Close()
		writeError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("%s must be %dMB or smaller", what, maxImageUploadBytes/(1024*1024)))
		return nil, nil, false
	}

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

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		writeError(w, http.StatusInternalServerError, "failed to reset uploaded file")
		return nil, nil, false
	}
	return file, header, true
}

func (c *UploadController) uploadImage(ctx context.Context, src io.Reader, params uploader.UploadParams) (*uploader.UploadResult, error) {
	cld, err := cloudinary.NewFromParams(c.cloudName, c.apiKey, c.apiSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCloudinaryInit, err)
	}

	params.Folder = c.uploadFolder
	result, err := cld.Upload.Upload(ctx, src, params)
	if err != nil {
		return nil, err
	}

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

func writeUploadError(w http.ResponseWriter, err error) {
	if errors.Is(err, errCloudinaryInit) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeError(w, http.StatusBadGateway, fmt.Sprintf("cloudinary upload error: %v", err))
}

func (c *UploadController) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
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

	token, err := utils.GenerateRandomToken(8)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate a photo identifier")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), cloudinaryUploadTimeout)
	defer cancel()

	uploadResult, err := c.uploadImage(ctx, file, uploader.UploadParams{
		PublicID:  fmt.Sprintf("posts/%s/%s", claims.UserID, token),
		Overwrite: boolPtr(false),
		Transformation: "c_limit,w_2048,h_2048/f_auto,q_auto",
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
