package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/livemix/config"
	"github.com/yourorg/livemix/internal/handler"
	"github.com/yourorg/livemix/internal/middleware"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/internal/router"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/internal/ws"
	"github.com/yourorg/livemix/pkg/database"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := config.Load("config.yaml"); err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}
	cfg := config.Get()

	// 初始化日志系统
	logCfg := &logger.Config{
		Level:        logger.Level(cfg.Log.Level),
		Format:       cfg.Log.Format,
		OutputPaths:  cfg.Log.OutputPaths,
		EnableCaller: cfg.Log.EnableCaller,
	}
	if err := logger.Init(logCfg); err != nil {
		logger.Fatalf("初始化日志系统失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("配置加载成功",
		zap.Int("port", cfg.Server.Port),
		zap.String("log_level", cfg.Log.Level),
		zap.String("log_format", cfg.Log.Format),
	)

	// 初始化Origin验证器
	middleware.InitializeOriginValidator(cfg)
	logger.Info("Origin验证器初始化完成")

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		logger.Fatal("初始化数据库失败", logger.WithError(err))
	}
	defer database.Close()
	logger.Info("数据库连接成功")

	// 初始化Redis（可选）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Warn("初始化Redis失败（将继续运行）", logger.WithError(err))
	} else {
		defer database.CloseRedis()
		logger.Info("Redis连接成功")
	}

	// 初始化仓储
	db := database.MustGetDB()
	userRepo := repository.NewUserRepository(db)
	userStatusRepo := repository.NewUserStatusRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	serverRepo := repository.NewServerRepository(db)
	memberRepo := repository.NewServerMemberRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	voiceRepo := repository.NewVoiceRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)

	// 初始化事务管理器
	txManager := repository.NewTransactionManager(db)

	// 初始化服务
	userService := service.NewUserService(userRepo, userStatusRepo, txManager)
	friendService := service.NewFriendService(friendRepo, userRepo, userStatusRepo, txManager)
	uploadService := service.NewUploadService()
	serverService := service.NewServerService(serverRepo, memberRepo, roleRepo, channelRepo, userRepo, txManager)
	channelService := service.NewChannelService(channelRepo, serverRepo, serverService)
	roleService := service.NewRoleService(roleRepo, serverRepo, serverService)
	messageService := service.NewMessageService(msgRepo, reactionRepo, channelRepo, userRepo, serverService)
	voiceService := service.NewVoiceService(voiceRepo, channelRepo, userRepo, serverService)
	playlistService := service.NewPlaylistService(playlistRepo, channelRepo, serverService)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(userService)
	friendHandler := handler.NewFriendHandler(friendService)
	uploadHandler := handler.NewUploadHandler(uploadService)
	serverHandler := handler.NewServerHandler(serverService)
	channelHandler := handler.NewChannelHandler(channelService)
	roleHandler := handler.NewRoleHandler(roleService)
	messageHandler := handler.NewMessageHandler(messageService)
	voiceHandler := handler.NewVoiceHandler(voiceService)
	playlistHandler := handler.NewPlaylistHandler(playlistService)

	// 初始化WebSocket Hub
	hub := ws.NewHub(userService, cfg.WebSocket.HeartbeatTimeout)
	go hub.Run()
	wsHandler := ws.NewHandler(hub)
	logger.Info("WebSocket Hub初始化完成",
		zap.Duration("heartbeat_timeout", cfg.WebSocket.HeartbeatTimeout))

	// 设置路由
	r := router.SetupRouter(
		authHandler,
		serverHandler,
		channelHandler,
		messageHandler,
		roleHandler,
		voiceHandler,
		playlistHandler,
		friendHandler,
		uploadHandler,
		wsHandler,
	)

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		logger.Info("服务器启动", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", logger.WithError(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器关闭失败", logger.WithError(err))
	}

	logger.Info("服务器已关闭")
}
