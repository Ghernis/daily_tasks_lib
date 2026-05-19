package cli

import (
	"fmt"
	"time"

	"daily-tasks/go_version/internal/tui/logwizard"

	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	var entryDate string
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Interactive wizard to log daily work",
		RunE: func(cmd *cobra.Command, args []string) error {
			logDate := entryDate
			if logDate == "" {
				logDate = time.Now().Format("2006-01-02")
			}
			if _, err := time.Parse("2006-01-02", logDate); err != nil {
				return fmt.Errorf("--date must be YYYY-MM-DD: %w", err)
			}
			if err := logwizard.Run(DB, logDate); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&entryDate, "date", "", "Log date as YYYY-MM-DD (default: today)")
	return cmd
}
