package teammembers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	teamsRepo "mize.app/app/teams/repository"
	types "mize.app/app/teams/types"
	"mize.app/utils"
)

func RemoveTeamMembers(ctx *gin.Context, ids types.IDArray, teamId string, removeTeam bool) error {
	teamMemberRepo := teamsRepo.GetTeamMemberRepo()
	teamMemberRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		var deleted int64
		var err error
		if removeTeam {
			deleted, err = teamMemberRepo.DeleteMany(ctx, map[string]interface{}{
				"parentTeamId": map[string]interface{}{
					"$in": ids,
				},
				"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			})
			fmt.Println(ids)
			fmt.Println(deleted)
			if err != nil {
				(*sc).AbortTransaction(*c)
			}
		} else {
			deleted, err = teamMemberRepo.DeleteMany(ctx, map[string]interface{}{
				"_id": map[string]interface{}{
					"$in": ids,
				},
				"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			})
			if err != nil {
				(*sc).AbortTransaction(*c)
			}
		}

		teamRepo := teamsRepo.GetTeamRepo()
		_, err = teamRepo.UpdateWithOperator(map[string]interface{}{
			"_id":         *utils.HexToMongoId(ctx, teamId),
			"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		}, map[string]interface{}{
			"$inc": map[string]interface{}{
				"membersCount": -deleted,
			}},
		)
		if err != nil {
			(*sc).AbortTransaction(*c)
		}
		(*sc).CommitTransaction(*c)
		return err
	})
	return nil
}
