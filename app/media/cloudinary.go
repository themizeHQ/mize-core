package media

import (
	"context"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"mize.app/utils"
)

func UploadToCloudinary(ctx *gin.Context, data multipart.File, folder string, public_id *string) (*Upload, error) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		return nil, err
	}
	uploadParam, err := cld.Upload.Upload(c, data, uploader.UploadParams{
		Folder: func() string {
			if public_id == nil || *public_id == "" {
				return folder
			}
			return ""
		}(),
		Format:     "png",
		Invalidate: true,
		PublicID: func() string {
			if public_id == nil {
				return ""
			}
			return *public_id
		}(),
	})

	if err != nil {
		return nil, err
	}
	return &Upload{
		Url:      uploadParam.SecureURL,
		Bytes:    uploadParam.Bytes,
		PublicID: uploadParam.PublicID,
		Service:  "CLOUDINARY",
		UploadBy: *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, nil
}
