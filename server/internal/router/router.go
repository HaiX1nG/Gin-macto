package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/handler"
	"github.com/yourorg/livemix/internal/middleware"
	"github.com/yourorg/livemix/internal/ws"
)

// SetupRouter 设置路由
func SetupRouter(
	authHandler *handler.AuthHandler,
	serverHandler *handler.ServerHandler,
	channelHandler *handler.ChannelHandler,
	messageHandler *handler.MessageHandler,
	roleHandler *handler.RoleHandler,
	voiceHandler *handler.VoiceHandler,
	playlistHandler *handler.PlaylistHandler,
	friendHandler *handler.FriendHandler,
	uploadHandler *handler.UploadHandler,
	wsHandler *ws.Handler,
) *gin.Engine {
	r := gin.New()

	// 中间件注册顺序：日志追踪 -> 限流 -> CORS -> 认证 -> 授权 -> 参数校验 -> handler
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(100))

	// API v1
	v1 := r.Group("/api/v1")

	// 认证相关（无需鉴权）
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
	}

	// 用户相关（需要鉴权）
	userGroup := v1.Group("/user")
	userGroup.Use(middleware.JWTAuth())
	{
		userGroup.GET("/info", authHandler.GetUserInfo)
		userGroup.GET("/profile", authHandler.GetUserInfo)
		userGroup.PUT("/profile", authHandler.UpdateProfile)
		userGroup.PUT("/password", authHandler.ChangePassword)
		userGroup.PUT("/status", authHandler.SetCustomStatus)
		userGroup.DELETE("/account", authHandler.DeleteAccount)
	}

	// 用户状态查询
	v1.GET("/users/:id/status", middleware.JWTAuth(), authHandler.GetUserOnlineStatus)
	v1.GET("/users/:id/online", middleware.JWTAuth(), authHandler.GetUserOnlineStatus)
	v1.GET("/users/:id/info", middleware.JWTAuth(), authHandler.GetUserInfoByID)

	// 服务器相关（需要鉴权）
	serverGroup := v1.Group("/servers")
	serverGroup.Use(middleware.JWTAuth())
	{
		serverGroup.POST("", serverHandler.CreateServer)
		serverGroup.GET("", serverHandler.GetServerList)
		serverGroup.GET("/:id", serverHandler.GetServerDetail)
		serverGroup.PUT("/:id", serverHandler.UpdateServer)
		serverGroup.DELETE("/:id", serverHandler.DeleteServer)
		serverGroup.POST("/:id/join", serverHandler.JoinServer)
		serverGroup.POST("/:id/leave", serverHandler.LeaveServer)

		// 成员管理
		serverGroup.GET("/:id/members", serverHandler.GetMembers)
		serverGroup.GET("/:id/members/:uid", serverHandler.GetMember)
		serverGroup.PUT("/:id/members/:uid", serverHandler.UpdateMember)
		serverGroup.DELETE("/:id/members/:uid", serverHandler.KickMember)

		// 角色管理
		serverGroup.GET("/:id/roles", roleHandler.GetRoles)
		serverGroup.POST("/:id/roles", roleHandler.CreateRole)
		serverGroup.PUT("/:id/roles/:rid", roleHandler.UpdateRole)
		serverGroup.DELETE("/:id/roles/:rid", roleHandler.DeleteRole)

		// 频道管理
		serverGroup.POST("/:sid/channels", channelHandler.CreateChannel)
		serverGroup.GET("/:sid/channels", channelHandler.GetChannelTree)
		serverGroup.PUT("/:sid/channels/reorder", channelHandler.ReorderChannels)
		serverGroup.PUT("/:sid/channels/:cid", channelHandler.UpdateChannel)
		serverGroup.DELETE("/:sid/channels/:cid", channelHandler.DeleteChannel)
	}

	// 频道维度操作（消息/语音/屏幕共享/播放列表）
	channelGroup := v1.Group("/channels/:cid")
	channelGroup.Use(middleware.JWTAuth())
	{
		// 消息
		channelGroup.GET("/messages", messageHandler.GetMessages)
		channelGroup.POST("/messages", messageHandler.SendMessage)
		channelGroup.PUT("/messages/:mid", messageHandler.UpdateMessage)
		channelGroup.DELETE("/messages/:mid", messageHandler.DeleteMessage)
		channelGroup.POST("/messages/:mid/pin", messageHandler.PinMessage)
		channelGroup.POST("/messages/:mid/reactions", messageHandler.AddReaction)
		channelGroup.DELETE("/messages/:mid/reactions/:emoji", messageHandler.RemoveReaction)

		// 语音
		channelGroup.POST("/voice/join", voiceHandler.JoinVoice)
		channelGroup.POST("/voice/leave", voiceHandler.LeaveVoice)
		channelGroup.GET("/voice/participants", voiceHandler.GetVoiceParticipants)
		channelGroup.POST("/voice/mute", voiceHandler.SetMute)

		// 屏幕共享
		channelGroup.POST("/screenshare/start", voiceHandler.StartScreenShare)
		channelGroup.POST("/screenshare/stop", voiceHandler.StopScreenShare)
		channelGroup.GET("/screenshare", voiceHandler.GetActiveScreenShare)

		// 播放列表
		channelGroup.GET("/playlist", playlistHandler.GetPlaylist)
		channelGroup.POST("/playlist", playlistHandler.AddItem)
		channelGroup.DELETE("/playlist/:itemId", playlistHandler.RemoveItem)
		channelGroup.POST("/playlist/play", playlistHandler.Play)
		channelGroup.POST("/playlist/pause", playlistHandler.Pause)
		channelGroup.POST("/playlist/skip", playlistHandler.Skip)
		channelGroup.POST("/playlist/reorder", playlistHandler.Reorder)
	}

	// 消息搜索（全局，需要鉴权）
	v1.GET("/messages/search", middleware.JWTAuth(), messageHandler.SearchMessages)

	// 好友相关（需要鉴权）
	friendGroup := v1.Group("/friends")
	friendGroup.Use(middleware.JWTAuth())
	{
		friendGroup.POST("/request", friendHandler.SendFriendRequest)
		friendGroup.POST("/request/:id/handle", friendHandler.HandleFriendRequest)
		friendGroup.GET("/requests", friendHandler.GetPendingRequests)
		friendGroup.GET("", friendHandler.GetFriendList)
		friendGroup.DELETE("/:id", friendHandler.DeleteFriend)
		friendGroup.GET("/search", friendHandler.SearchUser)
		friendGroup.POST("/messages", friendHandler.SendPrivateMessage)
		friendGroup.GET("/messages/unread", friendHandler.GetUnreadCount)
		friendGroup.GET("/:id/messages", friendHandler.GetPrivateMessages)
		friendGroup.GET("/conversations", friendHandler.GetConversations)
	}

	// 文件上传（需要鉴权）
	v1.POST("/upload", middleware.JWTAuth(), uploadHandler.Upload)

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	return r
}
