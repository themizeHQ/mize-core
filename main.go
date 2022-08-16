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
	notificationControllers "mize.app/app/notification/controllers"
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

			userV1.GET("/profile", middlewares.AuthenticationMiddleware(false), userControllers.FetchProfile)

			userV1.GET("/fetch-user/:id", middlewares.AuthenticationMiddleware(false), userControllers.FetchUsersProfile)
		}

		notificationV1 := v1.Group("/notification", middlewares.AuthenticationMiddleware(false))
		{
			notificationV1.GET("/fetch", notificationControllers.FetchUserNotifications)

			notificationV1.DELETE("/delete", notificationControllers.DeleteNotifications)
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

			workspaceV1.GET("/invite/fetch", middlewares.AuthenticationMiddleware(true), workspaceControllers.FetchWorkspaceInvites)
		}

		channelV1 := v1.Group("/channel", middlewares.AuthenticationMiddleware(true))
		{
			channelV1.POST("/create", workspaceControllers.CreateChannel)

			channelV1.GET("/fetch", workspaceControllers.FetchChannels)

			channelV1.POST("/join/:id", workspaceControllers.CreateChannelMember)

			channelV1.DELETE("/delete/:id", workspaceControllers.DeleteChannel)
		}

		authV1 := v1.Group("/auth")
		{
			authV1.GET("/generate-access-token", auth.GenerateAccessTokenFromRefresh)

			authV1.GET("/resend-otp", auth.ResendOtp)
		}
	}

	server.GET("/who-is-the-goat", func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusOK, "lionel Messi is the GOAT!", true, nil)
	})

	server.NoRoute(func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusNotFound, "this route does not exist", false, nil)
	})

	server.Run(":" + os.Getenv("PORT"))

}
