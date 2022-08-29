package media

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"mize.app/utils"
)

func UploadToCloudinary(ctx *gin.Context, data multipart.File, folder string, public_id *string) (*Upload, error) {
	fmt.Println("abeg")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println(public_id)
	fmt.Println(folder)
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("hiii")
	fmt.Println(uploader.UploadParams{
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
