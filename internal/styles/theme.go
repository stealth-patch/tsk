package styles

import "github.com/charmbracelet/lipgloss"

// Colors - More vibrant palette
var (
	// Primary colors
	Primary      = lipgloss.Color("#8B5CF6") // Vivid Purple
	PrimaryDark  = lipgloss.Color("#6D28D9") // Darker purple
	PrimaryLight = lipgloss.Color("#A78BFA") // Lighter purple

	// Secondary colors
	Secondary     = lipgloss.Color("#10B981") // Emerald
	SecondaryDark = lipgloss.Color("#059669")

	// Accent colors
	Accent     = lipgloss.Color("#F59E0B") // Amber
	AccentDark = lipgloss.Color("#D97706")

	// Semantic colors
	Danger  = lipgloss.Color("#EF4444") // Red
	Warning = lipgloss.Color("#F59E0B") // Amber
	Success = lipgloss.Color("#10B981") // Green
	Info    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	Muted      = lipgloss.Color("#9CA3AF") // Gray-400
	MutedDark  = lipgloss.Color("#6B7280") // Gray-500
	Background = lipgloss.Color("#111827") // Gray-900
	Surface    = lipgloss.Color("#1F2937") // Gray-800
	Border     = lipgloss.Color("#374151") // Gray-700
	Foreground = lipgloss.Color("#F9FAFB") // Gray-50

	// Status colors
	StatusTodo  = lipgloss.Color("#9CA3AF") // Gray
	StatusDoing = lipgloss.Color("#3B82F6") // Blue
	StatusDone  = lipgloss.Color("#10B981") // Green

	// Priority colors
	PriorityLow    = lipgloss.Color("#9CA3AF") // Gray
	PriorityMedium = lipgloss.Color("#F59E0B") // Amber
	PriorityHigh   = lipgloss.Color("#EF4444") // Red
)
