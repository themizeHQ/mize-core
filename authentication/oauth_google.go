package authentication

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func SetUpConfig() *oauth2.Config {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID_WEB"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET_WEB"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/user.phonenumbers.read",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return conf
}
