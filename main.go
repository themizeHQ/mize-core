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
	conversationControllers "mize.app/app/conversation/controllers"
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
			userV1.GET("/profile", middlewares.AuthenticationMiddleware(false, false), userControllers.FetchProfile)

			userV1.GET("/fetch-user/:id", middlewares.AuthenticationMiddleware(false, false), userControllers.FetchUsersProfile)

			userV1.PUT("/update", middlewares.AuthenticationMiddleware(false, false), userControllers.UpdateUserData)
		}

		notificationV1 := v1.Group("/notification")
		{
			notificationV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), notificationControllers.FetchUserNotifications)

			notificationV1.DELETE("/delete", middlewares.AuthenticationMiddleware(false, false), notificationControllers.DeleteNotifications)

			// alert routes
			notificationV1.POST("/alert/send", middlewares.AuthenticationMiddleware(true, true), notificationControllers.SendAlert)
		}

		appV1 := v1.Group("/application", middlewares.AuthenticationMiddleware(false, false))
		{
			appV1.POST("/create", appControllers.CreateApplication)
		}

		workspaceV1 := v1.Group("/workspace")
		{
			workspaceV1.POST("/create", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.CreateWorkspace)

			workspaceV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.FetchUserWorkspaces)

			workspaceV1.POST("/invite", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.InviteToWorkspace)

			workspaceV1.PUT("/invite/:inviteId/reject", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.RejectWorkspaceInvite)

			workspaceV1.PUT("/invite/:inviteId/accept", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.AcceptWorkspaceInvite)

			workspaceV1.GET("/invite/fetch", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.FetchWorkspaceInvites)
		}

		channelV1 := v1.Group("/channel")
		{
			channelV1.POST("/create", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.CreateChannel)

			channelV1.GET("/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchChannels)

			channelV1.POST("/join/:id", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.CreateChannelMember)

			channelV1.DELETE("/delete/:id", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.DeleteChannel)

			channelV1.GET("/members/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchChannelMembers)

			channelV1.DELETE("/leave", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.LeaveChannel)

			channelV1.PUT("/pin", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.PinChannel)

			channelV1.PUT("/unpin", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.UnPinChannel)

			channelV1.GET("/pinned/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchPinnedChannels)

			channelV1.POST("/add/username", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.AdminAddUserByUsername)
		}

		messageV1 := v1.Group("/message")
		{
			messageV1.POST("/send", middlewares.AuthenticationMiddleware(true, false), conversationControllers.SendMessage)

			messageV1.GET("/fetch", middlewares.AuthenticationMiddleware(true, false), conversationControllers.FetchMessages)
		}

		authV1 := v1.Group("/auth")
		{
			authV1.POST("/create", auth.CacheUserUseCase)

			authV1.POST("/verify", auth.VerifyAccountUseCase)

			authV1.POST("/login", auth.LoginUser)

			authV1.PUT("/update-password", middlewares.AuthenticationMiddleware(false, false), auth.UpdateLoggedInUsersPassword)

			authV1.GET("/generate-access-token", auth.GenerateAccessTokenFromRefresh)

			authV1.GET("/resend-otp", auth.ResendOtp)

			// centrifugo
			authV1.GET("/realtime/authenticate", middlewares.AuthenticationMiddleware(false, false), auth.GenerateCentrifugoToken)
		}
	}

	server.GET("/who-is-the-goat", func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusOK, "lionel Messi is the GOAT!", true, nil)
	})

	server.NoRoute(func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusNotFound, "this route does not exist", false, nil)
	})

	server.Run(os.Getenv("PORT"))

}
