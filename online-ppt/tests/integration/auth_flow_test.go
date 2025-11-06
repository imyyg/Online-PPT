package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"online-ppt/internal/auth"
	"online-ppt/internal/config"
	internalhttp "online-ppt/internal/http"
	"online-ppt/internal/http/handlers"
	"online-ppt/internal/storage"
)

func TestAuthRegisterAndLoginFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	repo, err := auth.NewRepository(db)
	require.NoError(t, err)

	tokenManager, err := auth.NewTokenManager("test-secret", time.Minute*5, time.Hour*24*30)
	require.NoError(t, err)

	auditLogger := storage.NewAuditLogger(log.New(io.Discard, "", 0))

	authService, err := auth.NewService(repo, tokenManager, auditLogger)
	require.NoError(t, err)

	cfg := &config.Config{
		Server: config.ServerConfig{Addr: ":8080"},
		Security: config.SecurityConfig{
			JWTSecret:       "test-secret",
			AccessTokenTTL:  time.Minute * 5,
			RefreshTokenTTL: time.Hour * 24 * 30,
		},
	}

	authHandler := handlers.NewAuthHandler(authService, cfg)
	router := internalhttp.NewRouter(cfg)
	internalhttp.RegisterAuthRoutes(router, authHandler)

	email := "user@example.com"
	password := "PptDemo123!"
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	now := time.Now().UTC()
	uuidValue := "123e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec("INSERT INTO user_accounts").
		WithArgs(email, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id, uuid, email, password_hash, status, last_login_at, created_at, updated_at FROM user_accounts WHERE id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "status", "last_login_at", "created_at", "updated_at"}).
			AddRow(int64(1), uuidValue, email, hashedPassword, "active", sql.NullTime{}, now, now))

	mock.ExpectQuery("SELECT id, uuid, email, password_hash, status, last_login_at, created_at, updated_at FROM user_accounts WHERE email = \\?").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "status", "last_login_at", "created_at", "updated_at"}).
			AddRow(int64(1), uuidValue, email, hashedPassword, "active", sql.NullTime{}, now, now))

	mock.ExpectExec("INSERT INTO user_sessions").
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id, user_id, refresh_token_hash, expires_at, issued_at, client_fingerprint, revoked_at, created_at FROM user_sessions WHERE id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "refresh_token_hash", "expires_at", "issued_at", "client_fingerprint", "revoked_at", "created_at"}).
			AddRow(int64(1), int64(1), "hash", now.Add(24*time.Hour), now, sql.NullString{}, sql.NullTime{}, now))

	mock.ExpectExec("UPDATE user_accounts SET last_login_at = \\?").
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, email, resp.User.Email)

	cookies := rec.Result().Cookies()
	foundCookie := false
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.Value != "" {
			foundCookie = true
		}
	}
	require.True(t, foundCookie, "refresh token cookie not set")

	require.NoError(t, mock.ExpectationsWereMet())
}
