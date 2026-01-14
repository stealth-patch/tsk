package styles

import "github.com/charmbracelet/lipgloss"

// Theme defines a color scheme
type Theme struct {
	Name         string
	Primary      lipgloss.Color
	PrimaryDark  lipgloss.Color
	PrimaryLight lipgloss.Color
	Accent       lipgloss.Color
	AccentDark   lipgloss.Color
}

// Available themes
var Themes = map[string]Theme{
	"purple": {
		Name:         "Purple",
		Primary:      lipgloss.Color("#8B5CF6"),
		PrimaryDark:  lipgloss.Color("#6D28D9"),
		PrimaryLight: lipgloss.Color("#A78BFA"),
		Accent:       lipgloss.Color("#F59E0B"),
		AccentDark:   lipgloss.Color("#D97706"),
	},
	"blue": {
		Name:         "Blue",
		Primary:      lipgloss.Color("#3B82F6"),
		PrimaryDark:  lipgloss.Color("#1D4ED8"),
		PrimaryLight: lipgloss.Color("#60A5FA"),
		Accent:       lipgloss.Color("#F59E0B"),
		AccentDark:   lipgloss.Color("#D97706"),
	},
	"green": {
		Name:         "Green",
		Primary:      lipgloss.Color("#10B981"),
		PrimaryDark:  lipgloss.Color("#059669"),
		PrimaryLight: lipgloss.Color("#34D399"),
		Accent:       lipgloss.Color("#F59E0B"),
		AccentDark:   lipgloss.Color("#D97706"),
	},
	"rose": {
		Name:         "Rose",
		Primary:      lipgloss.Color("#F43F5E"),
		PrimaryDark:  lipgloss.Color("#E11D48"),
		PrimaryLight: lipgloss.Color("#FB7185"),
		Accent:       lipgloss.Color("#8B5CF6"),
		AccentDark:   lipgloss.Color("#7C3AED"),
	},
	"orange": {
		Name:         "Orange",
		Primary:      lipgloss.Color("#F97316"),
		PrimaryDark:  lipgloss.Color("#EA580C"),
		PrimaryLight: lipgloss.Color("#FB923C"),
		Accent:       lipgloss.Color("#3B82F6"),
		AccentDark:   lipgloss.Color("#2563EB"),
	},
	"cyan": {
		Name:         "Cyan",
		Primary:      lipgloss.Color("#06B6D4"),
		PrimaryDark:  lipgloss.Color("#0891B2"),
		PrimaryLight: lipgloss.Color("#22D3EE"),
		Accent:       lipgloss.Color("#F43F5E"),
		AccentDark:   lipgloss.Color("#E11D48"),
	},
}

// ThemeNames returns sorted theme names for UI display
var ThemeNames = []string{"purple", "blue", "green", "rose", "orange", "cyan"}

// Current active theme colors
var (
	// Primary colors
	Primary      = lipgloss.Color("#8B5CF6")
	PrimaryDark  = lipgloss.Color("#6D28D9")
	PrimaryLight = lipgloss.Color("#A78BFA")

	// Secondary colors
	Secondary     = lipgloss.Color("#10B981")
	SecondaryDark = lipgloss.Color("#059669")

	// Accent colors
	Accent     = lipgloss.Color("#F59E0B")
	AccentDark = lipgloss.Color("#D97706")

	// Semantic colors
	Danger  = lipgloss.Color("#EF4444")
	Warning = lipgloss.Color("#F59E0B")
	Success = lipgloss.Color("#10B981")
	Info    = lipgloss.Color("#3B82F6")

	// Neutral colors
	Muted      = lipgloss.Color("#9CA3AF")
	MutedDark  = lipgloss.Color("#6B7280")
	Background = lipgloss.Color("#111827")
	Surface    = lipgloss.Color("#1F2937")
	Border     = lipgloss.Color("#374151")
	Foreground = lipgloss.Color("#F9FAFB")

	// Status colors
	StatusTodo  = lipgloss.Color("#9CA3AF")
	StatusDoing = lipgloss.Color("#3B82F6")
	StatusDone  = lipgloss.Color("#10B981")

	// Priority colors
	PriorityLow    = lipgloss.Color("#9CA3AF")
	PriorityMedium = lipgloss.Color("#F59E0B")
	PriorityHigh   = lipgloss.Color("#EF4444")
)

// ApplyTheme applies a theme by name
func ApplyTheme(name string) {
	theme, ok := Themes[name]
	if !ok {
		theme = Themes["purple"]
	}

	Primary = theme.Primary
	PrimaryDark = theme.PrimaryDark
	PrimaryLight = theme.PrimaryLight
	Accent = theme.Accent
	AccentDark = theme.AccentDark

	// Reinitialize styles with new colors
	reinitStyles()
}

// reinitStyles reinitializes all styles with current theme colors
func reinitStyles() {
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Background(Surface).
		Padding(0, 2).
		MarginBottom(1)

	AppTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground).
		Background(Primary).
		Padding(0, 2)

	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground).
		Background(Primary).
		Padding(0, 3)

	TaskItemSelected = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(Foreground).
		Background(Surface).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Primary)

	ColumnActive = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Primary).
		Padding(1, 1)

	DoneSectionHeaderActive = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		PaddingLeft(1)

	HelpKey = lipgloss.NewStyle().
		Foreground(PrimaryLight).
		Bold(true)

	InputPrompt = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	AccentStyle = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	DueToday = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	SearchBadge = lipgloss.NewStyle().
		Foreground(Foreground).
		Background(Accent).
		Padding(0, 1)

	Overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Background(Surface).
		Padding(1, 3)
}
