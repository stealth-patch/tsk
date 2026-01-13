package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/hwanchang/tsk/internal/app"
	"github.com/hwanchang/tsk/internal/db"
	"github.com/hwanchang/tsk/internal/store"
)

var (
	dbPath string
	st     *store.SQLiteStore
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tsk",
		Short: "A terminal task manager",
		Long:  `tsk is a terminal-based task manager with both TUI and CLI interfaces.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip db init for help commands
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}
			return initStore()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if st != nil {
				st.Close()
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: run TUI
			return runTUI()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database file path (default: ~/.local/share/tsk/tsk.db)")

	// Add subcommands
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newDoneCmd())
	rootCmd.AddCommand(newDoingCmd())
	rootCmd.AddCommand(newRmCmd())
	rootCmd.AddCommand(newProjectCmd())
	rootCmd.AddCommand(newTagCmd())
	rootCmd.AddCommand(newRecurrenceCmd())

	return rootCmd
}

func initStore() error {
	path := dbPath
	if path == "" {
		path = db.GetDBPath()
	}

	database, err := db.New(path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := database.Migrate(); err != nil {
		database.Close()
		return fmt.Errorf("migrate database: %w", err)
	}

	st = store.New(database)
	return nil
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI() error {
	m := app.New(st)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}
