package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const authorID = 1

type AuthorProfile struct {
	ID    int64
	Name  string
	Email string
	SNID  string
}

func GetAuthorProfile(ctx context.Context, db *sql.DB) (*AuthorProfile, error) {
	var a AuthorProfile
	err := db.QueryRowContext(ctx,
		`SELECT id, name, email, sn_id FROM author_profile WHERE id = ?`, authorID,
	).Scan(&a.ID, &a.Name, &a.Email, &a.SNID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func SetAuthorProfile(ctx context.Context, db *sql.DB, name, email, snID string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("author name is required")
	}
	email = strings.TrimSpace(email)
	snID = strings.TrimSpace(snID)
	_, err := db.ExecContext(ctx, `
		INSERT INTO author_profile (id, name, email, sn_id) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, email = excluded.email, sn_id = excluded.sn_id
	`, authorID, name, email, snID)
	return err
}

func DefaultTicketDescription(ctx context.Context, db *sql.DB, entryDateISO string) (string, error) {
	a, err := GetAuthorProfile(ctx, db)
	if err != nil {
		return "", err
	}
	if a == nil || strings.TrimSpace(a.Name) == "" {
		return fmt.Sprintf("Tareas realizadas el %s", entryDateISO), nil
	}
	if strings.TrimSpace(a.SNID) != "" {
		return fmt.Sprintf("Tareas realizadas el %s por %s(%s)", entryDateISO, strings.TrimSpace(a.Name), strings.TrimSpace(a.SNID)), nil
	}
	return fmt.Sprintf("Tareas realizadas el %s por %s", entryDateISO, strings.TrimSpace(a.Name)), nil
}
