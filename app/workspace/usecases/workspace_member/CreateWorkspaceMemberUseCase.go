package workspace_member

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	userRepository "mize.app/app/user/repository"
	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	workspace_constants "mize.app/constants/workspace"
	"mize.app/logger"
	"mize.app/utils"
)

func CreateWorkspaceMemberUseCase(ctx *gin.Context, workspace_id *string, name *string, admin bool) (*string, error) {
	userRepo := userRepository.GetUserRepo()
	user, err := userRepo.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]interface{}{
		"profileImage":          1,
		"profileImageThumbnail": 1,
	}))
	if err != nil {
		logger.Error(errors.New("could not fetch user to create workspace member"), zap.Error(err))
		return nil, errors.New("could not join workspace")
	}
	workspaceMemberRepo := repository.GetWorkspaceMember()
	workspace_id_hex := *utils.HexToMongoId(ctx, *workspace_id)
	payload := models.WorkspaceMember{
		WorkspaceId:   workspace_id_hex,
		Username:      ctx.GetString("Username"),
		Firstname:     ctx.GetString("Firstname"),
		Lastname:      ctx.GetString("Lastname"),
		WorkspaceName: *name,
		UserId:        *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		Admin:         admin,
		AdminAccess: func() []workspace_constants.AdminAccessType {
			if admin {
				return []workspace_constants.AdminAccessType{workspace_constants.AdminAccess.FULL_ACCESS}
			}
			return []workspace_constants.AdminAccessType{}
		}(),
		ProfileImage:          user.ProfileImage,
		ProfileImageThumbNail: user.ProfileImageThumbNail,
		JoinDate:              time.Now().Unix()}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	member, err := workspaceMemberRepo.CreateOne(payload)
	id := member.Id.Hex()
	return &id, err
}
