package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
	"backend/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

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

// UploadAvatarHandler handles POST /api/upload/avatar
func (c *UploadController) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	
	//Only post req allowed
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	//checks for login
	var tokenStr string
	if cookie, err := r.Cookie(AuthCookieName); err == nil && cookie.Value != "" {
		tokenStr = cookie.Value
	}

	if tokenStr == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return
	}

	claims, err := utils.ValidateJWT(tokenStr, c.jwtSecret)

	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return
	}

	// 2. Verify Cloudinary credentials are configured
	if c.cloudName == "" || c.apiKey == "" || c.apiSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "Cloudinary credentials are not configured")
		return
	}

	// 3. Parse multipart form with 5MB limit
	const maxUploadSize = 5 * 1024 * 1024 // 5 MB
	// allows browser to send files
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file size exceeds 5MB limit")
		return
	}
     //get the uploaded file
	file, header, err := r.FormFile("avatar")
    if err != nil {
		writeError(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	//temp space to store first 512 bytes of image 
	buf := make([]byte, 512)
	n, err := file.Read(buf)

	if err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "failed to read uploaded file")
		return
	}
     //to identify the file type
	mimeType := http.DetectContentType(buf[:n])

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}

	if !allowedTypes[mimeType] {
		writeError(w, http.StatusBadRequest, "invalid image type: must be JPEG, PNG, WEBP, or GIF")
		return
	}
//moving back to start of the file to read it completely
_, err = file.Seek(0, io.SeekStart)

if err != nil {
    writeError(w, http.StatusInternalServerError,
        "failed to reset uploaded file")
    return
}

fullBytes, err := io.ReadAll(file)
if err != nil {
    writeError(w, http.StatusInternalServerError,
        "failed to read image data")
    return
}

	// 5. Upload to Cloudinary
	cld, err := cloudinary.NewFromParams(
		c.cloudName,
		c.apiKey,
		c.apiSecret)

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("cloudinary init error: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
    //create a unique filename for cloudinary public id
	publicID := fmt.Sprintf("avatar_%s_%d", claims.UserID, time.Now().Unix())
// configuring the upload
	uploadParams := uploader.UploadParams{
		Folder:         c.uploadFolder,
		PublicID:       publicID,
		Overwrite:      boolPtr(true),
		Transformation: "c_thumb,g_face,w_400,h_400/f_auto,q_auto",
	}
//upload to cloudinary
	uploadResult, err := cld.Upload.Upload(ctx, bytes.NewReader(fullBytes), uploadParams)
	
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("cloudinary upload error: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Avatar uploaded successfully",
		Data: map[string]interface{}{
			"avatar_url": uploadResult.SecureURL,
			"filename":   header.Filename,
			"size":       len(fullBytes),
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}
