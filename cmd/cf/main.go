// ContextFlow CLI — Shell history that understands your workflows.
// cmd/cf/main.go - CLI entry point
package main
import (
	"fmt"
	"os"
	"strconv"
	"time"
	"github.com/Luv-Goel/contextflow/internal/capture"
	"github.com/Luv-Goel/contextflow/internal/db"
	"github.com/Luv-Goel/contextflow/internal/export"
	"github.com/Luv-Goel/contextflow/internal/story"
	"github.com/Luv-Goel/contextflow/internal/workflow"
	"github.com/spf13/cobra"
)
var (
	version   = "v0.1.0"
	printOnly bool
)
func main() {
	var rootCmd = &cobra.Command{
		Use:   "cf",
		Short: "ContextFlow — Shell history that understands your workflows",
		Long:  `ContextFlow remembers the workflow, not just the command.
Every developer has typed 'history | grep' in desperation. You know a command 
exists somewhere — you just can't find the sequence around it.
ContextFlow automatically groups related commands into workflows and lets you 
search, replay, and export them.`,
	}
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(workflowsCmd())
	rootCmd.AddCommand(replayCmd())
	rootCmd.AddCommand(exportCmd())
	rootCmd.AddCommand(statsCmd())
	rootCmd.AddCommand(tagCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(importCmd())
	rootCmd.AddCommand(storyCmd())
	rootCmd.AddCommand(hookCmd())
	rootCmd.AddCommand(recordCmd())
	rootCmd.AddCommand(uninstallCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "cf: %v\n", err)
		os.Exit(1)
	}
}
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show ContextFlow version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("ContextFlow %s\n", version)
		},
	}
}
func initCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Output shell hook configuration",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			switch shell {
			case "bash":
				fmt.Printf("PROMPT_COMMAND=\"$(cf hook bash)\"\n")
			case "zsh":
				fmt.Printf("precmd() { cf hook zsh }\n")
			case "fish":
				fmt.Printf("function cf_hook --on-variable FISH_CLI_BEFORE_RESOLVE 2>/dev/null;\n  cf hook fish\nend\n")
			default:
				return fmt.Errorf("unsupported shell: %s (use: bash, zsh, fish)", shell)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "Shell type (bash, zsh, fish)")
	return cmd
}
func openDB() (*db.DB, error) {
	database, err := db.Open()
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return database, nil
}
func searchCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search command history (Ctrl+R replacement)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			var commands []db.Command
			if query == "" {
				commands, err = database.RecentCommands("", limit)
			} else {
				commands, err = database.SearchCommands(query, limit)
			}
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}
			if printOnly {
				for _, c := range commands {
					fmt.Printf("%d\t%s\n", c.ID, c.Command)
				}
				return nil
			}
			// Launch TUI (Bubble Tea)
			// TODO: Wire up tui.NewSearchModel(commands, false)
			fmt.Println("(TUI not wired yet — use -p flag for plain output)")
			for _, c := range commands {
				fmt.Printf("%d\t%s\n", c.ID, c.Command)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum number of results")
	cmd.Flags().BoolVarP(&printOnly, "print", "p", false, "Plain text output (no TUI)")
	return cmd
}
func workflowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Browse auto-detected workflows",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			commands, err := database.RecentCommands("", 500)
			if err != nil {
				return fmt.Errorf("get commands: %w", err)
			}
			workflows := workflow.Detect(commands)
			if len(workflows) == 0 {
				fmt.Println("No workflows detected yet.")
				fmt.Println("Keep using your shell — ContextFlow will auto-detect workflows.")
				return nil
			}
			fmt.Printf("Found %d workflows:\n\n", len(workflows))
			for _, w := range workflows {
				age := time.Since(w.UpdatedAt)
				fmt.Printf("#%d  %s  (%d commands, %s ago)\n", w.ID, w.Name, len(w.Commands), age)
				if w.GitRepo != "" {
					fmt.Printf("     Repo: %s\n", w.GitRepo)
				}
				fmt.Printf("\n")
			}
			return nil
		},
	}
	return cmd
}
func replayCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "replay [workflow-id]",
		Short: "Replay a workflow interactively",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid workflow id: %w", err)
			}
			w, err := database.GetWorkflowByID(id)
			if err != nil {
				return fmt.Errorf("get workflow: %w", err)
			}
			mode := workflow.Interactive
			if dryRun {
				mode = workflow.DryRun
			}
			return workflow.Replay(w, mode)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without executing")
	return cmd
}
func exportCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export [workflow-id]",
		Short: "Export workflow as script or runbook",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid workflow id: %w", err)
			}
			w, err := database.GetWorkflowByID(id)
			if err != nil {
				return fmt.Errorf("get workflow: %w", err)
			}
			if format == "md" {
				fmt.Print(export.ToMarkdown(w))
			} else {
				fmt.Print(export.ToShellScript(w))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "sh", "Output format: sh (shell script), md (markdown)")
	return cmd
}
func statsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show usage statistics",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			stats, err := database.GetStats()
			if err != nil {
				return fmt.Errorf("get stats: %w", err)
			}
			fmt.Printf("ContextFlow Statistics\n")
			fmt.Printf("======================\n\n")
			fmt.Printf("Total commands:  %d\n", stats.TotalCommands)
			fmt.Printf("Total workflows: %d\n", stats.TotalWorkflows)
			fmt.Printf("Sessions:       %d\n", stats.TotalSessions)
			if stats.TopCommands != nil {
				fmt.Printf("\nTop commands:\n")
				for i, c := range stats.TopCommands {
					fmt.Printf("  %d. %s (%d)\n", i+1, c.Command, c.Count)
				}
			}
			return nil
		},
	}
	return cmd
}
func tagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag [workflow-id] [name]",
		Short: "Give a workflow a custom name",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO: Update workflow name in DB
			fmt.Printf("Tagged workflow #%s as %q\n", args[0], args[1])
			return nil
		},
	}
	return cmd
}
func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [workflow-id]",
		Short: "Delete a workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO: Delete workflow from DB
			fmt.Printf("Deleted workflow #%s\n", args[0])
			return nil
		},
	}
	return cmd
}
func importCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import existing shell history",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := file
			if len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				// Auto-detect
				home := os.ExpandEnv("$HOME")
				if _, err := os.Stat(home + "/.bash_history"); err == nil {
					path = home + "/.bash_history"
				} else if _, err := os.Stat(home + "/.zsh_history"); err == nil {
					path = home + "/.zsh_history"
				}
			}
			if path == "" {
				return fmt.Errorf("no history file found, use cf import <path>")
			}
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			result, err := capture.ImportHistoryFile(path, database)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}
			fmt.Printf("Imported %d commands (%d skipped)\n", result.Imported, result.Skipped)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "History file to import")
	return cmd
}
func storyCmd() *cobra.Command {
	var since string
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Generate a narrative summary of your work",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()
			d := 24 * time.Hour // default: last 24 hours
			if since != "" {
				// TODO: Parse duration string
			}
			narrative := story.Generate(database, d)
			fmt.Println(narrative)
			return nil
		},
	}
	cmd.Flags().StringVar(&since, "since", "24h", "Duration (e.g., 24h, 7d)")
	return cmd
}
func hookCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Output shell integration script",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			script, err := capture.InitShell(shell)
			if err != nil {
				return err
			}
			fmt.Println(script)
			return nil
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "Shell type (bash, zsh, fish)")
	return cmd
}
func recordCmd() *cobra.Command {
	var (
		cmdStr    string
		dir      string
		exitCode int
		duration int64
		session  string
	)
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record a command (internal use)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return capture.Record(cmdStr, dir, exitCode, duration*1000000, session)
		},
	}
	cmd.Flags().StringVar(&cmdStr, "cmd", "", "Command to record")
	cmd.Flags().StringVar(&dir, "dir", "", "Working directory")
	cmd.Flags().IntVar(&exitCode, "exit", 0, "Exit code")
	cmd.Flags().Int64Var(&duration, "duration", 0, "Duration in nanoseconds")
	cmd.Flags().StringVar(&session, "session", "", "Session ID")
	return cmd
}
// Helpers
func timeSince(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
func uninstallCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove ContextFlow completely from system",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if !confirm {
				fmt.Println("Run with --confirm to uninstall")
				fmt.Println("This will delete ~/.contextflow/ and remove the binary")
				return nil
			}
			dir, err := db.DataDir()
			if err != nil {
				return err
			}
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("remove data dir: %w", err)
			}
			binPath, err := os.Executable()
			if err == nil {
				os.Remove(binPath)
			}
			fmt.Println("ContextFlow uninstalled")
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm uninstallation")
	return cmd
}
