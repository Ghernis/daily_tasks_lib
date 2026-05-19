package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"daily-tasks/go_version/internal/export"

	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var outputPath string
	var exportAll bool
	var dateFlag string
	var fromDate string
	var toDate string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export tickets to JSON (default: today only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			today := time.Now().Format("2006-01-02")

			var fromPtr, toPtr *string
			labelDate := today

			switch {
			case exportAll:
				// no date filter
			case dateFlag != "":
				fromPtr = &dateFlag
				toPtr = &dateFlag
				labelDate = dateFlag
			case fromDate != "" || toDate != "":
				from := fromDate
				to := toDate
				if from == "" {
					from = to
				}
				if to == "" {
					to = from
				}
				fromPtr = &from
				toPtr = &to
				if from == to {
					labelDate = from
				} else {
					labelDate = fmt.Sprintf("%s_to_%s", from, to)
				}
			default:
				fromPtr = &today
				toPtr = &today
				labelDate = today
			}

			doc, err := export.Build(cmd.Context(), DB, fromPtr, toPtr)
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = fmt.Sprintf("export-%s.json", labelDate)
			}
			if err := os.WriteFile(outputPath, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("Wrote %s (%d ticket(s))\n", outputPath, len(doc.Tickets))
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output JSON file (default: export-<date>.json)")
	cmd.Flags().BoolVar(&exportAll, "all", false, "Export all tickets (no date filter)")
	cmd.Flags().StringVar(&dateFlag, "date", "", "Export a single day (YYYY-MM-DD); default is today")
	cmd.Flags().StringVar(&fromDate, "from", "", "Range start date (YYYY-MM-DD); with --to only, or alone for that single day")
	cmd.Flags().StringVar(&toDate, "to", "", "Range end date (YYYY-MM-DD)")
	return cmd
}
