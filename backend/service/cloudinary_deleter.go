package service

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryPhotoDeleter struct {
	cloudName string
	apiKey    string
	apiSecret string
}

func NewCloudinaryPhotoDeleter(
	cloudName, apiKey, apiSecret string,

) *CloudinaryPhotoDeleter {
	return &CloudinaryPhotoDeleter{
		cloudName: cloudName,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (c *CloudinaryPhotoDeleter) DeletePhoto(
	ctx context.Context,
	publicID string,
) error {
	cld, err := cloudinary.NewFromParams(
		c.cloudName,
		c.apiKey,
		c.apiSecret,
	)
	if err != nil {
		return fmt.Errorf("intialize Cloudinary: %W", err)
	}

	result, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("delete Cloudinary photo: %w", err)
	}

	if result.Error.Message != "" {
		return fmt.Errorf("Cloudinary rejected deletion: %s", result.Error.Message)
	}
	return nil
}
