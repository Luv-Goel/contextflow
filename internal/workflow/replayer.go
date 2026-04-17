package workflow

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// ReplayMode controls how a workflow is replayed.
type ReplayMode int

const (
	// Interactive asks before each command.
	Interactive ReplayMode = iota
	// DryRun prints commands without executing.
	DryRun
)

// Replay steps through a workflow's commands.
func Replay(w db.Workflow, mode ReplayMode) error {
	if len(w.Commands) == 0 {
		return fmt.Errorf("workflow %q has no commands", w.Name)
	}

	name := w.Name
	if name == "" {
		name = fmt.Sprintf("workflow #%d", w.ID)
	}

	fmt.Printf("\n🔁  Replaying \033[1m%s\033[0m (%d commands)\n\n", name, len(w.Commands))

	reader := bufio.NewReader(os.Stdin)

	for i, cmd := range w.Commands {
		fmt.Printf("  \033[2m[%d/%d]\033[0m \033[36m%s\033[0m\n", i+1, len(w.Commands), cmd.Command)

		if mode == DryRun {
			continue
		}

		fmt.Printf("       Run? [y/N/q] ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))

		switch line {
		case "y", "yes":
			if err := runCommand(cmd.Command, cmd.Directory); err != nil {
				fmt.Printf("       \033[31m✗ exit: %v\033[0m\n", err)
			} else {
				fmt.Printf("       \033[32m✓ done\033[0m\n")
			}
		case "q", "quit":
			fmt.Println("\n  Stopped.")
			return nil
		default:
			fmt.Println("       Skipped.")
		}
	}

	fmt.Println("\n✅  Replay complete.")
	return nil
}

func runCommand(command, dir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
