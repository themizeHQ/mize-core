package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"mize.app/app/auth"
	"mize.app/middlewares"
	"mize.app/server_response"

	appControllers "mize.app/app/application/controllers"
	userControllers "mize.app/app/user/controllers"
	workspaceControllers "mize.app/app/workspace/controllers"
)

func main() {
	// loads all env vars from current dir
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	server := gin.Default()

	// start all services
	StartServices()

	defer func() {
		// clean up resources
		CleanUp()
	}()

	// set up routing
	v1 := server.Group("/api/v1")
	{
		userV1 := v1.Group("/user")
		{
			userV1.POST("/create", userControllers.CacheUser)

			userV1.POST("/verify", userControllers.VerifyUser)

			userV1.POST("/login", userControllers.LoginUser)
		}

		appV1 := v1.Group("/application", middlewares.AuthenticationMiddleware(false))
		{
			appV1.POST("/create", appControllers.CreateApplication)
		}

		workspaceV1 := v1.Group("/workspace")
		{
			workspaceV1.POST("/create", middlewares.AuthenticationMiddleware(false), workspaceControllers.CreateWorkspace)

			workspaceV1.GET("/fetch", middlewares.AuthenticationMiddleware(false), workspaceControllers.FetchUserWorkspaces)

			workspaceV1.POST("/invite", middlewares.AuthenticationMiddleware(true), workspaceControllers.InviteToWorkspace)

			workspaceV1.PUT("/invite/:inviteId/reject", middlewares.AuthenticationMiddleware(false), workspaceControllers.RejectWorkspaceInvite)

			workspaceV1.PUT("/invite/:inviteId/accept", middlewares.AuthenticationMiddleware(false), workspaceControllers.AcceptWorkspaceInvite)
		}

		channelV1 := v1.Group("/channel", middlewares.AuthenticationMiddleware(true))
		{
			channelV1.POST("/create", workspaceControllers.CreateChannel)

			channelV1.POST("/join/:id", workspaceControllers.CreateChannelMember)
		}

		// channelMemV1 := v1.GET("/channel-member", middlewares.AuthenticationMiddleware(true))
		// {
		// }

		authV1 := v1.Group("/auth")
		{
			authV1.GET("/generate-access-token", auth.GenerateAccessTokenFromRefresh)

			authV1.GET("/resend-otp", auth.ResendOtp)
		}
	}

	server.GET("/who-is-the-goat", func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusOK, "Lionel Messi is the GOAT!", true, nil)
	})

	server.NoRoute(func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusNotFound, "This route does not exist", false, nil)
	})

	server.Run(":" + os.Getenv("PORT"))

}
