package styles

import "github.com/charmbracelet/lipgloss"

var (
	// App container - full screen with padding
	App = lipgloss.NewStyle().
		Padding(1, 3)

	// Header - big and bold
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Background(Surface).
		Padding(0, 2).
		MarginBottom(1)

	// App title - extra prominent
	AppTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground).
		Background(Primary).
		Padding(0, 2)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground)

	// Tab styles - more prominent
	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground).
		Background(Primary).
		Padding(0, 3)

	InactiveTab = lipgloss.NewStyle().
		Foreground(Muted).
		Background(Surface).
		Padding(0, 3)

	// Task list
	TaskItem = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingTop(0).
		PaddingBottom(0)

	TaskItemSelected = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(Foreground).
		Background(Surface).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Primary)

	TaskTitle = lipgloss.NewStyle().
		Foreground(Foreground)

	TaskTitleDone = lipgloss.NewStyle().
		Foreground(MutedDark).
		Strikethrough(true)

	TaskMeta = lipgloss.NewStyle().
		Foreground(Muted).
		MarginLeft(2)

	// Status icons - with backgrounds
	StatusTodoStyle = lipgloss.NewStyle().
		Foreground(StatusTodo)

	StatusDoingStyle = lipgloss.NewStyle().
		Foreground(StatusDoing).
		Bold(true)

	StatusDoneStyle = lipgloss.NewStyle().
		Foreground(StatusDone)

	// Priority styles
	PriorityLowStyle = lipgloss.NewStyle().
		Foreground(PriorityLow)

	PriorityMediumStyle = lipgloss.NewStyle().
		Foreground(PriorityMedium).
		Bold(true)

	PriorityHighStyle = lipgloss.NewStyle().
		Foreground(PriorityHigh).
		Bold(true)

	// Tag styles - pill shaped
	Tag = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E9D5FF")).
		Background(lipgloss.Color("#5B21B6")).
		Padding(0, 1).
		MarginRight(1)

	// Due date styles
	DueNormal = lipgloss.NewStyle().
		Foreground(Muted).
		Italic(true)

	DueToday = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	DueOverdue = lipgloss.NewStyle().
		Foreground(Danger).
		Bold(true).
		Blink(true)

	// Completed date style
	CompletedDate = lipgloss.NewStyle().
		Foreground(Muted).
		Italic(true)

	// Done section header
	DoneSectionHeader = lipgloss.NewStyle().
		Foreground(MutedDark).
		Bold(true).
		PaddingLeft(1)

	DoneSectionHeaderActive = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		PaddingLeft(1)

	// Kanban board - larger columns
	Column = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 1)

	ColumnActive = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Primary).
		Padding(1, 1)

	ColumnTitle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	ColumnTitleTodo = ColumnTitle.
		Foreground(Foreground).
		Background(lipgloss.Color("#374151"))

	ColumnTitleDoing = ColumnTitle.
		Foreground(Foreground).
		Background(Info)

	ColumnTitleDone = ColumnTitle.
		Foreground(Foreground).
		Background(Success)

	// Status bar
	StatusBar = lipgloss.NewStyle().
		Foreground(Muted).
		MarginTop(1)

	StatusBarError = lipgloss.NewStyle().
		Foreground(Foreground).
		Background(Danger).
		Padding(0, 2).
		MarginTop(1)

	// Help styles
	HelpKey = lipgloss.NewStyle().
		Foreground(PrimaryLight).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(Muted)

	// Input
	InputPrompt = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	// Empty state
	Empty = lipgloss.NewStyle().
		Foreground(Muted).
		Italic(true).
		Padding(2, 0).
		Align(lipgloss.Center)

	// Accent text style
	AccentStyle = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	// Muted text style
	MutedStyle = lipgloss.NewStyle().
		Foreground(Muted)

	// Project badge
	ProjectBadge = lipgloss.NewStyle().
		Foreground(Foreground).
		Background(MutedDark).
		Padding(0, 1)

	// Search indicator
	SearchBadge = lipgloss.NewStyle().
		Foreground(Foreground).
		Background(Accent).
		Padding(0, 1)

	// Overlay styles
	Overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Background(Surface).
		Padding(1, 3)

	OverlayDanger = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Danger).
		Background(Surface).
		Padding(1, 3)
)
