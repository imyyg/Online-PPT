package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"online-ppt/internal/auth"
	"online-ppt/internal/config"
	internalhttp "online-ppt/internal/http"
	"online-ppt/internal/http/handlers"
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

	authService, err := auth.NewService(authRepo, tokenManager, auditLogger)
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
