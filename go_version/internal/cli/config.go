package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"daily-tasks/go_version/internal/store"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Global settings (author profile for exports)",
	}
	cmd.AddCommand(configSetAuthorCmd())
	cmd.AddCommand(configShowAuthorCmd())
	return cmd
}

func configSetAuthorCmd() *cobra.Command {
	var name, email, snID string
	cmd := &cobra.Command{
		Use:   "set-author",
		Short: "Set author profile for ticket export",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := store.SetAuthorProfile(cmd.Context(), DB, name, email, snID); err != nil {
				return err
			}
			fmt.Println("Author profile saved.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Display name on tickets")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	cmd.Flags().StringVar(&snID, "sn-id", "", "ServiceNow / corporate ID")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func configShowAuthorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-author",
		Short: "Show author profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := store.GetAuthorProfile(cmd.Context(), DB)
			if err != nil {
				return err
			}
			if a == nil {
				return fmt.Errorf("no author configured; run: daily-tasks config set-author --name \"...\"")
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "FIELD\tVALUE")
			fmt.Fprintf(w, "name\t%s\n", a.Name)
			fmt.Fprintf(w, "email\t%s\n", a.Email)
			fmt.Fprintf(w, "sn_id\t%s\n", a.SNID)
			return w.Flush()
		},
	}
}
