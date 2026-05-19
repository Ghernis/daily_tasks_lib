package cli

import (
	"errors"
	"fmt"
	"strings"

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
			p, err := store.AddProject(cmd.Context(), DB, args[0], description, subtasks)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE") {
					return fmt.Errorf("project name already exists: %q", args[0])
				}
				return err
			}
			fmt.Printf("Created project #%d %q.\n", p.ID, p.Name)
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
		Short: "List projects (full descriptions)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := store.ListProjects(cmd.Context(), DB)
			if err != nil {
				return err
			}
			for _, p := range rows {
				fmt.Printf("#%-4d  %s\n", p.ID, p.Name)
				desc := strings.TrimSpace(p.Description)
				if desc != "" {
					fmt.Printf("       %s\n", desc)
				} else {
					fmt.Println("       (no description)")
				}
				fmt.Printf("       created %s\n\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			if len(rows) == 0 {
				fmt.Println("No projects yet.")
			}
			return nil
		},
	}
}

func projectSetDescriptionCmd() *cobra.Command {
	var projectID int64
	var description string
	cmd := &cobra.Command{
		Use:   "set-description [NAME]",
		Short: "Set project description (use -p ID or NAME)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				if projectID > 0 && description == "" {
					description = args[0]
				} else {
					name = args[0]
					if len(args) > 1 && description == "" {
						description = args[1]
					}
				}
			}
			if description == "" {
				return fmt.Errorf("description is required (-d or as argument)")
			}
			if projectID <= 0 && name == "" {
				return fmt.Errorf("pass -p/--project with id, or the project NAME")
			}
			p, err := resolveProject(cmd.Context(), DB, projectID, name)
			if err != nil {
				return err
			}
			if err := store.SetProjectDescription(cmd.Context(), DB, p.ID, "", description); err != nil {
				return err
			}
			fmt.Printf("Updated description for #%d %q.\n", p.ID, p.Name)
			return nil
		},
	}
	cmd.Flags().Int64VarP(&projectID, "project", "p", 0, "Project id (from project list)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Fixed task description (or pass as argument with -p)")
	return cmd
}

func projectSetDefaultSubtasksCmd() *cobra.Command {
	var projectID int64
	var subtasks []string
	cmd := &cobra.Command{
		Use:   "set-default-subtasks [NAME]",
		Short: "Replace default subtasks (use -p ID or NAME)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if projectID <= 0 && name == "" {
				return fmt.Errorf("pass -p/--project with id, or the project NAME")
			}
			p, err := resolveProject(cmd.Context(), DB, projectID, name)
			if err != nil {
				return err
			}
			if err := store.SetDefaultSubtasks(cmd.Context(), DB, p.ID, "", subtasks); err != nil {
				if errors.Is(err, store.ErrProjectNotFound) {
					return fmt.Errorf("unknown project id: %d", projectID)
				}
				return err
			}
			fmt.Printf("Updated default subtasks for #%d %q.\n", p.ID, p.Name)
			return nil
		},
	}
	cmd.Flags().Int64VarP(&projectID, "project", "p", 0, "Project id (from project list)")
	cmd.Flags().StringArrayVarP(&subtasks, "subtask", "s", nil, "Default subtask (repeatable)")
	return cmd
}
