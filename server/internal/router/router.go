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
	wsHandler *ws.Handler,
) *gin.Engine {
	r := gin.New()

	// 中间件注册顺序：日志追踪 -> 限流 -> CORS -> 认证 -> 授权 -> 参数校验 -> handler
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
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
	}

	// 用户相关（需要鉴权）
	userGroup := v1.Group("/user")
	userGroup.Use(middleware.JWTAuth())
	{
		userGroup.GET("/info", authHandler.GetUserInfo)
		userGroup.PUT("/profile", authHandler.UpdateProfile)
		userGroup.PUT("/password", authHandler.ChangePassword)
		userGroup.GET("/messages", chatHandler.GetUserHistoryMessages)
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
	}

	// 用户状态查询（需要鉴权）
	v1.GET("/users/:id/status", middleware.JWTAuth(), roomHandler.GetUserStatus)

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

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	return r
}
