package cli

import (
	"fmt"

	"daily-tasks/go_version/internal/store"

	"github.com/spf13/cobra"
)

func newTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Store identifiers on daily tickets after creation",
	}
	cmd.AddCommand(ticketSetRITMCmd())
	return cmd
}

func ticketSetRITMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-ritm DATE RITM",
		Short: "Store RITM number for a ticket day",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := store.SetRITMForDate(cmd.Context(), DB, args[0], args[1])
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("no ticket found for %s; run `log` for that day first", args[0])
			}
			fmt.Printf("Stored RITM %q on %s.\n", args[1], args[0])
			return nil
		},
	}
}
