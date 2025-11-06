package auth

import (
	databaseSql "database/sql"
	time "time"
)

// UserAccount mirrors the user_accounts table.
type UserAccount struct {
	ID           int64
	UUID         string
	Email        string
	PasswordHash string
	Status       string
	LastLoginAt  databaseSql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserSession mirrors the user_sessions table.
type UserSession struct {
	ID                int64
	UserID            int64
	RefreshTokenHash  string
	ExpiresAt         time.Time
	IssuedAt          time.Time
	ClientFingerprint databaseSql.NullString
	RevokedAt         databaseSql.NullTime
	CreatedAt         time.Time
}

type scanner interface {
	Scan(dest ...any) error
}

// scanUserAccount builds a UserAccount from the current row.
func scanUserAccount(row scanner) (UserAccount, error) {
	var ua UserAccount
	if err := row.Scan(
		&ua.ID,
		&ua.UUID,
		&ua.Email,
		&ua.PasswordHash,
		&ua.Status,
		&ua.LastLoginAt,
		&ua.CreatedAt,
		&ua.UpdatedAt,
	); err != nil {
		return UserAccount{}, err
	}
	return ua, nil
}

// scanUserSession builds a UserSession from the current row.
func scanUserSession(row scanner) (UserSession, error) {
	var us UserSession
	if err := row.Scan(
		&us.ID,
		&us.UserID,
		&us.RefreshTokenHash,
		&us.ExpiresAt,
		&us.IssuedAt,
		&us.ClientFingerprint,
		&us.RevokedAt,
		&us.CreatedAt,
	); err != nil {
		return UserSession{}, err
	}
	return us, nil
}
