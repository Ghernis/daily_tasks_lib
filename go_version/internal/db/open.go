package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// DefaultPath returns the same default DB path as the Python CLI.
func DefaultPath() string {
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "daily_tasks", "data.db")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "daily_tasks", "data.db")
	}
	return filepath.Join(home, ".daily_tasks", "data.db")
}

// Open opens SQLite at path, runs migrations, and returns *sql.DB.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", filepath.ToSlash(path))
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		conn.Close()
		return nil, err
	}
	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func migrate(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL UNIQUE,
			description VARCHAR(4000) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS author_profile (
			id INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL DEFAULT '',
			sn_id VARCHAR(64) NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS project_default_subtask (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL REFERENCES project(id),
			title VARCHAR(500) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_daily (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			"date" DATE NOT NULL UNIQUE,
			title VARCHAR(500) NOT NULL,
			description VARCHAR(4000) NOT NULL DEFAULT '',
			ritm_number VARCHAR(128),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_daily_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_daily_id INTEGER NOT NULL REFERENCES ticket_daily(id),
			project_id INTEGER NOT NULL REFERENCES project(id),
			hours REAL NOT NULL,
			task_description VARCHAR(4000) NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(ticket_daily_id, project_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_daily_task_subtask (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_daily_task_id INTEGER NOT NULL REFERENCES ticket_daily_task(id),
			name VARCHAR(500) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(s); err != nil {
			return err
		}
	}
	return ensureProjectDescriptionColumn(conn)
}

func ensureProjectDescriptionColumn(conn *sql.DB) error {
	rows, err := conn.Query(`PRAGMA table_info(project)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	hasDesc := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "description" {
			hasDesc = true
			break
		}
	}
	if hasDesc {
		return rows.Err()
	}
	_, err = conn.Exec(`ALTER TABLE project ADD COLUMN description VARCHAR(4000) NOT NULL DEFAULT ''`)
	return err
}
