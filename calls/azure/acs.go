package azure

import (
	"encoding/json"
	"errors"
	"os"

	"mize.app/network"
)

type acsUserId struct {
	CommunicationUserId string `json:"communicationUserId"`
}

type ACSUserToken struct {
	Token     string    `json:"token"`
	User      acsUserId `json:"user"`
	ExpiresOn string    `json:"expiresOn"`
}

type res struct {
	Data ACSUserToken `json:"data"`
}

func GenerateUserAndToken() (data *ACSUserToken, err error) {
	acs := network.NetworkController{
		BaseUrl: "https://acs-authenticator.azurewebsites.net/api/issue-acs-token",
	}

	response, err := acs.Post("/", nil, nil, &map[string]string{
		"code": os.Getenv("ACS_GEN_USER_AND_TOKEN_CODE"),
	})
	if err != nil {
		return nil, errors.New("could not register user to acs")
	}
	var r res
	json.Unmarshal([]byte(*response), &r)
	return &r.Data, nil
}
