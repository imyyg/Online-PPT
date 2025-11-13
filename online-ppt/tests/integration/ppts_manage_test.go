package integration

import (
	"bytes"
	"database/sql"
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

type recordsTestContext struct {
	router   *gin.Engine
	mock     sqlmock.Sqlmock
	token    string
	userID   int64
	userUUID string
	root     string
}

const (
	selectRecordQuery = "SELECT id, user_id, name, title, description, group_name, relative_path, canonical_path, tags, created_at, updated_at FROM ppt_records WHERE user_id = \\? AND id = \\? LIMIT 1"
	baseDescription   = "Primary deck"
)

func newRecordsTestContext(t *testing.T) *recordsTestContext {
	gin.SetMode(gin.TestMode)

	tempRoot := t.TempDir()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repo, err := records.NewRepository(db)
	require.NoError(t, err)

	logger := log.New(io.Discard, "", 0)
	auditLogger := storage.NewAuditLogger(logger)

	service, err := records.NewService(repo, tempRoot, auditLogger)
	require.NoError(t, err)

	tokenManager, err := auth.NewTokenManager("test-secret", time.Minute*5, time.Hour*24)
	require.NoError(t, err)

	handler := handlers.NewRecordsHandler(service, tokenManager)

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
	internalhttp.RegisterRecordRoutes(router, handler)

	userID := int64(1)
	userUUID := "123e4567-e89b-12d3-a456-426614174000"
	token, _, err := tokenManager.IssueAccessToken(userID, userUUID)
	require.NoError(t, err)

	return &recordsTestContext{
		router:   router,
		mock:     mock,
		token:    token,
		userID:   userID,
		userUUID: userUUID,
		root:     tempRoot,
	}
}

func (ctx *recordsTestContext) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+ctx.token)
}

func TestListPptRecords(t *testing.T) {
	ctx := newRecordsTestContext(t)

	rel := filepath.ToSlash(filepath.Join("presentations", ctx.userUUID, "deckone", "slides"))
	canonicalValid := filepath.Join(ctx.root, ctx.userUUID, "deckone", "slides")
	require.NoError(t, os.MkdirAll(canonicalValid, 0o755))

	relMissing := filepath.ToSlash(filepath.Join("presentations", ctx.userUUID, "decktwo", "slides"))
	canonicalMissing := filepath.Join(ctx.root, ctx.userUUID, "decktwo", "slides")

	now := time.Now().UTC()

	like := "%demo%"
	ctx.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM ppt_records").
		WithArgs(ctx.userID, like, like, like, "tag1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	ctx.mock.ExpectQuery("SELECT id, user_id, name, title, description, group_name, relative_path, canonical_path, tags, created_at, updated_at FROM ppt_records").
		WithArgs(ctx.userID, like, like, like, "tag1", 10, 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "title", "description", "group_name", "relative_path", "canonical_path", "tags", "created_at", "updated_at"}).
			AddRow(int64(10), ctx.userID, "DeckOne", nil, baseDescription, "deckone", rel, canonicalValid, "[\"tag1\",\"tag2\"]", now, now).
			AddRow(int64(11), ctx.userID, "DeckTwo", nil, nil, "decktwo", relMissing, canonicalMissing, nil, now, now))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ppts?q=demo&tag=tag1&sort=name_asc&limit=10&offset=5", nil)
	ctx.authorize(req)

	rec := httptest.NewRecorder()
	ctx.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Items  []struct {
			ID           int64     `json:"id"`
			Name         string    `json:"name"`
			Description  *string   `json:"description"`
			RelativePath string    `json:"relativePath"`
			Canonical    string    `json:"canonicalPath"`
			Tags         []string  `json:"tags"`
			PathStatus   string    `json:"pathStatus"`
			CreatedAt    time.Time `json:"createdAt"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, 2, resp.Total)
	require.Equal(t, 10, resp.Limit)
	require.Equal(t, 5, resp.Offset)
	require.Len(t, resp.Items, 2)

	require.Equal(t, int64(10), resp.Items[0].ID)
	require.Equal(t, "DeckOne", resp.Items[0].Name)
	require.NotNil(t, resp.Items[0].Description)
	require.Equal(t, baseDescription, *resp.Items[0].Description)
	require.Equal(t, rel, resp.Items[0].RelativePath)
	require.Equal(t, canonicalValid, resp.Items[0].Canonical)
	require.ElementsMatch(t, []string{"tag1", "tag2"}, resp.Items[0].Tags)
	require.Equal(t, "valid", resp.Items[0].PathStatus)

	require.Equal(t, int64(11), resp.Items[1].ID)
	require.Nil(t, resp.Items[1].Description)
	require.Equal(t, 0, len(resp.Items[1].Tags))
	require.Equal(t, "missing", resp.Items[1].PathStatus)

	require.NoError(t, ctx.mock.ExpectationsWereMet())
}

func TestGetPptRecord(t *testing.T) {
	ctx := newRecordsTestContext(t)

	rel := filepath.ToSlash(filepath.Join("presentations", ctx.userUUID, "deckone", "slides"))
	canonical := filepath.Join(ctx.root, ctx.userUUID, "deckone", "slides")
	require.NoError(t, os.MkdirAll(canonical, 0o755))

	now := time.Now().UTC()

	ctx.mock.ExpectQuery(selectRecordQuery).
		WithArgs(ctx.userID, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "title", "description", "group_name", "relative_path", "canonical_path", "tags", "created_at", "updated_at"}).
			AddRow(int64(42), ctx.userID, "DeckOne", nil, baseDescription, "deckone", rel, canonical, "[\"tag1\"]", now, now))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ppts/42", nil)
	ctx.authorize(req)

	rec := httptest.NewRecorder()
	ctx.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		ID           int64    `json:"id"`
		Name         string   `json:"name"`
		GroupName    string   `json:"groupName"`
		Description  *string  `json:"description"`
		RelativePath string   `json:"relativePath"`
		Canonical    string   `json:"canonicalPath"`
		Tags         []string `json:"tags"`
		PathStatus   string   `json:"pathStatus"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, int64(42), resp.ID)
	require.Equal(t, "valid", resp.PathStatus)
	require.NotNil(t, resp.Description)

	require.NoError(t, ctx.mock.ExpectationsWereMet())
}

func TestGetPptRecordNotFound(t *testing.T) {
	ctx := newRecordsTestContext(t)

	ctx.mock.ExpectQuery(selectRecordQuery).
		WithArgs(ctx.userID, int64(99)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ppts/99", nil)
	ctx.authorize(req)

	rec := httptest.NewRecorder()
	ctx.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.NoError(t, ctx.mock.ExpectationsWereMet())
}

func TestUpdatePptRecord(t *testing.T) {
	ctx := newRecordsTestContext(t)

	existingRel := filepath.ToSlash(filepath.Join("presentations", ctx.userUUID, "deckone", "slides"))
	existingCanonical := filepath.Join(ctx.root, ctx.userUUID, "deckone", "slides")
	require.NoError(t, os.MkdirAll(existingCanonical, 0o755))

	now := time.Now().UTC()

	ctx.mock.ExpectQuery(selectRecordQuery).
		WithArgs(ctx.userID, int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "title", "description", "group_name", "relative_path", "canonical_path", "tags", "created_at", "updated_at"}).
			AddRow(int64(7), ctx.userID, "DeckOne", nil, baseDescription, "deckone", existingRel, existingCanonical, "[\"tag1\"]", now, now))

	newRel := filepath.ToSlash(filepath.Join("presentations", ctx.userUUID, "decktwo", "slides"))
	newCanonical := filepath.Join(ctx.root, ctx.userUUID, "decktwo", "slides")

	ctx.mock.ExpectExec("UPDATE ppt_records SET").
		WithArgs("DeckTwo", sqlmock.AnyArg(), sqlmock.AnyArg(), "decktwo", newRel, newCanonical, sqlmock.AnyArg(), ctx.userID, int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx.mock.ExpectQuery(selectRecordQuery).
		WithArgs(ctx.userID, int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "title", "description", "group_name", "relative_path", "canonical_path", "tags", "created_at", "updated_at"}).
			AddRow(int64(7), ctx.userID, "DeckTwo", nil, "Updated deck", "decktwo", newRel, newCanonical, "[\"tag2\"]", now, now))

	payload := map[string]any{
		"name":        "DeckTwo",
		"description": "Updated deck",
		"tags":        []string{"Tag2"},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/ppts/7", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.authorize(req)

	rec := httptest.NewRecorder()
	ctx.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Name       string   `json:"name"`
		GroupName  string   `json:"groupName"`
		Tags       []string `json:"tags"`
		PathStatus string   `json:"pathStatus"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, "DeckTwo", resp.Name)
	require.Equal(t, "decktwo", resp.GroupName)
	require.ElementsMatch(t, []string{"tag2"}, resp.Tags)
	require.Equal(t, "valid", resp.PathStatus)

	_, statErr := os.Stat(newCanonical)
	require.NoError(t, statErr)

	require.NoError(t, ctx.mock.ExpectationsWereMet())
}

func TestDeletePptRecord(t *testing.T) {
	ctx := newRecordsTestContext(t)

	ctx.mock.ExpectExec("DELETE FROM ppt_records WHERE user_id = \\? AND id = \\?").
		WithArgs(ctx.userID, int64(77)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/ppts/77", nil)
	ctx.authorize(req)

	rec := httptest.NewRecorder()
	ctx.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.NoError(t, ctx.mock.ExpectationsWereMet())
}
