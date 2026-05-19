package export

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"daily-tasks/go_version/internal/store"
)

func Build(ctx context.Context, db *sql.DB, dateFrom, dateTo *string) (*Document, error) {
	query := `SELECT id, "date", title, description, ritm_number FROM ticket_daily WHERE 1=1`
	var args []interface{}
	if dateFrom != nil {
		query += ` AND "date" >= ?`
		args = append(args, *dateFrom)
	}
	if dateTo != nil {
		query += ` AND "date" <= ?`
		args = append(args, *dateTo)
	}
	query += ` ORDER BY "date"`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors, err := authorsForExport(ctx, db)
	if err != nil {
		return nil, err
	}

	var tickets []TicketDaily
	for rows.Next() {
		var id int64
		var entryDate, title, description string
		var ritm sql.NullString
		if err := rows.Scan(&id, &entryDate, &title, &description, &ritm); err != nil {
			return nil, err
		}
		tasks, err := loadTasks(ctx, db, id)
		if err != nil {
			return nil, err
		}
		td := TicketDaily{
			Title:       title,
			Date:        normalizeDate(entryDate),
			Description: description,
			Author:      authors,
			Tasks:       tasks,
		}
		if ritm.Valid {
			s := ritm.String
			td.RITMNumber = &s
		}
		tickets = append(tickets, td)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tickets == nil {
		tickets = []TicketDaily{}
	}
	return &Document{
		ExportedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000000Z"),
		Tickets:    tickets,
	}, nil
}

func authorsForExport(ctx context.Context, db *sql.DB) ([]Author, error) {
	a, err := store.GetAuthorProfile(ctx, db)
	if err != nil {
		return nil, err
	}
	if a == nil || strings.TrimSpace(a.Name) == "" {
		return []Author{}, nil
	}
	return []Author{{
		Name:  strings.TrimSpace(a.Name),
		Email: a.Email,
		SNID:  a.SNID,
	}}, nil
}

func loadTasks(ctx context.Context, db *sql.DB, ticketID int64) ([]Task, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT t.id, t.project_id, t.hours, t.task_description, p.name, p.description
		FROM ticket_daily_task t
		JOIN project p ON p.id = t.project_id
		WHERE t.ticket_daily_id = ?
		ORDER BY t.sort_order, t.id
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var taskRowID, projectID int64
		var hours float64
		var taskDesc, projName, projDesc string
		if err := rows.Scan(&taskRowID, &projectID, &hours, &taskDesc, &projName, &projDesc); err != nil {
			return nil, err
		}
		desc := strings.TrimSpace(taskDesc)
		if desc == "" {
			desc = strings.TrimSpace(projDesc)
		}
		subs, err := loadSubtasks(ctx, db, taskRowID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, Task{
			Name:        projName,
			Description: desc,
			Subtasks:    subs,
			Time:        hours,
		})
	}
	return tasks, rows.Err()
}

func loadSubtasks(ctx context.Context, db *sql.DB, taskRowID int64) ([]Subtask, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name FROM ticket_daily_task_subtask
		WHERE ticket_daily_task_id = ?
		ORDER BY sort_order, id
	`, taskRowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []Subtask
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		subs = append(subs, Subtask{Name: n})
	}
	if subs == nil {
		subs = []Subtask{}
	}
	return subs, rows.Err()
}

func normalizeDate(raw string) string {
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}
