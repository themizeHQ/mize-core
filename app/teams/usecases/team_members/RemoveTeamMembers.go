package teammembers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	teamsRepo "mize.app/app/teams/repository"
	types "mize.app/app/teams/types"
	"mize.app/utils"
)

func RemoveTeamMembers(ctx *gin.Context, ids types.IDArray, teamId string) error {
	teamMemberRepo := teamsRepo.GetTeamMemberRepo()
	var err error
	teamMemberRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		err1 := make(chan error)
		go func(err chan error) {
			_, e := teamMemberRepo.DeleteMany(ctx, map[string]interface{}{
				"_id": map[string]interface{}{
					"$in": ids,
				},
				"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			})
			if err != nil {
				err <- e
				return
			}
			err <- nil
		}(err1)

		teamRepo := teamsRepo.GetTeamRepo()
		err2 := make(chan error)
		go func(err chan error) {
			_, e := teamRepo.UpdateWithOperator(map[string]interface{}{
				"_id":         *utils.HexToMongoId(ctx, teamId),
				"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			}, map[string]interface{}{
				"$inc": map[string]interface{}{
					"membersCount": -len(ids),
				}},
			)
			if err != nil {
				err <- e
				return
			}
			err <- nil
		}(err2)
		if err != nil {
			(*sc).AbortTransaction(*c)
		}
		return err
	})
	return nil
}
