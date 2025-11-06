package records

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var groupNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Paths encapsulates relative and absolute locations for a PPT deck.
type Paths struct {
	Relative  string
	Canonical string
}

// BuildPaths validates inputs and constructs normalized paths under the root.
func BuildPaths(rootDir, userUUID, groupName string) (Paths, error) {
	if rootDir == "" {
		return Paths{}, fmt.Errorf("root directory required")
	}
	if _, err := uuid.Parse(userUUID); err != nil {
		return Paths{}, fmt.Errorf("invalid user uuid: %w", err)
	}
	if !groupNamePattern.MatchString(groupName) {
		return Paths{}, fmt.Errorf("invalid group name")
	}

	cleanRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve root: %w", err)
	}

	relative := filepath.ToSlash(filepath.Join("presentations", userUUID, groupName, "slides"))
	canonical := filepath.Clean(filepath.Join(cleanRoot, userUUID, groupName, "slides"))

	if !strings.HasPrefix(canonical, cleanRoot) {
		return Paths{}, fmt.Errorf("canonical path escapes root")
	}

	return Paths{Relative: relative, Canonical: canonical}, nil
}

// EnsureDirectories makes sure the canonical directory structure exists.
func EnsureDirectories(paths Paths) error {
	if paths.Canonical == "" {
		return fmt.Errorf("canonical path required")
	}
	return os.MkdirAll(paths.Canonical, 0o755)
}
