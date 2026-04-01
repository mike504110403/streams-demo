package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/streams-demo/backend/config"
	"github.com/streams-demo/backend/internal/handler"
	"github.com/streams-demo/backend/internal/middleware"
	"github.com/streams-demo/backend/internal/repository"
	"github.com/streams-demo/backend/internal/service"
)

func main() {
	// 載入設定
	cfg := config.Load()

	ctx := context.Background()

	// 連線 PostgreSQL
	dbPool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("無法連線 PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	// 測試 DB 連線
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("PostgreSQL ping 失敗: %v", err)
	}
	log.Println("PostgreSQL 連線成功")

	// 連線 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer rdb.Close()

	// 測試 Redis 連線
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis ping 失敗: %v", err)
	}
	log.Println("Redis 連線成功")

	// 初始化各層
	userRepo := repository.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, rdb, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)
	streamHandler := handler.NewStreamHandler()
	srsHandler := handler.NewSRSHandler()

	// 設定 Gin
	r := gin.Default()

	// CORS（MVP 開發用，允許所有來源）
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 路由
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(authService))
		{
			users.GET("/me", userHandler.GetMe)
			users.PUT("/me", userHandler.UpdateMe)
		}

		// 直播間 API（需認證）
		streams := v1.Group("/streams")
		streams.Use(middleware.AuthMiddleware(authService))
		{
			streams.POST("", streamHandler.CreateStream)
			streams.GET("", streamHandler.ListStreams)
			streams.GET("/:id", streamHandler.GetStream)
			streams.PUT("/:id", streamHandler.UpdateStream)
			streams.DELETE("/:id", streamHandler.DeleteStream)
		}

		// SRS Callback（內部使用，不走 JWT）
		srs := v1.Group("/internal/srs")
		{
			srs.POST("/on_publish", srsHandler.OnPublish)
			srs.POST("/on_unpublish", srsHandler.OnUnpublish)
			srs.POST("/on_play", srsHandler.OnPlay)
			srs.POST("/on_stop", srsHandler.OnStop)
		}
	}

	// 啟動 HTTP server（graceful shutdown）
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("HTTP server 啟動於 :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server 錯誤: %v", err)
		}
	}()

	// 等待中斷信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在關閉 server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server 強制關閉: %v", err)
	}

	log.Println("Server 已正常關閉")
}
