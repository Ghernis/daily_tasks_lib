package export_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"daily-tasks/go_version/internal/db"
	"daily-tasks/go_version/internal/export"
	"daily-tasks/go_version/internal/store"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestBuildSingleDayFilter(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, _ := store.AddProject(ctx, conn, "P", "", nil)
	_ = store.SaveDailyTicket(ctx, conn, "2026-05-18", "d18", "", []store.TaskInput{{ProjectID: p.ID, Hours: 1}})
	_ = store.SaveDailyTicket(ctx, conn, "2026-05-19", "d19", "", []store.TaskInput{{ProjectID: p.ID, Hours: 2}})

	from := "2026-05-18"
	to := "2026-05-18"
	doc, err := export.Build(ctx, conn, &from, &to)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Tickets) != 1 {
		t.Fatalf("want 1 ticket, got %d", len(doc.Tickets))
	}
	if doc.Tickets[0].Date != "2026-05-18" {
		t.Fatalf("date: %s", doc.Tickets[0].Date)
	}
}
