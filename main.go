package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"mize.app/app/auth"
	"mize.app/logger"
	"mize.app/middlewares"
	"mize.app/schedule_manager"
	"mize.app/server_response"

	appControllers "mize.app/app/application/controllers"
	conversationControllers "mize.app/app/conversation/controllers"
	notificationControllers "mize.app/app/notification/controllers"
	scheduleControllers "mize.app/app/schedule/controllers"
	teamControllers "mize.app/app/teams/controllers"
	userControllers "mize.app/app/user/controllers"
	workspaceControllers "mize.app/app/workspace/controllers"
)

func main() {
	// loads all env vars from current dir
	err := godotenv.Load()

	server := gin.Default()

	// CORS
	server.Use(cors.Default())

	// start all services
	StartServices()

	if err != nil {
		logger.Error(errors.New("no .env file found"))
	}

	defer func() {
		// clean up resources
		CleanUp()
	}()

	// start scheduler
	schedule_manager.StartScheduleManager()

	// set up routing
	v1 := server.Group("/api/v1")
	// v1 := server.Group("/api/v1", middlewares.RateLimiter(60, 20, ""))

	{
		userV1 := v1.Group("/user")
		{
			userV1.GET("/profile", middlewares.AuthenticationMiddleware(false, false), userControllers.FetchProfile)

			userV1.GET("/fetch-user/:id", middlewares.AuthenticationMiddleware(false, false), userControllers.FetchUsersProfile)

			userV1.PUT("/update", middlewares.AuthenticationMiddleware(false, false), userControllers.UpdateUserData)

			userV1.PUT("/update/profile-image", middlewares.AuthenticationMiddleware(false, false), userControllers.UpdateProfileImage)

			userV1.GET("/websocket/channels", middlewares.AuthenticationMiddleware(false, false), auth.FetchUserWebsocketChannels)

			// can be called only 3 times per day
			userV1.POST("/update/phone", middlewares.AuthenticationMiddleware(false, false), userControllers.UpdatePhone)
			// userV1.POST("/update/phone", middlewares.RateLimiter(86400, 3, "phone-limiter"), middlewares.AuthenticationMiddleware(false, false), userControllers.UpdatePhone)

			userV1.POST("/fetch-users/phone", middlewares.AuthenticationMiddleware(false, false), userControllers.FetchUsersByPhoneNumber)
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

			workspaceV1.PUT("/members/add-priviledges", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.AddAdminAccess)

			workspaceV1.PUT("/members/deactivate/:id", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.DeactivateWorkspaceMember)

			workspaceV1.PUT("/update/workspace-image", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.UpdatWorkspaceProfileImage)

			workspaceV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.FetchUserWorkspaces)

			workspaceV1.GET("/members/fetch", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.FetchWorkspacesMembers)

			workspaceV1.GET("/search/members", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.SearchWorkspaceMembers)

			workspaceV1.POST("/invite", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.InviteToWorkspace)

			workspaceV1.PUT("/invite/:inviteId/reject", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.RejectWorkspaceInvite)

			workspaceV1.PUT("/invite/:inviteId/accept", middlewares.AuthenticationMiddleware(false, false), workspaceControllers.AcceptWorkspaceInvite)

			workspaceV1.GET("/invite/fetch", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.FetchWorkspaceInvites)
		}

		channelV1 := v1.Group("/channel")
		{
			channelV1.POST("/create", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.CreateChannel)

			channelV1.PUT("/update/channel-image", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.UpdateChannelProfileImage)

			channelV1.GET("/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchChannels)

			channelV1.GET("/fetch/all", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchAllChannels)

			channelV1.POST("/join/:id", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.CreateChannelMember)

			channelV1.DELETE("/delete/:id", middlewares.AuthenticationMiddleware(true, true), workspaceControllers.DeleteChannel)

			channelV1.POST("/add", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.AdminAddUserToChannel)

			channelV1.GET("/members/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchChannelMembers)

			channelV1.DELETE("/leave", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.LeaveChannel)

			channelV1.PUT("/pin", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.PinChannel)

			channelV1.PUT("/unpin", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.UnPinChannel)

			channelV1.GET("/pinned/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchPinnedChannels)

			channelV1.GET("/resources/fetch", middlewares.AuthenticationMiddleware(true, false), workspaceControllers.FetchChannelMedia)

		}

		messageV1 := v1.Group("/message")
		{
			messageV1.POST("/send", middlewares.AuthenticationMiddleware(false, false), conversationControllers.SendMessage)

			messageV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), conversationControllers.FetchMessages)

			messageV1.DELETE("/channel/delete", middlewares.AuthenticationMiddleware(true, false), conversationControllers.DeleteMessages)
		}

		conversationV1 := v1.Group("/conversation")
		{
			conversationV1.POST("/start", middlewares.AuthenticationMiddleware(false, false), conversationControllers.StartConversation)

			conversationV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), conversationControllers.FetchConversation)
		}

		teamV1 := v1.Group("/team")
		{
			teamV1.POST("/create", middlewares.AuthenticationMiddleware(true, true), teamControllers.CreateTeam)

			teamV1.GET("/fetch", middlewares.AuthenticationMiddleware(true, false), teamControllers.FetchTeams)

			teamV1.POST("/members/add/:id", middlewares.AuthenticationMiddleware(true, true), teamControllers.CreateTeamMembers)

			teamV1.DELETE("/members/remove", middlewares.AuthenticationMiddleware(true, true), teamControllers.RemoveTeamMembers)

			teamV1.GET("/members/fetch/:id", middlewares.AuthenticationMiddleware(true, false), teamControllers.FetchTeamMembers)
		}

		schduleV1 := v1.Group("/schedule")
		{
			schduleV1.POST("/create", middlewares.AuthenticationMiddleware(true, false), scheduleControllers.CreateSchedule)

			schduleV1.GET("/fetch", middlewares.AuthenticationMiddleware(false, false), scheduleControllers.FetchUserSchedule)
		}

		authV1 := v1.Group("/auth")
		{
			authV1.POST("/create", auth.CreateUser)

			authV1.POST("/verify", auth.VerifyAccountUseCase)

			authV1.POST("/login", auth.LoginUser)

			authV1.PUT("/update-password", middlewares.AuthenticationMiddleware(false, false), auth.UpdateLoggedInUsersPassword)

			authV1.PUT("/reset-password", auth.ResetUserPassword)

			authV1.GET("/generate-access-token", auth.GenerateAccessTokenFromRefresh)

			// can be called only 5 times per day
			authV1.GET("/resend-otp", auth.ResendOtp)
			// authV1.GET("/resend-otp", middlewares.RateLimiter(86400, 5, "email-otp-limiter"), auth.ResendOtp)

			authV1.PUT("/phone/verify", middlewares.AuthenticationMiddleware(false, false), auth.VerifyPhone)

			authV1.POST("/signout", middlewares.AuthenticationMiddleware(false, false), auth.SignOut)

			// centrifugo
			authV1.GET("/realtime/authenticate", middlewares.AuthenticationMiddleware(false, false), auth.GenerateCentrifugoToken)

			// acs
			authV1.GET("/realtime/calls/token", middlewares.AuthenticationMiddleware(false, false), auth.GenerateAcsToken)

			// oauth
			authV1.GET("/google/web/login", auth.GoogleLogin)

			authV1.GET("/google/callback", auth.GoogleCallBack)
		}
	}

	server.GET("/who-is-the-goat", func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusOK, "lionel Messi is the GOAT!", true, nil)
	})

	server.NoRoute(func(ctx *gin.Context) {
		server_response.Response(ctx, http.StatusNotFound, "this route does not exist", false, nil)
	})

	gin_mode := os.Getenv("GIN_MODE")
	port := os.Getenv("PORT")
	if gin_mode == "debug" {
		server.Run(port)
	} else if gin_mode == "release" {
		server.Run(":" + port)
	} else {
		panic("invalid gin mode used")
	}
	logger.Info("server is up on port" + port)
}
