package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"online-ppt/internal/auth"
	"online-ppt/internal/records"
)

// RecordsHandler exposes record-related HTTP endpoints.
type RecordsHandler struct {
	service *records.Service
	tokens  *auth.TokenManager
}

var errMissingBearer = errors.New("missing or invalid authorization header")

// NewRecordsHandler constructs a handler for PPT record operations.
func NewRecordsHandler(service *records.Service, tokens *auth.TokenManager) *RecordsHandler {
	return &RecordsHandler{service: service, tokens: tokens}
}

// Create handles POST /ppts.
func (h *RecordsHandler) Create(c *gin.Context) {
	if h.service == nil {
		writeError(c, http.StatusInternalServerError, "server_error", "records service unavailable")
		return
	}

	claims, err := h.authorize(c)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	view, err := h.service.CreateRecord(c.Request.Context(), records.CreateParams{
		UserID:      claims.UserID,
		UserUUID:    claims.UserUUID,
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
	})
	if err != nil {
		switch {
		case err == records.ErrInvalidRecordName:
			writeError(c, http.StatusBadRequest, "invalid_name", err.Error())
		case err == records.ErrDuplicateRecord:
			writeError(c, http.StatusConflict, "record_exists", err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, makeRecordResponse(view))
}

func makeRecordResponse(view records.RecordView) gin.H {
	record := view.Record

	var title any
	if record.Title.Valid {
		title = record.Title.String
	}

	var description any
	if record.Description.Valid {
		description = record.Description.String
	}

	tags := record.Tags
	if tags == nil {
		tags = []string{}
	}

	return gin.H{
		"id":            record.ID,
		"name":          record.Name,
		"title":         title,
		"groupName":     record.GroupName,
		"description":   description,
		"relativePath":  record.RelativePath,
		"canonicalPath": record.CanonicalPath,
		"tags":          tags,
		"pathStatus":    view.PathStatus,
		"createdAt":     record.CreatedAt,
		"updatedAt":     record.UpdatedAt,
	}
}

func (h *RecordsHandler) authorize(c *gin.Context) (*auth.Claims, error) {
	if h.tokens == nil {
		return nil, errMissingBearer
	}

	header := c.GetHeader("Authorization")
	if header == "" {
		return nil, errMissingBearer
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return nil, errMissingBearer
	}

	claims, err := h.tokens.ParseAccessToken(parts[1])
	if err != nil {
		return nil, errMissingBearer
	}
	return claims, nil
}
