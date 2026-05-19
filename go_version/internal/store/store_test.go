package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
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

func TestAddProjectWithDefaultSubtasks(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, err := store.AddProject(ctx, conn, "Alpha", "", []string{"t1", "t2"})
	if err != nil {
		t.Fatal(err)
	}
	titles, err := store.ListDefaultSubtaskTitles(ctx, conn, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 2 || titles[0] != "t1" || titles[1] != "t2" {
		t.Fatalf("got %v", titles)
	}
}

func TestPastSubtaskNames(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, _ := store.AddProject(ctx, conn, "PastProj", "", nil)
	pid := p.ID
	if err := store.SaveDailyTicket(ctx, conn, "2026-05-01", "a", "", []store.TaskInput{
		{ProjectID: pid, Hours: 1, SubtaskNames: []string{"alpha", "beta"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDailyTicket(ctx, conn, "2026-05-02", "b", "", []store.TaskInput{
		{ProjectID: pid, Hours: 0.5, SubtaskNames: []string{"gamma"}},
	}); err != nil {
		t.Fatal(err)
	}
	beforeD1, _ := store.ListPastSubtaskNames(ctx, conn, pid, "2026-05-01")
	beforeD2, _ := store.ListPastSubtaskNames(ctx, conn, pid, "2026-05-02")
	beforeD3, _ := store.ListPastSubtaskNames(ctx, conn, pid, "2026-05-05")
	if len(beforeD1) != 0 {
		t.Fatalf("d1: %v", beforeD1)
	}
	if len(beforeD2) != 2 {
		t.Fatalf("d2: %v", beforeD2)
	}
	if len(beforeD3) != 3 || beforeD3[0] != "gamma" {
		t.Fatalf("d3: %v", beforeD3)
	}
}

func TestExportAuthorAndDescription(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	_ = store.SetAuthorProfile(ctx, conn, "Hernan Gomez", "hernan@example.com", "SE40634")
	p, _ := store.AddProject(ctx, conn, "Beta", "Standing project scope", nil)
	_ = store.SaveDailyTicket(ctx, conn, "2026-05-10", "Day", "ticket body", []store.TaskInput{
		{ProjectID: p.ID, Hours: 1.5, TaskDescription: "Standing project scope", SubtaskNames: []string{"x"}},
	})
	doc, err := export.Build(ctx, conn, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(doc)
	var payload map[string]interface{}
	_ = json.Unmarshal(data, &payload)
	tickets := payload["tickets"].([]interface{})
	td := tickets[0].(map[string]interface{})
	authors := td["author"].([]interface{})
	if len(authors) != 1 {
		t.Fatalf("authors: %v", authors)
	}
	tasks := td["tasks"].([]interface{})
	task := tasks[0].(map[string]interface{})
	if task["description"] != "Standing project scope" {
		t.Fatalf("desc: %v", task["description"])
	}
}

func TestRITMPreservedOnReLog(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, _ := store.AddProject(ctx, conn, "Gamma", "", nil)
	d := "2026-05-02"
	_ = store.SaveDailyTicket(ctx, conn, d, "T", "", []store.TaskInput{
		{ProjectID: p.ID, Hours: 1, SubtaskNames: []string{"a"}},
	})
	_, _ = store.SetRITMForDate(ctx, conn, d, "RITM001")
	_ = store.SaveDailyTicket(ctx, conn, d, "T2", "", []store.TaskInput{
		{ProjectID: p.ID, Hours: 2, SubtaskNames: []string{"b", "c"}},
	})
	ticket, _ := store.GetTicketByDate(ctx, conn, d)
	if !ticket.RITMNumber.Valid || ticket.RITMNumber.String != "RITM001" {
		t.Fatalf("ritm lost: %+v", ticket.RITMNumber)
	}
	if ticket.Title != "T2" {
		t.Fatalf("title: %s", ticket.Title)
	}
}

func TestMergeSubtaskRows(t *testing.T) {
	ordered, def := store.MergeSubtaskRows([]string{"a", "b"}, []string{"c", "a"})
	if len(ordered) != 3 || !def["a"] || !def["b"] || def["c"] {
		t.Fatalf("%v %v", ordered, def)
	}
}

func TestSaveDailyTicketWithoutCreatedAtDefault(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	_, err := conn.ExecContext(ctx, `DROP TABLE IF EXISTS ticket_daily_task_subtask`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.ExecContext(ctx, `DROP TABLE IF EXISTS ticket_daily_task`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.ExecContext(ctx, `DROP TABLE IF EXISTS ticket_daily`)
	if err != nil {
		t.Fatal(err)
	}
	// Schema like SQLModel/Python: NOT NULL created_at without SQL DEFAULT.
	_, err = conn.ExecContext(ctx, `
		CREATE TABLE ticket_daily (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			"date" DATE NOT NULL UNIQUE,
			title VARCHAR(500) NOT NULL,
			description VARCHAR(4000) NOT NULL DEFAULT '',
			ritm_number VARCHAR(128),
			created_at DATETIME NOT NULL
		)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDailyTicket(ctx, conn, "2026-05-20", "T", "D", nil); err != nil {
		t.Fatalf("save failed: %v", err)
	}
}

func TestResolveProjectByID(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, _ := store.AddProject(ctx, conn, "ByID", "hello", nil)
	got, err := store.ResolveProject(ctx, conn, p.ID, "")
	if err != nil || got.Name != "ByID" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestSetProjectDescriptionByID(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	p, _ := store.AddProject(ctx, conn, "Desc", "", nil)
	if err := store.SetProjectDescription(ctx, conn, p.ID, "", "updated"); err != nil {
		t.Fatal(err)
	}
	got, _ := store.GetProjectByID(ctx, conn, p.ID)
	if got.Description != "updated" {
		t.Fatalf("got %q", got.Description)
	}
}

func TestParseHours(t *testing.T) {
	v, err := store.ParseHours("1,5")
	if err != nil || v != 1.5 {
		t.Fatalf("%v %v", v, err)
	}
}
