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
	roomHandler *handler.RoomHandler,
	playlistHandler *handler.PlaylistHandler,
	chatHandler *handler.ChatHandler,
	screenShareHandler *handler.ScreenShareHandler,
	voiceHandler *handler.VoiceHandler,
	friendHandler *handler.FriendHandler,
	wsHandler *ws.Handler,
) *gin.Engine {
	r := gin.New()

	// 中间件注册顺序：日志追踪 -> 限流 -> CORS -> 认证 -> 授权 -> 参数校验 -> handler
	// RequestID 中间件必须放在最前面，确保所有后续中间件和处理器都能获取到 TraceID
	r.Use(middleware.RequestID())
	// 使用自定义的 zap 日志中间件替换 gin 默认日志
	r.Use(middleware.Logger())
	// 使用自定义的恢复中间件，记录完整的 panic 堆栈信息
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
		userGroup.GET("/profile", authHandler.GetUserInfo) // 别名路由，与 /info 返回相同数据
		userGroup.PUT("/profile", authHandler.UpdateProfile)
		userGroup.PUT("/password", authHandler.ChangePassword)
		userGroup.PUT("/status", authHandler.SetCustomStatus)
		userGroup.GET("/messages", chatHandler.GetUserHistoryMessages)
		userGroup.DELETE("/account", authHandler.DeleteAccount)
	}

	// 房间相关
	roomGroup := v1.Group("/rooms")
	{
		// 公开接口
		roomGroup.GET("", roomHandler.GetRoomList)
		roomGroup.GET("/:id", roomHandler.GetRoomInfo)
		roomGroup.GET("/:id/participants", roomHandler.GetRoomParticipants)
		roomGroup.GET("/:id/online/count", roomHandler.GetOnlineCount)
		roomGroup.GET("/:id/online/users", roomHandler.GetOnlineUsers)

		// 需要鉴权
		roomGroup.POST("", middleware.JWTAuth(), roomHandler.CreateRoom)
		roomGroup.POST("/join/:id", middleware.JWTAuth(), roomHandler.JoinRoom)
		roomGroup.POST("/leave/:id", middleware.JWTAuth(), roomHandler.LeaveRoom)
		roomGroup.DELETE("/:id", middleware.JWTAuth(), roomHandler.DeleteRoom)
	}

	// 用户状态查询（需要鉴权）
	v1.GET("/users/:id/status", middleware.JWTAuth(), roomHandler.GetUserStatus)
	v1.GET("/users/:id/online", middleware.JWTAuth(), authHandler.GetUserOnlineStatus)

	// 播放列表相关（需要鉴权）
	playlistGroup := v1.Group("/rooms/:id/playlist")
	playlistGroup.Use(middleware.JWTAuth())
	{
		playlistGroup.GET("", playlistHandler.GetPlaylist)
		playlistGroup.POST("", playlistHandler.AddItem)
		playlistGroup.DELETE("/:itemId", playlistHandler.RemoveItem)
		playlistGroup.POST("/play", playlistHandler.Play)
		playlistGroup.POST("/pause", playlistHandler.Pause)
		playlistGroup.POST("/skip", playlistHandler.Skip)
	}

	// 聊天相关（需要鉴权）
	chatGroup := v1.Group("/rooms/:id/messages")
	chatGroup.Use(middleware.JWTAuth())
	{
		chatGroup.GET("", chatHandler.GetMessages)
		chatGroup.POST("", chatHandler.SendMessage)
	}

	// 消息搜索（需要鉴权）
	v1.GET("/messages/search", middleware.JWTAuth(), chatHandler.SearchMessages)

	// 屏幕共享相关（需要鉴权）
	screenShareGroup := v1.Group("/rooms/:id/screenshare")
	screenShareGroup.Use(middleware.JWTAuth())
	{
		screenShareGroup.POST("/start", screenShareHandler.StartScreenShare)
		screenShareGroup.POST("/stop", screenShareHandler.StopScreenShare)
		screenShareGroup.GET("", screenShareHandler.GetActiveScreenShare)
	}

	// 语音相关（需要鉴权）
	voiceGroup := v1.Group("/rooms/:id/voice")
	voiceGroup.Use(middleware.JWTAuth())
	{
		voiceGroup.POST("/join", voiceHandler.JoinVoice)
		voiceGroup.POST("/leave", voiceHandler.LeaveVoice)
		voiceGroup.GET("/participants", voiceHandler.GetVoiceParticipants)
		voiceGroup.POST("/mute", voiceHandler.SetMute)
	}

	// 好友相关（需要鉴权）
	friendGroup := v1.Group("/friends")
	friendGroup.Use(middleware.JWTAuth())
	{
		// 好友请求
		friendGroup.POST("/request", friendHandler.SendFriendRequest)
		friendGroup.POST("/request/:id/handle", friendHandler.HandleFriendRequest)
		friendGroup.GET("/requests", friendHandler.GetPendingRequests)

		// 好友列表
		friendGroup.GET("", friendHandler.GetFriendList)
		friendGroup.DELETE("/:id", friendHandler.DeleteFriend)

		// 搜索用户
		friendGroup.GET("/search", friendHandler.SearchUser)

		// 私聊消息
		friendGroup.POST("/messages", friendHandler.SendPrivateMessage)
		friendGroup.GET("/messages/unread", friendHandler.GetUnreadCount)
		friendGroup.GET("/:id/messages", friendHandler.GetPrivateMessages)
		friendGroup.GET("/conversations", friendHandler.GetConversations)
	}

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	return r
}
