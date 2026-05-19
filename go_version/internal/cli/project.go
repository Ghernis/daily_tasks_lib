package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"daily-tasks/go_version/internal/store"

	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create and list projects",
	}
	cmd.AddCommand(projectAddCmd())
	cmd.AddCommand(projectListCmd())
	cmd.AddCommand(projectSetDescriptionCmd())
	cmd.AddCommand(projectSetDefaultSubtasksCmd())
	return cmd
}

func projectAddCmd() *cobra.Command {
	var description string
	var subtasks []string
	cmd := &cobra.Command{
		Use:   "add NAME",
		Short: "Add a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := store.AddProject(cmd.Context(), DB, args[0], description, subtasks)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE") {
					return fmt.Errorf("project name already exists: %q", args[0])
				}
				return err
			}
			fmt.Printf("Created project %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Fixed task description for this project")
	cmd.Flags().StringArrayVarP(&subtasks, "subtask", "s", nil, "Default subtask (repeatable)")
	return cmd
}

func projectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := store.ListProjects(cmd.Context(), DB)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tCREATED_AT")
			for _, p := range rows {
				desc := strings.TrimSpace(p.Description)
				if len(desc) > 48 {
					desc = desc[:45] + "..."
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", p.ID, p.Name, desc, p.CreatedAt.Format("2006-01-02T15:04:05"))
			}
			return w.Flush()
		},
	}
}

func projectSetDescriptionCmd() *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "set-description NAME",
		Short: "Set project description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := store.SetProjectDescription(cmd.Context(), DB, args[0], description); err != nil {
				if errors.Is(err, store.ErrProjectNotFound) {
					return fmt.Errorf("unknown project: %q", args[0])
				}
				return err
			}
			fmt.Printf("Updated description for %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Fixed task description")
	_ = cmd.MarkFlagRequired("description")
	return cmd
}

func projectSetDefaultSubtasksCmd() *cobra.Command {
	var subtasks []string
	cmd := &cobra.Command{
		Use:   "set-default-subtasks NAME",
		Short: "Replace default subtasks for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := store.SetDefaultSubtasks(cmd.Context(), DB, args[0], subtasks); err != nil {
				if errors.Is(err, store.ErrProjectNotFound) {
					return fmt.Errorf("unknown project: %q", args[0])
				}
				return err
			}
			fmt.Printf("Updated default subtasks for %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&subtasks, "subtask", "s", nil, "Default subtask (repeatable)")
	return cmd
}
