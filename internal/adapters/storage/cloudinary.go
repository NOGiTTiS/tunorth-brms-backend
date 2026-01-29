package storage

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryAdapter struct {
	cld *cloudinary.Cloudinary
}

// NewCloudinaryAdapter initializes a new Cloudinary client
// returns error if credentials are invalid or missing
func NewCloudinaryAdapter(cloudName, apiKey, apiSecret string) (*CloudinaryAdapter, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, errors.New("missing cloudinary credentials")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	return &CloudinaryAdapter{cld: cld}, nil
}

// Upload sends the file to Cloudinary and returns the secure URL
func (a *CloudinaryAdapter) Upload(file *multipart.FileHeader, filename string) (string, error) {
	ctx := context.Background()

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Upload to Cloudinary
	resp, err := a.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     filename,
		Tags:         []string{"tunorth-brms"},
		ResourceType: "auto",
		Folder:       "tunorth-brms",
	})

	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
