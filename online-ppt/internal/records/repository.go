package records

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repository provides persistence helpers for ppt_records.
type Repository struct {
	db *sql.DB
}

// PptRecord represents a ppt_records row.
type PptRecord struct {
	ID            int64
	UserID        int64
	Name          string
	Title         sql.NullString
	Description   sql.NullString
	GroupName     string
	RelativePath  string
	CanonicalPath string
	Tags          []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ListFilters captures optional filters for listing.
type ListFilters struct {
	Query  string
	Tag    string
	Sort   string
	Limit  int
	Offset int
}

// NewRepository instantiates a Repository.
func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("records repository requires db handle")
	}
	return &Repository{db: db}, nil
}

// WithTx executes fn within a transaction boundary.
func (r *Repository) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return fmt.Errorf("tx function: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	committed = true
	return nil
}

// RepositoryTx wraps sql.Tx for helper usage.
type RepositoryTx struct {
	tx *sql.Tx
}

// NewRepositoryTx converts a transaction into a RepositoryTx helper.
func NewRepositoryTx(tx *sql.Tx) RepositoryTx {
	return RepositoryTx{tx: tx}
}

// CreateWithinTx inserts a record inside an existing transaction.
func (rt RepositoryTx) CreateWithinTx(ctx context.Context, record PptRecord) (int64, error) {
	tagsJSON, err := marshalTags(record.Tags)
	if err != nil {
		return 0, err
	}

	stmt := `INSERT INTO ppt_records (user_id, name, title, description, group_name, relative_path, canonical_path, tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := rt.tx.ExecContext(ctx, stmt,
		record.UserID,
		record.Name,
		record.Title,
		record.Description,
		record.GroupName,
		record.RelativePath,
		record.CanonicalPath,
		tagsJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("insert ppt_record tx: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("derive record id: %w", err)
	}
	return id, nil
}

// Create inserts a new PPT record row.
func (r *Repository) Create(ctx context.Context, record PptRecord) (PptRecord, error) {
	tagsJSON, err := marshalTags(record.Tags)
	if err != nil {
		return PptRecord{}, err
	}

	stmt := `INSERT INTO ppt_records (user_id, name, title, description, group_name, relative_path, canonical_path, tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, stmt,
		record.UserID,
		record.Name,
		record.Title,
		record.Description,
		record.GroupName,
		record.RelativePath,
		record.CanonicalPath,
		tagsJSON,
	)
	if err != nil {
		return PptRecord{}, fmt.Errorf("insert ppt_record: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return PptRecord{}, fmt.Errorf("derive record id: %w", err)
	}

	return r.GetByID(ctx, record.UserID, id)
}

// GetByID fetches a single record for a user.
func (r *Repository) GetByID(ctx context.Context, userID, id int64) (PptRecord, error) {
	stmt := `SELECT id, user_id, name, title, description, group_name, relative_path, canonical_path, tags, created_at, updated_at FROM ppt_records WHERE user_id = ? AND id = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, stmt, userID, id)
	return scanRecord(row)
}

// Delete removes a record by user and id.
func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	stmt := `DELETE FROM ppt_records WHERE user_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, stmt, userID, id)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Update modifies editable fields of a record.
func (r *Repository) Update(ctx context.Context, record PptRecord) (PptRecord, error) {
	tagsJSON, err := marshalTags(record.Tags)
	if err != nil {
		return PptRecord{}, err
	}

	stmt := `UPDATE ppt_records SET name = ?, title = ?, description = ?, group_name = ?, relative_path = ?, canonical_path = ?, tags = ?, updated_at = NOW() WHERE user_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, stmt,
		record.Name,
		record.Title,
		record.Description,
		record.GroupName,
		record.RelativePath,
		record.CanonicalPath,
		tagsJSON,
		record.UserID,
		record.ID,
	)
	if err != nil {
		return PptRecord{}, fmt.Errorf("update record: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return PptRecord{}, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return PptRecord{}, sql.ErrNoRows
	}

	return r.GetByID(ctx, record.UserID, record.ID)
}

// List returns records filtered by the provided criteria.
func (r *Repository) List(ctx context.Context, userID int64, filters ListFilters) ([]PptRecord, int, error) {
	normalized := normalizeListFilters(filters)
	builder := newListQueryBuilder(userID, normalized)

	countSQL, countArgs := builder.countQuery()
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count records: %w", err)
	}

	selectSQL, selectArgs := builder.selectQuery()
	rows, err := r.db.QueryContext(ctx, selectSQL, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list records: %w", err)
	}
	defer rows.Close()

	var results []PptRecord
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func scanRecord(row interface{ Scan(dest ...any) error }) (PptRecord, error) {
	var (
		record     PptRecord
		tagsString sql.NullString
	)

	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.Name,
		&record.Title,
		&record.Description,
		&record.GroupName,
		&record.RelativePath,
		&record.CanonicalPath,
		&tagsString,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return PptRecord{}, err
	}

	if tagsString.Valid {
		if err := json.Unmarshal([]byte(tagsString.String), &record.Tags); err != nil {
			return PptRecord{}, fmt.Errorf("unmarshal tags: %w", err)
		}
	}

	return record, nil
}

func marshalTags(tags []string) (sql.NullString, error) {
	if len(tags) == 0 {
		return sql.NullString{}, nil
	}

	encoded, err := json.Marshal(tags)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("marshal tags: %w", err)
	}

	return sql.NullString{String: string(encoded), Valid: true}, nil
}
