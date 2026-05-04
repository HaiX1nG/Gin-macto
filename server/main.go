package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/livemix/config"
	"github.com/yourorg/livemix/internal/handler"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/internal/router"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/internal/ws"
	"github.com/yourorg/livemix/pkg/database"
)

func main() {
	// 加载配置
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := config.Get()

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 初始化Redis（可选）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Printf("初始化Redis失败（将继续运行）: %v", err)
	}
	defer database.CloseRedis()

	// 初始化仓储
	db := database.GetDB()
	userRepo := repository.NewUserRepository(db)
	userStatusRepo := repository.NewUserStatusRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	participantRepo := repository.NewRoomParticipantRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)
	chatMsgRepo := repository.NewChatMessageRepository(db)
	screenShareRepo := repository.NewScreenShareRepository(db)
	voiceSessionRepo := repository.NewVoiceSessionRepository(db)

	// 初始化服务
	userService := service.NewUserService(userRepo, userStatusRepo)
	roomService := service.NewRoomService(roomRepo, participantRepo, userRepo)
	playlistService := service.NewPlaylistService(playlistRepo, roomRepo, participantRepo)
	chatService := service.NewChatService(chatMsgRepo, userRepo, participantRepo)
	screenShareService := service.NewScreenShareService(screenShareRepo, roomRepo, userRepo, participantRepo)
	voiceService := service.NewVoiceService(voiceSessionRepo, roomRepo, userRepo, participantRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(userService)
	roomHandler := handler.NewRoomHandler(roomService)
	playlistHandler := handler.NewPlaylistHandler(playlistService)
	chatHandler := handler.NewChatHandler(chatService)
	screenShareHandler := handler.NewScreenShareHandler(screenShareService)
	voiceHandler := handler.NewVoiceHandler(voiceService)

	// 初始化WebSocket Hub
	hub := ws.NewHub(userService)
	go hub.Run()
	wsHandler := ws.NewHandler(hub)

	// 设置路由
	r := router.SetupRouter(authHandler, roomHandler, playlistHandler, chatHandler, screenShareHandler, voiceHandler, wsHandler)

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		log.Printf("服务器启动，监听端口 %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}
