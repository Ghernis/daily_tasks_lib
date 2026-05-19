package cli

import (
	"context"
	"database/sql"
	"fmt"

	"daily-tasks/go_version/internal/store"
)

func resolveProject(ctx context.Context, db *sql.DB, projectID int64, nameArg string) (*store.Project, error) {
	p, err := store.ResolveProject(ctx, db, projectID, nameArg)
	if err == nil {
		return p, nil
	}
	if err == store.ErrProjectNotFound {
		if projectID > 0 {
			return nil, fmt.Errorf("unknown project id: %d", projectID)
		}
		return nil, fmt.Errorf("unknown project: %q", nameArg)
	}
	return nil, err
}
