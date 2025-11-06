package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"online-ppt/internal/auth"
	"online-ppt/internal/config"
	internalhttp "online-ppt/internal/http"
	"online-ppt/internal/http/handlers"
	"online-ppt/internal/records"
	"online-ppt/internal/storage"
)

func TestCreatePptRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempRoot := t.TempDir()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	recordsRepo, err := records.NewRepository(db)
	require.NoError(t, err)

	logger := log.New(io.Discard, "", 0)
	auditLogger := storage.NewAuditLogger(logger)

	recordsService, err := records.NewService(recordsRepo, tempRoot, auditLogger)
	require.NoError(t, err)

	tokenManager, err := auth.NewTokenManager("test-secret", time.Minute*5, time.Hour*24)
	require.NoError(t, err)

	recordsHandler := handlers.NewRecordsHandler(recordsService, tokenManager)

	cfg := &config.Config{
		Server: config.ServerConfig{Addr: ":8080"},
		Security: config.SecurityConfig{
			JWTSecret:       "test-secret",
			AccessTokenTTL:  time.Minute * 5,
			RefreshTokenTTL: time.Hour * 24,
		},
		Paths: config.PathConfig{PresentationsRoot: tempRoot},
	}

	router := internalhttp.NewRouter(cfg)
	internalhttp.RegisterRecordRoutes(router, recordsHandler)

	userID := int64(1)
	userUUID := "123e4567-e89b-12d3-a456-426614174000"
	groupName := "demodeck"
	description := "Team deck"
	relativePath := filepath.ToSlash(filepath.Join("presentations", userUUID, groupName, "slides"))
	canonicalRoot, err := filepath.Abs(tempRoot)
	require.NoError(t, err)
	canonicalPath := filepath.Join(canonicalRoot, userUUID, groupName, "slides")

	mock.ExpectExec("INSERT INTO ppt_records").
		WithArgs(userID, "DemoDeck", sqlmock.AnyArg(), groupName, relativePath, canonicalPath, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(42, 1))

	now := time.Now().UTC()
	mock.ExpectQuery("SELECT id, user_id, name, description, group_name, relative_path, canonical_path, tags, created_at, updated_at FROM ppt_records WHERE user_id = \\? AND id = \\? LIMIT 1").
		WithArgs(userID, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "description", "group_name", "relative_path", "canonical_path", "tags", "created_at", "updated_at"}).
			AddRow(int64(42), userID, "DemoDeck", description, groupName, relativePath, canonicalPath, `["tag1","tag2"]`, now, now))

	token, _, err := tokenManager.IssueAccessToken(userID, userUUID)
	require.NoError(t, err)
	payload := map[string]any{
		"name":        "DemoDeck",
		"description": description,
		"tags":        []string{"Tag1", "tag2", "tag1"},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ppts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		ID            int64     `json:"id"`
		Name          string    `json:"name"`
		GroupName     string    `json:"groupName"`
		Description   *string   `json:"description"`
		RelativePath  string    `json:"relativePath"`
		CanonicalPath string    `json:"canonicalPath"`
		Tags          []string  `json:"tags"`
		PathStatus    string    `json:"pathStatus"`
		CreatedAt     time.Time `json:"createdAt"`
		UpdatedAt     time.Time `json:"updatedAt"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, int64(42), resp.ID)
	require.Equal(t, "DemoDeck", resp.Name)
	require.Equal(t, groupName, resp.GroupName)
	require.NotNil(t, resp.Description)
	require.Equal(t, description, *resp.Description)
	require.Equal(t, relativePath, resp.RelativePath)
	require.Equal(t, canonicalPath, resp.CanonicalPath)
	require.ElementsMatch(t, []string{"tag1", "tag2"}, resp.Tags)
	require.Equal(t, "valid", resp.PathStatus)

	_, statErr := os.Stat(canonicalPath)
	require.NoError(t, statErr)

	require.NoError(t, mock.ExpectationsWereMet())
}
