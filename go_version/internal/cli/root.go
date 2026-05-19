package cli

import (
	"database/sql"
	"os"

	"daily-tasks/go_version/internal/db"

	"github.com/spf13/cobra"
)

var dbPath string
var DB *sql.DB

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "daily-tasks",
		Short: "Log daily time per project with subtasks and JSON export",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			path := dbPath
			if path == "" {
				path = db.DefaultPath()
			}
			conn, err := db.Open(path)
			if err != nil {
				return err
			}
			DB = conn
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if DB != nil {
				DB.Close()
				DB = nil
			}
		},
	}
	root.PersistentFlags().StringVar(&dbPath, "db", "", "SQLite database file (default: same as Python CLI)")
	root.AddCommand(newLogCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newProjectCmd())
	root.AddCommand(newTicketCmd())
	root.AddCommand(newConfigCmd())
	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
