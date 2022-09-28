package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mize.app/app/teams/models"
	"mize.app/app/teams/usecases"
	"mize.app/app_errors"
	"mize.app/server_response"
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
