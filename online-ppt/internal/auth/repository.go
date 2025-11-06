package auth

import (
	context "context"
	databaseSql "database/sql"
	fmt "fmt"
	time "time"
)

// Repository provides persistence operations for users and sessions.
type Repository struct {
	db *databaseSql.DB
}

// NewRepository creates a new Repository instance.
func NewRepository(db *databaseSql.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("auth repository requires db handle")
	}
	return &Repository{db: db}, nil
}

// CreateUser inserts a new user row and returns the hydrated entity.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, uuid string) (UserAccount, error) {
	stmt := `INSERT INTO user_accounts (email, password_hash, uuid) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, stmt, email, passwordHash, uuid)
	if err != nil {
		return UserAccount{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return UserAccount{}, fmt.Errorf("derive user id: %w", err)
	}

	return r.GetUserByID(ctx, id)
}

// GetUserByEmail fetches a user by normalized email address.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (UserAccount, error) {
	stmt := `SELECT id, uuid, email, password_hash, status, last_login_at, created_at, updated_at FROM user_accounts WHERE email = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, stmt, email)
	user, err := scanUserAccount(row)
	if err != nil {
		return UserAccount{}, err
	}
	return user, nil
}

// GetUserByID fetches a user by primary key.
func (r *Repository) GetUserByID(ctx context.Context, id int64) (UserAccount, error) {
	stmt := `SELECT id, uuid, email, password_hash, status, last_login_at, created_at, updated_at FROM user_accounts WHERE id = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, stmt, id)
	user, err := scanUserAccount(row)
	if err != nil {
		return UserAccount{}, err
	}
	return user, nil
}

// UpdateLastLogin stores the latest successful login timestamp.
func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64, ts time.Time) error {
	stmt := `UPDATE user_accounts SET last_login_at = ?, updated_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, stmt, ts, userID); err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

// CreateSession persists a refresh token entry.
func (r *Repository) CreateSession(ctx context.Context, session UserSession) (UserSession, error) {
	stmt := `INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, issued_at, client_fingerprint, revoked_at) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, stmt,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
		session.IssuedAt,
		session.ClientFingerprint,
		session.RevokedAt,
	)
	if err != nil {
		return UserSession{}, fmt.Errorf("insert session: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return UserSession{}, fmt.Errorf("derive session id: %w", err)
	}

	return r.GetSessionByID(ctx, id)
}

// GetSessionByID fetches a session row.
func (r *Repository) GetSessionByID(ctx context.Context, id int64) (UserSession, error) {
	stmt := `SELECT id, user_id, refresh_token_hash, expires_at, issued_at, client_fingerprint, revoked_at, created_at FROM user_sessions WHERE id = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, stmt, id)
	return scanUserSession(row)
}

// FindActiveSession locates an active session by hash.
func (r *Repository) FindActiveSession(ctx context.Context, hash string) (UserSession, error) {
	stmt := `SELECT id, user_id, refresh_token_hash, expires_at, issued_at, client_fingerprint, revoked_at, created_at FROM user_sessions WHERE refresh_token_hash = ? AND revoked_at IS NULL LIMIT 1`
	row := r.db.QueryRowContext(ctx, stmt, hash)
	return scanUserSession(row)
}

// RevokeSession marks a session as revoked.
func (r *Repository) RevokeSession(ctx context.Context, sessionID int64, revokedAt time.Time) error {
	stmt := `UPDATE user_sessions SET revoked_at = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, stmt, revokedAt, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// PurgeExpiredSessions removes sessions whose expiry is in the past.
func (r *Repository) PurgeExpiredSessions(ctx context.Context, cutoff time.Time) (int64, error) {
	stmt := `DELETE FROM user_sessions WHERE expires_at < ?`
	res, err := r.db.ExecContext(ctx, stmt, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge expired sessions: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return count, nil
}

// WithTx wraps the provided function in a database transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(tx *databaseSql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	succeeded := false
	defer func() {
		if !succeeded {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return fmt.Errorf("tx function: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	succeeded = true
	return nil
}

// RepositoryTx exposes helpers when running inside a transaction.
type RepositoryTx struct {
	tx *databaseSql.Tx
}

// NewRepositoryTx converts a sql.Tx into a transactional repository helper.
func NewRepositoryTx(tx *databaseSql.Tx) RepositoryTx {
	return RepositoryTx{tx: tx}
}

// CreateSessionTx inserts a session when inside an existing transaction.
func (rt RepositoryTx) CreateSessionTx(ctx context.Context, session UserSession) (int64, error) {
	stmt := `INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, issued_at, client_fingerprint, revoked_at) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := rt.tx.ExecContext(ctx, stmt,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
		session.IssuedAt,
		session.ClientFingerprint,
		session.RevokedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert session tx: %w", err)
	}
	return res.LastInsertId()
}

// UpdateLastLoginTx updates last login timestamp inside a transaction.
func (rt RepositoryTx) UpdateLastLoginTx(ctx context.Context, userID int64, ts time.Time) error {
	stmt := `UPDATE user_accounts SET last_login_at = ?, updated_at = NOW() WHERE id = ?`
	if _, err := rt.tx.ExecContext(ctx, stmt, ts, userID); err != nil {
		return fmt.Errorf("update last login tx: %w", err)
	}
	return nil
}
