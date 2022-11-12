package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/teams/models"
	teamsRepository "mize.app/app/teams/repository"
	"mize.app/app/teams/usecases"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/server_response"
	"mize.app/utils"
)

func CreateTeam(ctx *gin.Context) {
	raw_form, err := ctx.MultipartForm()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in team data"), StatusCode: http.StatusBadRequest})
		return
	}
	parsed_form := map[string]interface{}{}
	for key, value := range raw_form.Value {
		parsed_form[key] = value[0]
	}
	payload := models.Team{}
	pJson, err := json.Marshal(parsed_form)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	err = json.Unmarshal(pJson, &payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	file, fileHeader, err := ctx.Request.FormFile("media")
	if err != nil {
		if err.Error() != "http: no such file" {
			server_response.Response(ctx, http.StatusBadRequest, err.Error(), false, nil)
			return
		}
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if file == nil {
		err = usecases.CreateTeamUseCase(ctx, &payload, file, "", "")
		if err != nil {
			return
		}
	} else {
		err = usecases.CreateTeamUseCase(ctx, &payload, file, strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1], fileHeader.Filename)
		if err != nil {
			return
		}
	}
	server_response.Response(ctx, http.StatusCreated, "team created", true, nil)
}

// FetchTeams - controller function to fetch teams
func FetchTeams(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	var teams *[]map[string]interface{}

	if authentication.RoleType(ctx.GetString("Role")) == authentication.ADMIN {
		teamsRepo := teamsRepository.GetTeamRepo()
		teams, err = teamsRepo.FindManyStripped(map[string]interface{}{
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		}, &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}, options.Find().SetProjection(
			map[string]interface{}{
				"name":         1,
				"membersCount": 1,
			}))

	} else {
		teamMemberRepo := teamsRepository.GetTeamMemberRepo()
		teams, err = teamMemberRepo.FindManyStripped(map[string]interface{}{
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		}, &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}, options.Find().SetProjection(
			map[string]interface{}{
				"name":        1,
				"memberCount": 1,
			}))

	}
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "could not fetch teams", false, teams)
	}
	server_response.Response(ctx, http.StatusOK, "teams fetched", true, teams)
}
