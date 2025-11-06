package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"online-ppt/internal/config"
	"online-ppt/internal/http/handlers"
	"online-ppt/internal/http/middleware"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
	apiPrefix       = "/api/v1"
)

// NewRouter builds a Gin engine with baseline middleware.
func NewRouter(cfg *config.Config, extra ...gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestLogger(nil))
	engine.Use(gin.Recovery())
	for _, fn := range extra {
		if fn != nil {
			engine.Use(fn)
		}
	}
	appendHealthRoutes(engine)
	return engine
}

// RunServer serves the HTTP API using http.Server for graceful shutdown hooks.
func RunServer(ctx context.Context, cfg *config.Config, engine *gin.Engine) error {
	if cfg == nil {
		return errNilConfig
	}
	if engine == nil {
		return errNilEngine
	}
	if ctx == nil {
		ctx = context.Background()
	}

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-errCh; err != nil {
		return err
	}

	return ctx.Err()
}

func appendHealthRoutes(engine *gin.Engine) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

var (
	errNilConfig = errors.New("http: config must not be nil")
	errNilEngine = errors.New("http: engine must not be nil")
)

// RegisterAuthRoutes wires authentication HTTP handlers under the API prefix.
func RegisterAuthRoutes(engine *gin.Engine, handler *handlers.AuthHandler) {
	if engine == nil || handler == nil {
		return
	}
	authGroup := engine.Group(apiPrefix + "/auth")
	authGroup.POST("/register", handler.Register)
	authGroup.POST("/login", handler.Login)
	authGroup.POST("/refresh", handler.Refresh)
	authGroup.POST("/logout", handler.Logout)
}

// RegisterRecordRoutes wires PPT record HTTP handlers under the API prefix.
func RegisterRecordRoutes(engine *gin.Engine, handler *handlers.RecordsHandler) {
	if engine == nil || handler == nil {
		return
	}
	recordGroup := engine.Group(apiPrefix + "/ppts")
	recordGroup.GET("", handler.List)
	recordGroup.POST("", handler.Create)
	recordGroup.GET("/:id", handler.Get)
	recordGroup.PATCH("/:id", handler.Update)
	recordGroup.DELETE("/:id", handler.Delete)
}
