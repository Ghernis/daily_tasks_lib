package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

func AddProject(ctx context.Context, db *sql.DB, name string, description string, defaultSubtasks []string) (*Project, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	res, err := db.ExecContext(ctx,
		`INSERT INTO project (name, description, created_at) VALUES (?, ?, ?)`,
		name, description, createdAt,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	for i, raw := range defaultSubtasks {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO project_default_subtask (project_id, title, sort_order) VALUES (?, ?, ?)`,
			id, t, i,
		); err != nil {
			return nil, err
		}
	}
	return GetProjectByID(ctx, db, id)
}

func ListProjects(ctx context.Context, db *sql.DB) ([]Project, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, description, created_at FROM project ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetProjectByID(ctx context.Context, db *sql.DB, id int64) (*Project, error) {
	var p Project
	err := db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at FROM project WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetProjectByName(ctx context.Context, db *sql.DB, name string) (*Project, error) {
	var p Project
	err := db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at FROM project WHERE name = ?`,
		strings.TrimSpace(name),
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func ListDefaultSubtaskTitles(ctx context.Context, db *sql.DB, projectID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT title FROM project_default_subtask
		 WHERE project_id = ? ORDER BY sort_order, id`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func ListPastSubtaskNames(ctx context.Context, db *sql.DB, projectID int64, beforeDate string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT s.name
		FROM ticket_daily_task_subtask s
		JOIN ticket_daily_task t ON t.id = s.ticket_daily_task_id
		JOIN ticket_daily d ON d.id = t.ticket_daily_id
		WHERE t.project_id = ? AND d."date" < ?
		GROUP BY s.name
		ORDER BY MAX(d."date") DESC, s.name
	`, projectID, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func SetDefaultSubtasks(ctx context.Context, db *sql.DB, projectID int64, projectName string, titles []string) error {
	p, err := ResolveProject(ctx, db, projectID, projectName)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_default_subtask WHERE project_id = ?`, p.ID); err != nil {
		return err
	}
	order := 0
	for _, raw := range titles {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_default_subtask (project_id, title, sort_order) VALUES (?, ?, ?)`,
			p.ID, t, order,
		); err != nil {
			return err
		}
		order++
	}
	return tx.Commit()
}

func SetProjectDescription(ctx context.Context, db *sql.DB, projectID int64, projectName, description string) error {
	p, err := ResolveProject(ctx, db, projectID, projectName)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE project SET description = ? WHERE id = ?`,
		strings.TrimSpace(description), p.ID)
	return err
}
