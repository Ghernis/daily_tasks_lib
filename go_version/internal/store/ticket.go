package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type DailyTicket struct {
	ID          int64
	EntryDate   string
	Title       string
	Description string
	RITMNumber  sql.NullString
}

type TaskInput struct {
	ProjectID       int64
	Hours           float64
	TaskDescription string
	SubtaskNames    []string
}

func SaveDailyTicket(ctx context.Context, db *sql.DB, entryDate, title, description string, tasks []TaskInput) error {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ticketID int64
	var existingTitle string
	err = tx.QueryRowContext(ctx,
		`SELECT id, title FROM ticket_daily WHERE "date" = ?`, entryDate,
	).Scan(&ticketID, &existingTitle)

	if errors.Is(err, sql.ErrNoRows) {
		if title == "" {
			title = fmt.Sprintf("Daily work %s", entryDate)
		}
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		res, err := tx.ExecContext(ctx,
			`INSERT INTO ticket_daily ("date", title, description, created_at) VALUES (?, ?, ?, ?)`,
			entryDate, title, description, createdAt,
		)
		if err != nil {
			return err
		}
		ticketID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if title == "" {
			title = existingTitle
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE ticket_daily SET title = ?, description = ? WHERE id = ?`,
			title, description, ticketID,
		); err != nil {
			return err
		}
		taskRows, err := tx.QueryContext(ctx, `SELECT id FROM ticket_daily_task WHERE ticket_daily_id = ?`, ticketID)
		if err != nil {
			return err
		}
		var taskIDs []int64
		for taskRows.Next() {
			var tid int64
			if err := taskRows.Scan(&tid); err != nil {
				taskRows.Close()
				return err
			}
			taskIDs = append(taskIDs, tid)
		}
		taskRows.Close()
		for _, tid := range taskIDs {
			if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_daily_task_subtask WHERE ticket_daily_task_id = ?`, tid); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_daily_task WHERE ticket_daily_id = ?`, ticketID); err != nil {
			return err
		}
	}

	for order, task := range tasks {
		var exists int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM project WHERE id = ?`, task.ProjectID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			INSERT INTO ticket_daily_task (ticket_daily_id, project_id, hours, task_description, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, ticketID, task.ProjectID, task.Hours, strings.TrimSpace(task.TaskDescription), order)
		if err != nil {
			return err
		}
		taskRowID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		subOrder := 0
		for _, raw := range task.SubtaskNames {
			name := strings.TrimSpace(raw)
			if name == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO ticket_daily_task_subtask (ticket_daily_task_id, name, sort_order) VALUES (?, ?, ?)`,
				taskRowID, name, subOrder,
			); err != nil {
				return err
			}
			subOrder++
		}
	}
	return tx.Commit()
}

func SetRITMForDate(ctx context.Context, db *sql.DB, entryDate, ritm string) (bool, error) {
	ritm = strings.TrimSpace(ritm)
	var ritmVal interface{}
	if ritm == "" {
		ritmVal = nil
	} else {
		ritmVal = ritm
	}
	res, err := db.ExecContext(ctx, `UPDATE ticket_daily SET ritm_number = ? WHERE "date" = ?`, ritmVal, entryDate)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func GetTicketByDate(ctx context.Context, db *sql.DB, entryDate string) (*DailyTicket, error) {
	var t DailyTicket
	err := db.QueryRowContext(ctx,
		`SELECT id, "date", title, description, ritm_number FROM ticket_daily WHERE "date" = ?`, entryDate,
	).Scan(&t.ID, &t.EntryDate, &t.Title, &t.Description, &t.RITMNumber)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
