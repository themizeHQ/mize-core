package teammembers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	teamModels "mize.app/app/teams/models"
	teamsRepo "mize.app/app/teams/repository"
	"mize.app/app/teams/types"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

// CreateTeamMemberUseCase - usecase function to create multiple team members
func CreateTeamMemberUseCase(ctx *gin.Context, teamID string, payload []types.TeamMembersPayload) error {
	teamRepo := teamsRepo.GetTeamRepo()
	teamMembersRepo := teamsRepo.GetTeamMemberRepo()
	team, err := teamRepo.FindById(teamID)
	if err != nil {
		err = errors.New("could not create team members")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	if team == nil {
		err = errors.New("team does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
	members := []teamModels.TeamMembers{}
	wg := sync.WaitGroup{}
	for _, member := range payload {
		wg.Add(1)
		go func(mem types.TeamMembersPayload) {
			defer func() {
				wg.Done()
			}()
			if mem.Type == types.UserType {
				member, err := workspaceMemberRepo.FindById(mem.UserID.Hex())
				if err != nil {
					return
				}
				if member == nil {
					return
				}
				teamMemberExists, err := teamMembersRepo.CountDocs(map[string]interface{}{
					"userId": member.UserId,
					"teamId": *utils.HexToMongoId(ctx, teamID),
				})
				if err != nil {
					return
				}
				fmt.Println(member.UserId)
				if teamMemberExists > 0 {
					return
				}
				members = append(members, teamModels.TeamMembers{
					FirstName:         member.Firstname,
					LastName:          member.Lastname,
					UserId:            member.UserId,
					WorkspaceId:       member.WorkspaceId,
					Username:          member.Username,
					WorkspaceMemberId: member.Id,
					TeamId:            team.Id,
					TeamName:          team.Name,
					ProfileImage:      member.ProfileImage,
					Type:              mem.Type,
				})
			} else if mem.Type == types.TeamType {
				teamMembers, err := teamMembersRepo.FindMany(map[string]interface{}{
					"teamId":      mem.UserID,
					"workspaceId": team.WorkspaceId,
				})
				if err != nil {
					return
				}
				if len(*teamMembers) == 0 {
					return
				}
				for _, teamMember := range *teamMembers {
					parentTeamName := teamMember.TeamName
					parentTeamNameID := teamMember.Id
					teamMember.ParentTeamName = &parentTeamName
					teamMember.ParentTeamID = &parentTeamNameID
					teamMember.TeamId = team.Id
					teamMember.TeamName = team.Name
					teamMember.Type = mem.Type
					members = append(members, teamMember)
				}
			}
		}(member)
	}
	wg.Wait()
	if len(members) == 0 {
		err = errors.New("invalid members selected")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	teamMembersRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		created, e := teamMembersRepo.CreateBulk(members)
		if e != nil {
			return err
		}
		_, err :=
			teamRepo.UpdateWithOperator(map[string]interface{}{
				"_id":         *utils.HexToMongoId(ctx, teamID),
				"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			}, map[string]interface{}{
				"$inc": map[string]interface{}{
					"membersCount": len(*created),
				}})
		if err != nil {
			return err
		}
		if err != nil {
			(*sc).AbortTransaction(*c)
		}
		(*sc).CommitTransaction(*c)
		return err
	})
	return err
}
