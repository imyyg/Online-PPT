package records

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"online-ppt/internal/storage"
)

var (
	// ErrInvalidRecordName reports validation failure for the record name field.
	ErrInvalidRecordName = errors.New("invalid record name")
	// ErrDuplicateRecord signals an existing record conflicts with the new entry.
	ErrDuplicateRecord = errors.New("record already exists")
	// ErrRecordNotFound indicates the requested record does not exist for the user.
	ErrRecordNotFound  = errors.New("record not found")
	errInvalidUserID   = errors.New("invalid user id")
	errInvalidRecordID = errors.New("invalid record id")
	errMissingUserUUID = errors.New("user uuid required for rename")
)

// Service orchestrates business rules around PPT records.
type Service struct {
	repo              *Repository
	presentationsRoot string
	audit             *storage.AuditLogger
	clockFn           func() time.Time
}

// RecordView enriches a record with runtime metadata for presentation.
type RecordView struct {
	Record     PptRecord
	PathStatus string
}

// ListResult bundles paginated record results.
type ListResult struct {
	Records []RecordView
	Total   int
	Limit   int
	Offset  int
}

// ListParams collects the inputs for fetching records.
type ListParams struct {
	UserID  int64
	Filters ListFilters
}

// CreateParams captures the inputs for creating a PPT record.
type CreateParams struct {
	UserID      int64
	UserUUID    string
	Name        string
	Description string
	Tags        []string
}

// UpdateParams describes a partial update request for a record.
type UpdateParams struct {
	UserID      int64
	UserUUID    string
	RecordID    int64
	Name        *string
	Description *string
	Tags        *[]string
}

// NewService constructs a Service instance with validated dependencies.
func NewService(repo *Repository, presentationsRoot string, audit *storage.AuditLogger) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("records service requires repository")
	}
	if presentationsRoot == "" {
		return nil, fmt.Errorf("records service requires presentations root")
	}

	absRoot, err := filepath.Abs(presentationsRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve presentations root: %w", err)
	}

	if audit == nil {
		audit = storage.NewAuditLogger(nil)
	}

	return &Service{
		repo:              repo,
		presentationsRoot: absRoot,
		audit:             audit,
		clockFn:           time.Now,
	}, nil
}

// CreateRecord provisions filesystem structure and persists PPT metadata.
func (s *Service) CreateRecord(ctx context.Context, params CreateParams) (RecordView, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		s.audit.Log("records.create", map[string]any{
			"status": "validation_failed",
			"userId": params.UserID,
			"reason": "name required",
		})
		return RecordView{}, fmt.Errorf("name required: %w", ErrInvalidRecordName)
	}
	if !groupNamePattern.MatchString(name) {
		s.audit.Log("records.create", map[string]any{
			"status": "validation_failed",
			"userId": params.UserID,
			"reason": ErrInvalidRecordName.Error(),
		})
		return RecordView{}, ErrInvalidRecordName
	}
	if params.UserID <= 0 {
		s.audit.Log("records.create", map[string]any{
			"status": "validation_failed",
			"reason": errInvalidUserID.Error(),
		})
		return RecordView{}, errInvalidUserID
	}

	groupName := strings.ToLower(name)

	tags, err := normalizeTags(params.Tags)
	if err != nil {
		s.audit.Log("records.create", map[string]any{
			"status": "validation_failed",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return RecordView{}, err
	}

	paths, err := BuildPaths(s.presentationsRoot, params.UserUUID, groupName)
	if err != nil {
		s.audit.Log("records.create", map[string]any{
			"status": "error",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return RecordView{}, err
	}

	if err := EnsureDirectories(paths); err != nil {
		s.audit.Log("records.create", map[string]any{
			"status": "error",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return RecordView{}, err
	}

	record := PptRecord{
		UserID:        params.UserID,
		Name:          name,
		GroupName:     groupName,
		RelativePath:  paths.Relative,
		CanonicalPath: paths.Canonical,
		Tags:          tags,
	}
	if desc := strings.TrimSpace(params.Description); desc != "" {
		record.Description = sql.NullString{String: desc, Valid: true}
	}

	created, err := s.repo.Create(ctx, record)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			s.audit.Log("records.create", map[string]any{
				"status": "conflict",
				"userId": params.UserID,
				"reason": ErrDuplicateRecord.Error(),
			})
			return RecordView{}, ErrDuplicateRecord
		}
		s.audit.Log("records.create", map[string]any{
			"status": "error",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return RecordView{}, err
	}

	view, err := s.makeRecordView(created)
	if err != nil {
		s.audit.Log("records.create", map[string]any{
			"status": "error",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return RecordView{}, err
	}

	s.audit.Log("records.create", map[string]any{
		"status":   "success",
		"userId":   params.UserID,
		"recordId": created.ID,
	})
	return view, nil
}

// ListRecords retrieves records for a user using optional filters.
func (s *Service) ListRecords(ctx context.Context, params ListParams) (ListResult, error) {
	if params.UserID <= 0 {
		s.audit.Log("records.list", map[string]any{
			"status": "validation_failed",
			"reason": errInvalidUserID.Error(),
		})
		return ListResult{}, errInvalidUserID
	}

	normalized := normalizeListFilters(params.Filters)
	records, total, err := s.repo.List(ctx, params.UserID, normalized)
	if err != nil {
		s.audit.Log("records.list", map[string]any{
			"status": "error",
			"userId": params.UserID,
			"reason": err.Error(),
		})
		return ListResult{}, err
	}

	result := ListResult{
		Records: make([]RecordView, 0, len(records)),
		Total:   total,
		Limit:   normalized.Limit,
		Offset:  normalized.Offset,
	}

	for _, record := range records {
		view, err := s.makeRecordView(record)
		if err != nil {
			return ListResult{}, err
		}
		result.Records = append(result.Records, view)
	}

	return result, nil
}

// GetRecord returns a single record for the user.
func (s *Service) GetRecord(ctx context.Context, userID, recordID int64) (RecordView, error) {
	if userID <= 0 {
		s.audit.Log("records.get", map[string]any{
			"status": "validation_failed",
			"reason": errInvalidUserID.Error(),
		})
		return RecordView{}, errInvalidUserID
	}
	if recordID <= 0 {
		s.audit.Log("records.get", map[string]any{
			"status": "validation_failed",
			"userId": userID,
			"reason": errInvalidRecordID.Error(),
		})
		return RecordView{}, errInvalidRecordID
	}

	record, err := s.repo.GetByID(ctx, userID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("records.get", map[string]any{
				"status":   "not_found",
				"userId":   userID,
				"recordId": recordID,
			})
			return RecordView{}, ErrRecordNotFound
		}
		s.audit.Log("records.get", map[string]any{
			"status":   "error",
			"userId":   userID,
			"recordId": recordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	return s.makeRecordView(record)
}

// UpdateRecord applies partial updates to a record and returns the refreshed view.
func (s *Service) UpdateRecord(ctx context.Context, params UpdateParams) (RecordView, error) {
	if params.UserID <= 0 {
		s.audit.Log("records.update", map[string]any{
			"status": "validation_failed",
			"reason": errInvalidUserID.Error(),
		})
		return RecordView{}, errInvalidUserID
	}
	if params.RecordID <= 0 {
		s.audit.Log("records.update", map[string]any{
			"status": "validation_failed",
			"userId": params.UserID,
			"reason": errInvalidRecordID.Error(),
		})
		return RecordView{}, errInvalidRecordID
	}

	current, err := s.repo.GetByID(ctx, params.UserID, params.RecordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("records.update", map[string]any{
				"status":   "not_found",
				"userId":   params.UserID,
				"recordId": params.RecordID,
			})
			return RecordView{}, ErrRecordNotFound
		}
		s.audit.Log("records.update", map[string]any{
			"status":   "error",
			"userId":   params.UserID,
			"recordId": params.RecordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	updated := current

	if err := s.applyNameUpdate(&updated, params.UserUUID, params.Name); err != nil {
		s.audit.Log("records.update", map[string]any{
			"status":   "validation_failed",
			"userId":   params.UserID,
			"recordId": params.RecordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	applyDescriptionUpdate(&updated, params.Description)

	if err := applyTagsUpdate(&updated, params.Tags); err != nil {
		s.audit.Log("records.update", map[string]any{
			"status":   "validation_failed",
			"userId":   params.UserID,
			"recordId": params.RecordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	saved, err := s.repo.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("records.update", map[string]any{
				"status":   "not_found",
				"userId":   params.UserID,
				"recordId": params.RecordID,
			})
			return RecordView{}, ErrRecordNotFound
		}
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			s.audit.Log("records.update", map[string]any{
				"status":   "conflict",
				"userId":   params.UserID,
				"recordId": params.RecordID,
				"reason":   ErrDuplicateRecord.Error(),
			})
			return RecordView{}, ErrDuplicateRecord
		}
		s.audit.Log("records.update", map[string]any{
			"status":   "error",
			"userId":   params.UserID,
			"recordId": params.RecordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	view, err := s.makeRecordView(saved)
	if err != nil {
		s.audit.Log("records.update", map[string]any{
			"status":   "error",
			"userId":   params.UserID,
			"recordId": params.RecordID,
			"reason":   err.Error(),
		})
		return RecordView{}, err
	}

	s.audit.Log("records.update", map[string]any{
		"status":   "success",
		"userId":   params.UserID,
		"recordId": saved.ID,
	})
	return view, nil
}

// DeleteRecord removes the record owned by the user.
func (s *Service) DeleteRecord(ctx context.Context, userID, recordID int64) error {
	if userID <= 0 {
		s.audit.Log("records.delete", map[string]any{
			"status": "validation_failed",
			"reason": errInvalidUserID.Error(),
		})
		return errInvalidUserID
	}
	if recordID <= 0 {
		s.audit.Log("records.delete", map[string]any{
			"status": "validation_failed",
			"userId": userID,
			"reason": errInvalidRecordID.Error(),
		})
		return errInvalidRecordID
	}

	if err := s.repo.Delete(ctx, userID, recordID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("records.delete", map[string]any{
				"status":   "not_found",
				"userId":   userID,
				"recordId": recordID,
			})
			return ErrRecordNotFound
		}
		s.audit.Log("records.delete", map[string]any{
			"status":   "error",
			"userId":   userID,
			"recordId": recordID,
			"reason":   err.Error(),
		})
		return err
	}

	s.audit.Log("records.delete", map[string]any{
		"status":   "success",
		"userId":   userID,
		"recordId": recordID,
	})
	return nil
}

func (s *Service) makeRecordView(record PptRecord) (RecordView, error) {
	status, err := s.computePathStatus(record.CanonicalPath)
	if err != nil {
		return RecordView{}, err
	}
	return RecordView{Record: record, PathStatus: status}, nil
}

func (s *Service) computePathStatus(canonicalPath string) (string, error) {
	if canonicalPath == "" {
		return "missing", nil
	}

	absCanonical, err := filepath.Abs(canonicalPath)
	if err != nil {
		return "", fmt.Errorf("resolve canonical path: %w", err)
	}

	rel, err := filepath.Rel(s.presentationsRoot, absCanonical)
	if err != nil {
		return "", fmt.Errorf("resolve relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, "..") {
		return "outside_root", nil
	}

	if _, statErr := os.Stat(absCanonical); statErr != nil {
		if errors.Is(statErr, fs.ErrNotExist) {
			return "missing", nil
		}
		return "", fmt.Errorf("stat canonical path: %w", statErr)
	}

	return "valid", nil
}

func (s *Service) applyNameUpdate(record *PptRecord, userUUID string, namePtr *string) error {
	if namePtr == nil {
		return nil
	}

	name := strings.TrimSpace(*namePtr)
	if name == "" {
		return fmt.Errorf("name required: %w", ErrInvalidRecordName)
	}
	if !groupNamePattern.MatchString(name) {
		return ErrInvalidRecordName
	}
	if userUUID == "" {
		return errMissingUserUUID
	}

	record.Name = name
	record.GroupName = strings.ToLower(name)

	paths, err := BuildPaths(s.presentationsRoot, userUUID, record.GroupName)
	if err != nil {
		return err
	}
	if err := EnsureDirectories(paths); err != nil {
		return err
	}
	record.RelativePath = paths.Relative
	record.CanonicalPath = paths.Canonical
	return nil
}

func applyDescriptionUpdate(record *PptRecord, description *string) {
	if description == nil {
		return
	}

	desc := strings.TrimSpace(*description)
	if desc == "" {
		record.Description = sql.NullString{}
		return
	}

	record.Description = sql.NullString{String: desc, Valid: true}
}

func applyTagsUpdate(record *PptRecord, tags *[]string) error {
	if tags == nil {
		return nil
	}

	normalized, err := normalizeTags(*tags)
	if err != nil {
		return err
	}
	record.Tags = normalized
	return nil
}

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	if len(tags) > 10 {
		return nil, fmt.Errorf("too many tags; maximum is 10")
	}

	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		slug := strings.ToLower(trimmed)
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		normalized = append(normalized, slug)
	}

	return normalized, nil
}
