package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultMigrationDir = "migrations"

// Migrator executes plain SQL migration files in lexical order.
type Migrator struct {
	Dir string
}

// Apply walks the migration directory and runs each *.sql file sequentially.
func (m Migrator) Apply(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("nil database handle")
	}

	dir := m.Dir
	if dir == "" {
		dir = defaultMigrationDir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	files := filterSQL(entries)
	sort.Strings(files)

	for _, name := range files {
		path := filepath.Join(dir, name)
		if err := runSQLFile(ctx, db, path); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}

func filterSQL(entries []fs.DirEntry) []string {
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	return files
}

func runSQLFile(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	stmts := splitStatements(string(data))
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec statement: %w", err)
		}
	}
	return nil
}

func splitStatements(sqlText string) []string {
	raw := strings.Split(sqlText, ";")
	result := make([]string, 0, len(raw))
	for _, stmt := range raw {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
