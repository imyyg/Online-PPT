package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"online-ppt/internal/auth"
	"online-ppt/internal/records"
)

const recordNotFoundMsg = "record not found"

// List handles GET /ppts.
func (h *RecordsHandler) List(c *gin.Context) {
	if h.service == nil {
		writeError(c, http.StatusInternalServerError, "server_error", "records service unavailable")
		return
	}

	claims, err := h.authorize(c)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	filters, ok := parseListFilters(c)
	if !ok {
		return
	}

	result, err := h.service.ListRecords(c.Request.Context(), records.ListParams{
		UserID:  claims.UserID,
		Filters: filters,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	items := make([]gin.H, 0, len(result.Records))
	for _, view := range result.Records {
		items = append(items, makeRecordResponse(view))
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  result.Total,
		"limit":  result.Limit,
		"offset": result.Offset,
		"items":  items,
	})
}

// Get handles GET /ppts/{id}.
func (h *RecordsHandler) Get(c *gin.Context) {
	claims, recordID, ok := h.authenticateWithID(c)
	if !ok {
		return
	}

	view, err := h.service.GetRecord(c.Request.Context(), claims.UserID, recordID)
	if err != nil {
		switch {
		case errors.Is(err, records.ErrRecordNotFound):
			writeError(c, http.StatusNotFound, "not_found", recordNotFoundMsg)
		default:
			writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, makeRecordResponse(view))
}

// Update handles PATCH /ppts/{id}.
func (h *RecordsHandler) Update(c *gin.Context) {
	claims, recordID, ok := h.authenticateWithID(c)
	if !ok {
		return
	}

	var req struct {
		Name        *string   `json:"name"`
		Description *string   `json:"description"`
		Tags        *[]string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	view, err := h.service.UpdateRecord(c.Request.Context(), records.UpdateParams{
		UserID:      claims.UserID,
		UserUUID:    claims.UserUUID,
		RecordID:    recordID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
	})
	if err != nil {
		switch {
		case errors.Is(err, records.ErrInvalidRecordName):
			writeError(c, http.StatusBadRequest, "invalid_name", err.Error())
		case errors.Is(err, records.ErrDuplicateRecord):
			writeError(c, http.StatusConflict, "record_exists", err.Error())
		case errors.Is(err, records.ErrRecordNotFound):
			writeError(c, http.StatusNotFound, "not_found", recordNotFoundMsg)
		default:
			writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, makeRecordResponse(view))
}

// Delete handles DELETE /ppts/{id}.
func (h *RecordsHandler) Delete(c *gin.Context) {
	claims, recordID, ok := h.authenticateWithID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteRecord(c.Request.Context(), claims.UserID, recordID); err != nil {
		switch {
		case errors.Is(err, records.ErrRecordNotFound):
			writeError(c, http.StatusNotFound, "not_found", recordNotFoundMsg)
		default:
			writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RecordsHandler) authenticateWithID(c *gin.Context) (*auth.Claims, int64, bool) {
	if h.service == nil {
		writeError(c, http.StatusInternalServerError, "server_error", "records service unavailable")
		return nil, 0, false
	}

	claims, err := h.authorize(c)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return nil, 0, false
	}

	idStr := c.Param("id")
	recordID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_id", "record id must be an integer")
		return nil, 0, false
	}

	if recordID <= 0 {
		writeError(c, http.StatusBadRequest, "invalid_id", "record id must be positive")
		return nil, 0, false
	}

	return claims, recordID, true
}

func parseListFilters(c *gin.Context) (records.ListFilters, bool) {
	filters := records.ListFilters{
		Query: c.Query("q"),
		Tag:   c.Query("tag"),
		Sort:  c.Query("sort"),
	}

	if v := c.Query("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_limit", "limit must be an integer")
			return records.ListFilters{}, false
		}
		filters.Limit = limit
	}

	if v := c.Query("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_offset", "offset must be an integer")
			return records.ListFilters{}, false
		}
		filters.Offset = offset
	}

	return filters, true
}
