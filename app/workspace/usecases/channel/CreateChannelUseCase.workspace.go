package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateChannelUseCase(ctx *gin.Context, payload workspaceModel.Channel) (*string, error) {
	workspace := workspaceRepo.WorkspaceRepository.FindOneByFilter(ctx, map[string]interface{}{
		"_id": payload.WorkspaceId.Hex(),
	})
	has_access := false
	var err error
	if workspace.CreatedBy.Hex() == ctx.GetString("UserId") {
		has_access = true
	} else {
		for _, admin := range workspace.Admins {
			if admin.Hex() == ctx.GetString("UserId") {
				has_access = true
				break
			}
		}
	}
	if !has_access {
		err = errors.New("only admins can create channels")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return nil, err
	}
	id, err := primitive.ObjectIDFromHex(ctx.GetString("UserId"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return nil, err
	}
	payload.CreatedBy = id
	result := workspaceRepo.ChannelRepository.CreateOne(ctx, &payload)
	return result, nil
}
