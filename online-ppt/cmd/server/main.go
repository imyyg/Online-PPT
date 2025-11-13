package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"online-ppt/internal/auth"
	"online-ppt/internal/cache"
	"online-ppt/internal/captcha"
	"online-ppt/internal/config"
	internalhttp "online-ppt/internal/http"
	"online-ppt/internal/http/handlers"
	"online-ppt/internal/mail"
	"online-ppt/internal/records"
	"online-ppt/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := storage.MySQLFactory{DSN: cfg.Storage.DSN}.Open(ctx)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer db.Close()

	migrator := storage.Migrator{}
	if err := migrator.Apply(ctx, db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	authRepo, err := auth.NewRepository(db)
	if err != nil {
		log.Fatalf("init auth repository: %v", err)
	}

	tokenManager, err := auth.NewTokenManager(cfg.Security.JWTSecret, cfg.Security.AccessTokenTTL, cfg.Security.RefreshTokenTTL)
	if err != nil {
		log.Fatalf("init token manager: %v", err)
	}

	auditLogger := storage.NewAuditLogger(nil)

	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer redisClient.Close()

	// 测试 Redis 连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	// 初始化 Cache Service
	cacheService := cache.NewRedisService(redisClient)

	// 初始化 Captcha Service
	captchaService := captcha.NewService(cacheService)

	// 初始化 Mail Service
	mailService := mail.NewSMTPService(mail.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		FromName: cfg.SMTP.FromName,
	})

	authService, err := auth.NewService(authRepo, tokenManager, auditLogger, cacheService, captchaService, mailService)
	if err != nil {
		log.Fatalf("init auth service: %v", err)
	}

	authHandler := handlers.NewAuthHandler(authService, cfg)

	recordsRepo, err := records.NewRepository(db)
	if err != nil {
		log.Fatalf("init records repository: %v", err)
	}

	recordsService, err := records.NewService(recordsRepo, cfg.Paths.PresentationsRoot, auditLogger)
	if err != nil {
		log.Fatalf("init records service: %v", err)
	}

	recordsHandler := handlers.NewRecordsHandler(recordsService, tokenManager)
	router := internalhttp.NewRouter(cfg)
	internalhttp.RegisterAuthRoutes(router, authHandler)
	internalhttp.RegisterRecordRoutes(router, recordsHandler)

	if err := internalhttp.RunServer(ctx, cfg, router); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Fatalf("server exited with error: %v", err)
	}
}
