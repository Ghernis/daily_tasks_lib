package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// ResolveProject finds a project by numeric id (if id > 0) or by exact name.
func ResolveProject(ctx context.Context, db *sql.DB, id int64, name string) (*Project, error) {
	if id > 0 {
		p, err := GetProjectByID(ctx, db, id)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, ErrProjectNotFound
		}
		return p, nil
	}
	name = strings.TrimSpace(name)
	if name != "" {
		p, err := GetProjectByName(ctx, db, name)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, ErrProjectNotFound
		}
		return p, nil
	}
	return nil, fmt.Errorf("project id or name is required")
}

// ResolveProjectRef parses a CLI token as id (digits) or project name.
func ResolveProjectRef(ctx context.Context, db *sql.DB, ref string) (*Project, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("project id or name is required")
	}
	if id, err := strconv.ParseInt(ref, 10, 64); err == nil && id > 0 {
		return ResolveProject(ctx, db, id, "")
	}
	return ResolveProject(ctx, db, 0, ref)
}
