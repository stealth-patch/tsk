package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hwanchang/tsk/internal/config"
	"github.com/hwanchang/tsk/internal/model"
	"github.com/hwanchang/tsk/internal/store"
	"github.com/hwanchang/tsk/internal/styles"
)

type ViewType int

const (
	ViewList ViewType = iota
	ViewBoard
)

type InputMode int

const (
	InputNone InputMode = iota
	InputAdd
	InputSearch
	InputEdit
	InputAddProject
	InputAddTag
	InputDueDate
)

type OverlayMode int

const (
	OverlayNone OverlayMode = iota
	OverlayHelp
	OverlayProjectSelect
	OverlayProjectCreate
	OverlayConfirmDelete
	OverlayDueDate
	OverlayDueDateCustom
	OverlayTagSelect
	OverlayTagCreate
	OverlayTaskDetail
	OverlayConfirmDeleteTag
	OverlayConfirmDeleteProject
	OverlayRecurrenceSelect
	OverlayThemeSelect
)

type Model struct {
	store *store.SQLiteStore

	// Data
	tasks    []model.Task
	projects []model.Project
	tags     []model.Tag

	// UI state
	activeView    ViewType
	cursor        int
	boardCol      int
	boardCursors  [3]int
	boardScrolls  [3]int // scroll position for each column
	width, height int

	// Overlay
	overlayMode   OverlayMode
	overlayCursor int

	// Input
	inputMode   InputMode
	textInput   textinput.Model
	inputPrompt string
	editTaskID  int64

	// Filter
	currentProject     *int64
	currentProjectName string
	searchQuery        string

	// Done section (list view)
	doneCollapsed bool // true = collapsed, false = expanded
	doneCursor    int  // cursor position in done section
	inDoneSection bool // true if cursor is in done section
	activeTasks   []model.Task
	doneTasksList []model.Task

	// Stats
	totalTasks    int
	doneTaskCount int

	// Project create form
	projectFormName  string
	projectFormDesc  string
	projectFormFocus int // 0: name, 1: desc

	// Tag create form
	tagFormName string

	// Due date custom form
	dueDateFormValue string

	// Status
	statusText  string
	statusError bool

	// Ready flag
	ready bool
}

func New(st *store.SQLiteStore) *Model {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 500
	ti.Width = 80

	return &Model{
		store:              st,
		textInput:          ti,
		currentProjectName: "All",
		doneCollapsed:      false, // done section expanded by default
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadTasks(m.store, store.TaskFilter{}),
		loadProjects(m.store),
		loadTags(m.store),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		// Handle overlay mode first
		if m.overlayMode != OverlayNone {
			return m.handleOverlay(msg)
		}

		// Handle input mode
		if m.inputMode != InputNone {
			return m.handleInputMode(msg)
		}

		// Global keys
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, Keys.Help):
			m.overlayMode = OverlayHelp
			return m, nil

		case key.Matches(msg, Keys.ToggleView):
			if m.activeView == ViewList {
				m.activeView = ViewBoard
			} else {
				m.activeView = ViewList
			}
			return m, m.reloadTasks()

		case key.Matches(msg, Keys.Add):
			m.inputMode = InputAdd
			m.inputPrompt = "New task: "
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Enter task title..."
			m.textInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, Keys.Edit):
			if task := m.selectedTask(); task != nil {
				m.inputMode = InputEdit
				m.inputPrompt = "Edit: "
				m.textInput.SetValue(task.Title)
				m.textInput.Placeholder = "Enter new title..."
				m.editTaskID = task.ID
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, Keys.Search):
			m.inputMode = InputSearch
			m.inputPrompt = "Search: "
			m.textInput.SetValue(m.searchQuery)
			m.textInput.Placeholder = "Search tasks..."
			m.textInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, Keys.Project):
			m.overlayMode = OverlayProjectSelect
			m.overlayCursor = 0
			return m, nil

		case key.Matches(msg, Keys.Theme):
			m.overlayMode = OverlayThemeSelect
			// Find current theme index
			currentTheme := config.GetTheme()
			for i, name := range styles.ThemeNames {
				if name == currentTheme {
					m.overlayCursor = i
					break
				}
			}
			return m, nil

		case msg.String() == "A":
			// Toggle done section collapse (list view only)
			if m.activeView == ViewList {
				m.doneCollapsed = !m.doneCollapsed
				if m.doneCollapsed && m.inDoneSection {
					// Move cursor back to active section
					m.inDoneSection = false
					if len(m.activeTasks) > 0 {
						m.cursor = len(m.activeTasks) - 1
					} else {
						m.cursor = 0
					}
				}
			}
			return m, nil

		case msg.String() == "c":
			// Clear search
			if m.searchQuery != "" {
				m.searchQuery = ""
				return m, m.reloadTasks()
			}

		case msg.String() == "d":
			// Set due date
			if m.selectedTask() != nil {
				m.overlayMode = OverlayDueDate
				m.overlayCursor = 0
				return m, nil
			}

		case msg.String() == "t":
			// Set tags
			if m.selectedTask() != nil {
				m.overlayMode = OverlayTagSelect
				m.overlayCursor = 0
				return m, nil
			}

		case msg.String() == "r":
			// Set/remove recurrence
			if m.selectedTask() != nil {
				m.overlayMode = OverlayRecurrenceSelect
				m.overlayCursor = 0
				return m, nil
			}
		}

		// View-specific keys
		if m.activeView == ViewList {
			return m.updateListView(msg)
		} else {
			return m.updateBoardView(msg)
		}

	case TasksLoadedMsg:
		m.tasks = msg.Tasks
		m.totalTasks = len(msg.Tasks)

		// Separate tasks into active (todo+doing) and done
		m.activeTasks = nil
		m.doneTasksList = nil
		for _, t := range msg.Tasks {
			if t.Status == model.StatusDone {
				m.doneTasksList = append(m.doneTasksList, t)
			} else {
				m.activeTasks = append(m.activeTasks, t)
			}
		}
		m.doneTaskCount = len(m.doneTasksList)

		m.clampCursor()
		m.clampDoneCursor()

	case ProjectsLoadedMsg:
		m.projects = msg.Projects

	case TagsLoadedMsg:
		m.tags = msg.Tags

	case TaskCreatedMsg:
		m.statusText = fmt.Sprintf("✓ Created: %s", msg.Task.Title)
		cmds = append(cmds, m.reloadTasks(), loadProjects(m.store), clearStatusAfter(1500*time.Millisecond))

	case TaskUpdatedMsg:
		m.statusText = "✓ Updated"
		cmds = append(cmds, m.reloadTasks(), loadProjects(m.store), clearStatusAfter(1500*time.Millisecond))

	case TaskDeletedMsg:
		m.statusText = "✓ Deleted"
		cmds = append(cmds, m.reloadTasks(), loadProjects(m.store), clearStatusAfter(1500*time.Millisecond))

	case ProjectCreatedMsg:
		m.statusText = fmt.Sprintf("✓ Created project: %s", msg.Project.Name)
		cmds = append(cmds, loadProjects(m.store), clearStatusAfter(1500*time.Millisecond))

	case TagCreatedMsg:
		m.statusText = fmt.Sprintf("✓ Created tag: %s", msg.Tag.Name)
		cmds = append(cmds, loadTags(m.store), m.reloadTasks(), clearStatusAfter(1500*time.Millisecond))

	case TagDeletedMsg:
		m.statusText = "✓ Tag deleted"
		cmds = append(cmds, loadTags(m.store), m.reloadTasks(), clearStatusAfter(1500*time.Millisecond))

	case ProjectDeletedMsg:
		m.statusText = "✓ Project deleted (tasks moved to Inbox)"
		cmds = append(cmds, loadProjects(m.store), m.reloadTasks(), clearStatusAfter(1500*time.Millisecond))

	case RecurrenceSetMsg:
		m.statusText = "✓ Recurrence set"
		cmds = append(cmds, m.reloadTasks(), clearStatusAfter(1500*time.Millisecond))

	case RecurrenceDeletedMsg:
		m.statusText = "✓ Recurrence removed"
		cmds = append(cmds, m.reloadTasks(), clearStatusAfter(1500*time.Millisecond))

	case ErrorMsg:
		m.statusText = msg.Err.Error()
		m.statusError = true
		cmds = append(cmds, clearStatusAfter(5*time.Second))

	case ClearStatusMsg:
		m.statusText = ""
		m.statusError = false
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.overlayMode {
	case OverlayHelp:
		// Any key closes help
		m.overlayMode = OverlayNone
		return m, nil

	case OverlayProjectSelect:
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case key.Matches(msg, Keys.Up):
			if m.overlayCursor > 0 {
				m.overlayCursor--
			}

		case key.Matches(msg, Keys.Down):
			if m.overlayCursor < len(m.projects) {
				m.overlayCursor++
			}

		case key.Matches(msg, Keys.Select):
			m.overlayMode = OverlayNone
			if m.overlayCursor == 0 {
				// "All" selected
				m.currentProject = nil
				m.currentProjectName = "All"
			} else {
				proj := m.projects[m.overlayCursor-1]
				m.currentProject = &proj.ID
				m.currentProjectName = proj.Name
			}
			return m, m.reloadTasks()

		case msg.String() == "n" || msg.String() == "a":
			// Open project create overlay
			m.overlayMode = OverlayProjectCreate
			m.projectFormName = ""
			m.projectFormDesc = ""
			m.projectFormFocus = 0
			return m, nil

		case msg.String() == "x":
			// Delete project (not allowed for "All" option or Inbox)
			if m.overlayCursor == 0 {
				m.statusText = "Cannot delete 'All' filter"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			proj := m.projects[m.overlayCursor-1]
			if proj.ID == 1 {
				m.statusText = "Cannot delete Inbox project"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			m.overlayMode = OverlayConfirmDeleteProject
			return m, nil
		}

	case OverlayConfirmDeleteProject:
		switch msg.String() {
		case "y", "Y":
			m.overlayMode = OverlayNone
			if m.overlayCursor > 0 && m.overlayCursor <= len(m.projects) {
				proj := m.projects[m.overlayCursor-1]
				// If currently viewing deleted project, switch to "All"
				if m.currentProject != nil && *m.currentProject == proj.ID {
					m.currentProject = nil
					m.currentProjectName = "All"
				}
				return m, deleteProject(m.store, proj.ID)
			}
		default:
			m.overlayMode = OverlayProjectSelect
		}
		return m, nil

	case OverlayConfirmDelete:
		switch msg.String() {
		case "y", "Y":
			m.overlayMode = OverlayNone
			if task := m.selectedTask(); task != nil {
				return m, deleteTask(m.store, task.ID)
			}
		default:
			m.overlayMode = OverlayNone
		}

	case OverlayDueDate:
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case key.Matches(msg, Keys.Up):
			if m.overlayCursor > 0 {
				m.overlayCursor--
			}

		case key.Matches(msg, Keys.Down):
			if m.overlayCursor < 4 {
				m.overlayCursor++
			}

		case key.Matches(msg, Keys.Select):
			m.overlayMode = OverlayNone
			task := m.selectedTask()
			if task == nil {
				return m, nil
			}

			now := time.Now()
			switch m.overlayCursor {
			case 0: // Today
				today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
				task.DueDate = &today
			case 1: // Tomorrow
				tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 23, 59, 0, 0, now.Location())
				task.DueDate = &tomorrow
			case 2: // Next week
				nextWeek := time.Date(now.Year(), now.Month(), now.Day()+7, 23, 59, 0, 0, now.Location())
				task.DueDate = &nextWeek
			case 3: // Clear
				task.DueDate = nil
			case 4: // Custom - open custom date overlay
				m.overlayMode = OverlayDueDateCustom
				m.dueDateFormValue = ""
				return m, nil
			}
			return m, updateTask(m.store, task)
		}

	case OverlayTagSelect:
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case key.Matches(msg, Keys.Up):
			if m.overlayCursor > 0 {
				m.overlayCursor--
			}

		case key.Matches(msg, Keys.Down):
			if m.overlayCursor < len(m.tags) {
				m.overlayCursor++
			}

		case key.Matches(msg, Keys.Select):
			task := m.selectedTask()
			if task == nil {
				m.overlayMode = OverlayNone
				return m, nil
			}

			if m.overlayCursor == len(m.tags) {
				// Add new tag - open tag create overlay
				m.overlayMode = OverlayTagCreate
				m.tagFormName = ""
				return m, nil
			}

			// Toggle tag on task
			tag := m.tags[m.overlayCursor]
			hasTag := false
			for _, t := range task.Tags {
				if t.ID == tag.ID {
					hasTag = true
					break
				}
			}

			var cmd tea.Cmd
			if hasTag {
				cmd = removeTagFromTask(m.store, task.ID, tag.ID)
			} else {
				cmd = addTagToTask(m.store, task.ID, tag.ID)
			}
			return m, cmd

		case msg.String() == "x":
			// Delete tag (not allowed for "+ New tag..." option)
			if m.overlayCursor >= len(m.tags) {
				return m, nil
			}
			m.overlayMode = OverlayConfirmDeleteTag
			return m, nil
		}

	case OverlayConfirmDeleteTag:
		switch msg.String() {
		case "y", "Y":
			m.overlayMode = OverlayNone
			if m.overlayCursor < len(m.tags) {
				tag := m.tags[m.overlayCursor]
				return m, deleteTag(m.store, tag.ID)
			}
		default:
			m.overlayMode = OverlayTagSelect
		}
		return m, nil

	case OverlayTagCreate:
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayTagSelect
			return m, nil

		case msg.Type == tea.KeyEnter:
			name := strings.TrimSpace(m.tagFormName)
			if name == "" {
				m.statusText = "Tag name is required"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			m.overlayMode = OverlayNone
			task := m.selectedTask()
			if task != nil {
				// Create tag and add to task
				return m, tea.Batch(
					createTagAndAddToTask(m.store, name, task.ID),
				)
			}
			return m, createTag(m.store, name)

		case msg.Type == tea.KeyBackspace:
			if len(m.tagFormName) > 0 {
				m.tagFormName = truncateRunes(m.tagFormName, 1)
			}

		case msg.Type == tea.KeyRunes:
			m.tagFormName += string(msg.Runes)

		case msg.Type == tea.KeySpace:
			m.tagFormName += " "
		}

	case OverlayDueDateCustom:
		// Default placeholder: 3 days from now
		placeholder := time.Now().AddDate(0, 0, 3).Format("2006-01-02")

		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayDueDate
			return m, nil

		case msg.Type == tea.KeyTab:
			// Autocomplete with placeholder
			if m.dueDateFormValue == "" {
				m.dueDateFormValue = placeholder
			}
			return m, nil

		case msg.Type == tea.KeyEnter:
			dateStr := strings.TrimSpace(m.dueDateFormValue)
			if dateStr == "" {
				m.statusText = "Date is required"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			task := m.selectedTask()
			if task == nil {
				m.overlayMode = OverlayNone
				return m, nil
			}
			// Parse date
			dueDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				m.statusText = "Invalid date format (use YYYY-MM-DD)"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 23, 59, 0, 0, time.Local)
			task.DueDate = &dueDate
			m.overlayMode = OverlayNone
			return m, updateTask(m.store, task)

		case msg.Type == tea.KeyBackspace:
			if len(m.dueDateFormValue) > 0 {
				m.dueDateFormValue = truncateRunes(m.dueDateFormValue, 1)
			}

		case msg.Type == tea.KeyRunes:
			m.dueDateFormValue += string(msg.Runes)
		}

	case OverlayProjectCreate:
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case msg.Type == tea.KeyTab || msg.Type == tea.KeyShiftTab:
			// Toggle between name and desc fields
			if m.projectFormFocus == 0 {
				m.projectFormFocus = 1
			} else {
				m.projectFormFocus = 0
			}

		case msg.Type == tea.KeyEnter:
			// Create project
			name := strings.TrimSpace(m.projectFormName)
			if name == "" {
				m.statusText = "Project name is required"
				m.statusError = true
				return m, clearStatusAfter(2 * time.Second)
			}
			m.overlayMode = OverlayNone
			return m, createProjectWithDesc(m.store, name, strings.TrimSpace(m.projectFormDesc))

		case msg.Type == tea.KeyBackspace:
			if m.projectFormFocus == 0 && len(m.projectFormName) > 0 {
				m.projectFormName = truncateRunes(m.projectFormName, 1)
			} else if m.projectFormFocus == 1 && len(m.projectFormDesc) > 0 {
				m.projectFormDesc = truncateRunes(m.projectFormDesc, 1)
			}

		case msg.Type == tea.KeyRunes:
			if m.projectFormFocus == 0 {
				m.projectFormName += string(msg.Runes)
			} else {
				m.projectFormDesc += string(msg.Runes)
			}

		case msg.Type == tea.KeySpace:
			if m.projectFormFocus == 0 {
				m.projectFormName += " "
			} else {
				m.projectFormDesc += " "
			}
		}

	case OverlayRecurrenceSelect:
		// Options: Daily, Weekly, Monthly, Yearly, Remove (if has recurrence)
		task := m.selectedTask()
		if task == nil {
			m.overlayMode = OverlayNone
			return m, nil
		}

		hasRecurrence := false
		if rec, _ := m.store.GetRecurrence(task.ID); rec != nil {
			hasRecurrence = true
		}

		maxCursor := 3 // Daily, Weekly, Monthly, Yearly (0-3)
		if hasRecurrence {
			maxCursor = 4 // + Remove option
		}

		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case key.Matches(msg, Keys.Up):
			if m.overlayCursor > 0 {
				m.overlayCursor--
			}

		case key.Matches(msg, Keys.Down):
			if m.overlayCursor < maxCursor {
				m.overlayCursor++
			}

		case key.Matches(msg, Keys.Select):
			m.overlayMode = OverlayNone
			switch m.overlayCursor {
			case 0:
				return m, setRecurrence(m.store, task.ID, model.Daily, 1)
			case 1:
				return m, setRecurrence(m.store, task.ID, model.Weekly, 1)
			case 2:
				return m, setRecurrence(m.store, task.ID, model.Monthly, 1)
			case 3:
				return m, setRecurrence(m.store, task.ID, model.Yearly, 1)
			case 4:
				// Remove recurrence
				return m, deleteRecurrence(m.store, task.ID)
			}
		}

	case OverlayTaskDetail:
		// Any key closes the detail overlay
		m.overlayMode = OverlayNone
		return m, nil

	case OverlayThemeSelect:
		maxCursor := len(styles.ThemeNames) - 1
		switch {
		case key.Matches(msg, Keys.Cancel):
			m.overlayMode = OverlayNone
			return m, nil

		case key.Matches(msg, Keys.Up):
			if m.overlayCursor > 0 {
				m.overlayCursor--
			}

		case key.Matches(msg, Keys.Down):
			if m.overlayCursor < maxCursor {
				m.overlayCursor++
			}

		case key.Matches(msg, Keys.Select):
			themeName := styles.ThemeNames[m.overlayCursor]
			config.SetTheme(themeName)
			config.Save()
			styles.ApplyTheme(themeName)
			m.overlayMode = OverlayNone
			m.statusText = "Theme: " + styles.Themes[themeName].Name
			return m, clearStatusAfter(2 * time.Second)
		}
	}

	return m, nil
}

func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.inputMode = InputNone
		m.textInput.Blur()
		return m, nil

	case tea.KeyTab:
		// Autocomplete with placeholder value
		if m.textInput.Value() == "" && m.textInput.Placeholder != "" {
			m.textInput.SetValue(m.textInput.Placeholder)
			m.textInput.CursorEnd()
		}
		return m, nil

	case tea.KeyEnter:
		value := strings.TrimSpace(m.textInput.Value())

		var cmd tea.Cmd
		switch m.inputMode {
		case InputAdd:
			if value != "" {
				cmd = createTask(m.store, value, m.currentProject)
			}
		case InputSearch:
			m.searchQuery = value
			cmd = m.reloadTasks()
		case InputEdit:
			if value != "" && m.editTaskID > 0 {
				task, _ := m.store.GetTask(m.editTaskID)
				if task != nil {
					task.Title = value
					cmd = updateTask(m.store, task)
				}
			}
		case InputAddProject:
			if value != "" {
				cmd = createProject(m.store, value)
			}
		case InputAddTag:
			if value != "" {
				cmd = createTag(m.store, value)
			}
		case InputDueDate:
			if value != "" {
				task := m.selectedTask()
				if task != nil {
					// Parse date
					t, err := time.Parse("2006-01-02", value)
					if err != nil {
						m.statusText = "Invalid date format (use YYYY-MM-DD)"
						m.statusError = true
						cmd = clearStatusAfter(2 * time.Second)
					} else {
						dueDate := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, time.Local)
						task.DueDate = &dueDate
						cmd = updateTask(m.store, task)
					}
				}
			}
		}

		m.inputMode = InputNone
		m.textInput.Blur()
		return m, cmd
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.inDoneSection {
			if m.doneCursor > 0 {
				m.doneCursor--
			} else if len(m.activeTasks) > 0 {
				// Move from done section to active section
				m.inDoneSection = false
				m.cursor = len(m.activeTasks) - 1
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case key.Matches(msg, Keys.Down):
		if m.inDoneSection {
			if m.doneCursor < len(m.doneTasksList)-1 {
				m.doneCursor++
			}
		} else {
			if m.cursor < len(m.activeTasks)-1 {
				m.cursor++
			} else if !m.doneCollapsed && len(m.doneTasksList) > 0 {
				// Move from active section to done section
				m.inDoneSection = true
				m.doneCursor = 0
			}
		}

	case key.Matches(msg, Keys.Select):
		// Enter: forward status (todo → doing → done)
		if task := m.selectedTask(); task != nil {
			switch task.Status {
			case model.StatusTodo:
				task.Status = model.StatusDoing
				return m, updateTask(m.store, task)
			case model.StatusDoing:
				return m, completeTask(m.store, task.ID)
			}
		}

	case msg.String() == "b":
		// b: backward status (done → doing → todo)
		if task := m.selectedTask(); task != nil {
			switch task.Status {
			case model.StatusDone:
				task.Status = model.StatusDoing
				task.CompletedAt = nil
				return m, updateTask(m.store, task)
			case model.StatusDoing:
				task.Status = model.StatusTodo
				return m, updateTask(m.store, task)
			}
		}

	case key.Matches(msg, Keys.Done):
		// d: directly mark as done (or toggle if already done)
		if task := m.selectedTask(); task != nil {
			if task.Status == model.StatusDone {
				task.Status = model.StatusTodo
				task.CompletedAt = nil
				return m, updateTask(m.store, task)
			}
			return m, completeTask(m.store, task.ID)
		}

	case key.Matches(msg, Keys.Delete):
		if m.selectedTask() != nil {
			m.overlayMode = OverlayConfirmDelete
			return m, nil
		}

	case msg.String() == "1":
		if task := m.selectedTask(); task != nil {
			task.Priority = model.PriorityHigh
			return m, updateTask(m.store, task)
		}
	case msg.String() == "2":
		if task := m.selectedTask(); task != nil {
			task.Priority = model.PriorityMedium
			return m, updateTask(m.store, task)
		}
	case msg.String() == "3":
		if task := m.selectedTask(); task != nil {
			task.Priority = model.PriorityLow
			return m, updateTask(m.store, task)
		}
	case msg.String() == "0":
		if task := m.selectedTask(); task != nil {
			task.Priority = model.PriorityNone
			return m, updateTask(m.store, task)
		}

	case msg.String() == "v":
		// v: view detail
		if task := m.selectedTask(); task != nil {
			m.overlayMode = OverlayTaskDetail
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) updateBoardView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	columns := m.tasksByStatus()

	switch {
	case key.Matches(msg, Keys.Left):
		if m.boardCol > 0 {
			m.boardCol--
			m.clampBoardCursor(columns)
		}

	case key.Matches(msg, Keys.Right):
		if m.boardCol < 2 {
			m.boardCol++
			m.clampBoardCursor(columns)
		}

	case key.Matches(msg, Keys.Up):
		if m.boardCursors[m.boardCol] > 0 {
			m.boardCursors[m.boardCol]--
		}

	case key.Matches(msg, Keys.Down):
		col := columns[m.boardCol]
		if m.boardCursors[m.boardCol] < len(col)-1 {
			m.boardCursors[m.boardCol]++
		}

	case key.Matches(msg, Keys.Done), key.Matches(msg, Keys.Select):
		// Enter: forward status (todo → doing → done)
		if task := m.selectedBoardTask(columns); task != nil {
			switch task.Status {
			case model.StatusTodo:
				task.Status = model.StatusDoing
				return m, updateTask(m.store, task)
			case model.StatusDoing:
				return m, completeTask(m.store, task.ID)
			}
		}

	case msg.String() == "b":
		// b: backward status (done → doing → todo)
		if task := m.selectedBoardTask(columns); task != nil {
			switch task.Status {
			case model.StatusDone:
				task.Status = model.StatusDoing
				task.CompletedAt = nil
				return m, updateTask(m.store, task)
			case model.StatusDoing:
				task.Status = model.StatusTodo
				return m, updateTask(m.store, task)
			}
		}

	case key.Matches(msg, Keys.Delete):
		if m.selectedBoardTask(columns) != nil {
			m.overlayMode = OverlayConfirmDelete
			return m, nil
		}

	case msg.String() == "1":
		if task := m.selectedBoardTask(columns); task != nil {
			task.Priority = model.PriorityHigh
			return m, updateTask(m.store, task)
		}
	case msg.String() == "2":
		if task := m.selectedBoardTask(columns); task != nil {
			task.Priority = model.PriorityMedium
			return m, updateTask(m.store, task)
		}
	case msg.String() == "3":
		if task := m.selectedBoardTask(columns); task != nil {
			task.Priority = model.PriorityLow
			return m, updateTask(m.store, task)
		}
	case msg.String() == "0":
		if task := m.selectedBoardTask(columns); task != nil {
			task.Priority = model.PriorityNone
			return m, updateTask(m.store, task)
		}

	case msg.String() == "v":
		// v: view detail
		if task := m.selectedBoardTask(columns); task != nil {
			m.overlayMode = OverlayTaskDetail
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "\n\n  Loading..."
	}

	// Calculate box dimensions (centered with margins)
	boxWidth := m.width - 8
	boxHeight := m.height - 4
	if boxWidth < 60 {
		boxWidth = 60
	}
	if boxWidth > 180 {
		boxWidth = 180
	}
	if boxHeight < 20 {
		boxHeight = 20
	}
	if boxHeight > 45 {
		boxHeight = 45
	}

	// Inner content width (accounting for border and padding)
	innerWidth := boxWidth - 6

	var b strings.Builder

	// Header with tabs
	b.WriteString(m.renderHeaderWithWidth(innerWidth))
	b.WriteString("\n\n")

	// Input area (if in input mode)
	if m.inputMode != InputNone {
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary).
			Padding(0, 1).
			Width(innerWidth - 4).
			Render(styles.InputPrompt.Render(m.inputPrompt) + m.textInput.View())
		b.WriteString(inputBox)
		b.WriteString("\n\n")
	}

	// Inner height (accounting for header, status bar, border, padding)
	innerHeight := boxHeight - 8

	// Main content
	if m.activeView == ViewList {
		b.WriteString(m.renderListViewWithSize(innerWidth, innerHeight))
	} else {
		b.WriteString(m.renderBoardViewWithSize(innerWidth, innerHeight))
	}

	// Build main content
	content := b.String()

	// Status bar
	statusBar := m.renderStatusBar()

	// Add status bar at bottom
	fullContent := content + "\n\n" + statusBar

	// Create bordered box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(boxWidth).
		Height(boxHeight).
		Render(fullContent)

	// Center the box on screen
	fullScreen := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(styles.Background),
	)

	// Render overlay on top
	if m.overlayMode != OverlayNone {
		overlay := m.renderOverlay()
		return m.placeOverlay(fullScreen, overlay)
	}

	return fullScreen
}

func (m Model) renderOverlay() string {
	switch m.overlayMode {
	case OverlayHelp:
		return m.renderHelpOverlay()
	case OverlayProjectSelect:
		return m.renderProjectSelectOverlay()
	case OverlayProjectCreate:
		return m.renderProjectCreateOverlay()
	case OverlayConfirmDelete:
		return m.renderConfirmDeleteOverlay()
	case OverlayConfirmDeleteTag:
		return m.renderConfirmDeleteTagOverlay()
	case OverlayConfirmDeleteProject:
		return m.renderConfirmDeleteProjectOverlay()
	case OverlayDueDate:
		return m.renderDueDateOverlay()
	case OverlayDueDateCustom:
		return m.renderDueDateCustomOverlay()
	case OverlayTagSelect:
		return m.renderTagSelectOverlay()
	case OverlayTagCreate:
		return m.renderTagCreateOverlay()
	case OverlayRecurrenceSelect:
		return m.renderRecurrenceSelectOverlay()
	case OverlayTaskDetail:
		return m.renderTaskDetailOverlay()
	case OverlayThemeSelect:
		return m.renderThemeSelectOverlay()
	}
	return ""
}

func (m Model) renderHelpOverlay() string {
	title := styles.Header.Render("Keyboard Shortcuts")

	sections := []string{
		title,
		"",
		styles.HelpKey.Render("Navigation"),
		"  ↑/k, ↓/j    Move up/down",
		"  ←/h, →/l    Move between columns (board)",
		"  Tab         Switch view (List/Board)",
		"",
		styles.HelpKey.Render("Status"),
		"  Enter       Forward (todo → doing → done)",
		"  b           Backward (done → doing → todo)",
		"",
		styles.HelpKey.Render("Actions"),
		"  a           Add new task",
		"  e           Edit task title",
		"  D           Toggle done",
		"  x           Delete task",
		"  d           Set due date",
		"  t           Set tags",
		"  r           Set recurrence",
		"  1/2/3/0     Set priority (high/med/low/none)",
		"  v           View task detail",
		"",
		styles.HelpKey.Render("Filter & Search"),
		"  /           Search tasks",
		"  p           Select project",
		"  A           Toggle Done section",
		"  c           Clear search",
		"",
		styles.HelpKey.Render("General"),
		"  ?           Show this help",
		"  q           Quit",
		"",
		styles.MutedStyle.Render("Press any key to close"),
	}

	content := strings.Join(sections, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderProjectSelectOverlay() string {
	title := styles.Header.Render("Select Project")

	var items []string
	items = append(items, title, "")

	// "All" option
	allStyle := styles.TaskItem
	if m.overlayCursor == 0 {
		allStyle = styles.TaskItemSelected
	}
	items = append(items, allStyle.Render("All Projects"))

	// Project list
	for i, proj := range m.projects {
		style := styles.TaskItem
		if m.overlayCursor == i+1 {
			style = styles.TaskItemSelected
		}
		text := fmt.Sprintf("%s (%d/%d)", proj.Name, proj.DoneCount, proj.TaskCount)
		items = append(items, style.Render(text))
	}

	items = append(items, "", styles.MutedStyle.Render("Enter: select  n: new  x: delete  Esc: cancel"))

	content := strings.Join(items, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(50).
		Render(content)
}

func (m Model) renderProjectCreateOverlay() string {
	title := styles.Header.Render("Create Project")

	// Name field
	nameLabel := "Name:"
	nameValue := m.projectFormName
	if nameValue == "" {
		nameValue = "_"
	}
	nameStyle := styles.TaskItem
	if m.projectFormFocus == 0 {
		nameStyle = styles.TaskItemSelected
	}
	nameField := nameStyle.Render(fmt.Sprintf("  %s %s", nameLabel, nameValue))

	// Description field
	descLabel := "Description:"
	descValue := m.projectFormDesc
	if descValue == "" {
		descValue = "(optional)"
	}
	descStyle := styles.TaskItem
	if m.projectFormFocus == 1 {
		descStyle = styles.TaskItemSelected
	}
	descField := descStyle.Render(fmt.Sprintf("  %s %s", descLabel, descValue))

	content := strings.Join([]string{
		title,
		"",
		nameField,
		"",
		descField,
		"",
		styles.MutedStyle.Render("Tab: switch  Enter: create  Esc: cancel"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(45).
		Render(content)
}

func (m Model) renderConfirmDeleteOverlay() string {
	task := m.selectedTask()
	if task == nil {
		return ""
	}

	title := styles.Header.Render("Delete Task?")
	taskTitle := styles.TaskTitle.Render(task.Title)

	content := strings.Join([]string{
		title,
		"",
		taskTitle,
		"",
		styles.MutedStyle.Render("y: yes  n: no"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Danger).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderDueDateOverlay() string {
	title := styles.Header.Render("Set Due Date")

	options := []string{"Today", "Tomorrow", "Next week", "Clear", "Custom..."}
	var items []string
	items = append(items, title, "")

	for i, opt := range options {
		style := styles.TaskItem
		if m.overlayCursor == i {
			style = styles.TaskItemSelected
		}
		items = append(items, style.Render(opt))
	}

	items = append(items, "", styles.MutedStyle.Render("Enter: select  Esc: cancel"))

	content := strings.Join(items, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(30).
		Render(content)
}

func (m Model) renderDueDateCustomOverlay() string {
	title := styles.Header.Render("Custom Due Date")
	placeholder := time.Now().AddDate(0, 0, 3).Format("2006-01-02")

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Primary).
		Padding(0, 1).
		Width(20)

	var inputField string
	if m.dueDateFormValue == "" {
		// Empty: cursor on left, placeholder as hint on right
		inputField = inputStyle.Render("▌" + styles.MutedStyle.Render(placeholder))
	} else {
		// Has value: show value with cursor at end
		inputField = inputStyle.Render(m.dueDateFormValue + "▌")
	}

	content := strings.Join([]string{
		title,
		"",
		"Format: YYYY-MM-DD",
		"",
		inputField,
		"",
		styles.MutedStyle.Render("Tab: autocomplete  Enter: confirm  Esc: back"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderTagSelectOverlay() string {
	title := styles.Header.Render("Select Tags")

	task := m.selectedTask()
	var items []string
	items = append(items, title, "")

	// List existing tags
	for i, tag := range m.tags {
		style := styles.TaskItem
		if m.overlayCursor == i {
			style = styles.TaskItemSelected
		}

		// Check if task has this tag
		hasTag := false
		if task != nil {
			for _, t := range task.Tags {
				if t.ID == tag.ID {
					hasTag = true
					break
				}
			}
		}

		marker := "  "
		if hasTag {
			marker = "✓ "
		}
		items = append(items, style.Render(marker+tag.Name))
	}

	// Add new tag option
	style := styles.TaskItem
	if m.overlayCursor == len(m.tags) {
		style = styles.TaskItemSelected
	}
	items = append(items, style.Render("+ New tag..."))

	items = append(items, "", styles.MutedStyle.Render("Enter: toggle  x: delete  Esc: cancel"))

	content := strings.Join(items, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(45).
		Render(content)
}

func (m Model) renderTagCreateOverlay() string {
	title := styles.Header.Render("Create Tag")

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Primary).
		Padding(0, 1).
		Width(25)

	displayValue := m.tagFormName
	if displayValue == "" {
		displayValue = styles.MutedStyle.Render("Tag name...")
	}
	inputField := inputStyle.Render(displayValue + "▌")

	content := strings.Join([]string{
		title,
		"",
		inputField,
		"",
		styles.MutedStyle.Render("Enter: create  Esc: back"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderTaskDetailOverlay() string {
	// Get selected task based on current view
	var task *model.Task
	if m.activeView == ViewBoard {
		columns := m.tasksByStatus()
		task = m.selectedBoardTask(columns)
	} else {
		task = m.selectedTask()
	}

	if task == nil {
		return ""
	}

	title := styles.Header.Render("Task Detail")

	var lines []string
	lines = append(lines, title, "")

	// Project
	lines = append(lines, styles.HelpKey.Render("Project"))
	projectName := "None"
	if task.ProjectID != nil {
		for _, p := range m.projects {
			if p.ID == *task.ProjectID {
				projectName = p.Name
				break
			}
		}
	}
	lines = append(lines, "  "+projectName, "")

	// Title (wrap long titles)
	lines = append(lines, styles.HelpKey.Render("Title"))
	titleText := lipgloss.NewStyle().Width(56).Render("  " + task.Title)
	lines = append(lines, titleText, "")

	// Status
	lines = append(lines, styles.HelpKey.Render("Status"))
	var statusText string
	switch task.Status {
	case model.StatusTodo:
		statusText = "○ Todo"
	case model.StatusDoing:
		statusText = "◐ Doing"
	case model.StatusDone:
		statusText = "● Done"
	}
	lines = append(lines, "  "+statusText, "")

	// Priority
	lines = append(lines, styles.HelpKey.Render("Priority"))
	var priorityText string
	switch task.Priority {
	case model.PriorityHigh:
		priorityText = "↑ High"
	case model.PriorityMedium:
		priorityText = "→ Medium"
	case model.PriorityLow:
		priorityText = "↓ Low"
	default:
		priorityText = "- None"
	}
	lines = append(lines, "  "+priorityText, "")

	// Due Date
	lines = append(lines, styles.HelpKey.Render("Due Date"))
	if task.DueDate != nil {
		lines = append(lines, "  "+task.DueDate.Format("2006-01-02 (Mon)"))
	} else {
		lines = append(lines, "  Not set")
	}
	lines = append(lines, "")

	// Completed At
	if task.Status == model.StatusDone && task.CompletedAt != nil {
		lines = append(lines, styles.HelpKey.Render("Completed At"))
		lines = append(lines, "  "+task.CompletedAt.Format("2006-01-02 15:04"))
		if task.DueDate != nil {
			due := *task.DueDate
			completed := *task.CompletedAt
			dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location())
			completedDay := time.Date(completed.Year(), completed.Month(), completed.Day(), 0, 0, 0, 0, completed.Location())
			diffDays := int(completedDay.Sub(dueDay).Hours() / 24)
			if diffDays > 0 {
				lines = append(lines, fmt.Sprintf("  (%d days late)", diffDays))
			} else if diffDays < 0 {
				lines = append(lines, fmt.Sprintf("  (%d days early)", -diffDays))
			} else {
				lines = append(lines, "  (On time)")
			}
		}
		lines = append(lines, "")
	}

	// Tags
	lines = append(lines, styles.HelpKey.Render("Tags"))
	if len(task.Tags) > 0 {
		var tagNames []string
		for _, tag := range task.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		lines = append(lines, "  "+strings.Join(tagNames, ", "))
	} else {
		lines = append(lines, "  No tags")
	}
	lines = append(lines, "")

	// Created At
	lines = append(lines, styles.HelpKey.Render("Created At"))
	lines = append(lines, "  "+task.CreatedAt.Format("2006-01-02 15:04"), "")

	lines = append(lines, styles.MutedStyle.Render("Press any key to close"))

	content := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(60).
		Render(content)
}

func (m Model) renderConfirmDeleteTagOverlay() string {
	if m.overlayCursor >= len(m.tags) {
		return ""
	}
	tag := m.tags[m.overlayCursor]

	title := styles.Header.Render("Delete Tag?")
	tagName := styles.Tag.Render(tag.Name)
	warning := styles.MutedStyle.Render("This will remove the tag from all tasks.")

	content := strings.Join([]string{
		title,
		"",
		tagName,
		warning,
		"",
		styles.MutedStyle.Render("y: yes  n: no"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Danger).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderConfirmDeleteProjectOverlay() string {
	if m.overlayCursor <= 0 || m.overlayCursor > len(m.projects) {
		return ""
	}
	proj := m.projects[m.overlayCursor-1]

	title := styles.Header.Render("Delete Project?")
	projName := styles.ProjectBadge.Render(proj.Name)
	warning := styles.MutedStyle.Render("Tasks will be moved to Inbox.")

	content := strings.Join([]string{
		title,
		"",
		projName,
		warning,
		"",
		styles.MutedStyle.Render("y: yes  n: no"),
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Danger).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderRecurrenceSelectOverlay() string {
	task := m.selectedTask()
	if task == nil {
		return ""
	}

	title := styles.Header.Render("Set Recurrence")

	hasRecurrence := false
	var currentPattern string
	if rec, _ := m.store.GetRecurrence(task.ID); rec != nil {
		hasRecurrence = true
		currentPattern = rec.PatternString()
	}

	options := []string{"Daily", "Weekly", "Monthly", "Yearly"}
	if hasRecurrence {
		options = append(options, "Remove")
	}

	var items []string
	items = append(items, title, "")

	if hasRecurrence {
		items = append(items, styles.MutedStyle.Render("Current: "+currentPattern), "")
	}

	for i, opt := range options {
		style := styles.TaskItem
		if m.overlayCursor == i {
			style = styles.TaskItemSelected
		}

		prefix := "  "
		if i == 4 {
			// Remove option
			items = append(items, styles.MutedStyle.Render("─────────"))
			prefix = "  "
		}
		items = append(items, prefix+style.Render(opt))
	}

	items = append(items, "", styles.MutedStyle.Render("↑/↓: select  Enter: confirm  Esc: cancel"))

	content := strings.Join(items, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(content)
}

func (m Model) renderThemeSelectOverlay() string {
	title := styles.Header.Render("Select Theme")

	currentTheme := config.GetTheme()

	var items []string
	items = append(items, title, "")

	for i, name := range styles.ThemeNames {
		theme := styles.Themes[name]
		style := styles.TaskItem
		if m.overlayCursor == i {
			style = styles.TaskItemSelected
		}

		// Show theme name with color preview
		preview := lipgloss.NewStyle().
			Foreground(theme.Primary).
			Render("●")

		check := "  "
		if name == currentTheme {
			check = "✓ "
		}

		items = append(items, style.Render(check+preview+" "+theme.Name))
	}

	items = append(items, "", styles.MutedStyle.Render("↑/↓: select  Enter: apply  Esc: cancel"))

	content := strings.Join(items, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(content)
}

func (m Model) placeOverlay(base, overlay string) string {
	// Center the overlay on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(styles.Background),
	)
}

func (m Model) renderHeader() string {
	return m.renderHeaderWithWidth(m.width - 8)
}

func (m Model) renderHeaderWithWidth(width int) string {
	// App title with icon
	title := styles.AppTitle.Render(" tsk ")

	// View tabs
	listTab := styles.InactiveTab.Render(" List ")
	boardTab := styles.InactiveTab.Render(" Board ")

	if m.activeView == ViewList {
		listTab = styles.ActiveTab.Render(" List ")
	} else {
		boardTab = styles.ActiveTab.Render(" Board ")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, listTab, boardTab)

	// Project badge
	projectBadge := styles.ProjectBadge.Render(m.currentProjectName)

	// Search indicator
	searchBadge := ""
	if m.searchQuery != "" {
		searchBadge = " " + styles.SearchBadge.Render("/" + m.searchQuery)
	}

	// Task count
	var taskCount string
	activeCount := len(m.activeTasks)
	if m.doneTaskCount > 0 {
		taskCount = styles.MutedStyle.Render(fmt.Sprintf(" %d tasks, %d done", activeCount, m.doneTaskCount))
	} else {
		taskCount = styles.MutedStyle.Render(fmt.Sprintf(" %d tasks", activeCount))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", tabs)
	right := lipgloss.JoinHorizontal(lipgloss.Center, projectBadge, searchBadge, taskCount)

	// Create header bar
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gap), right)
}

func (m Model) renderListView() string {
	return m.renderListViewWithSize(m.width-8, m.height-10)
}

func (m Model) renderListViewWithSize(width, height int) string {
	if len(m.activeTasks) == 0 && len(m.doneTasksList) == 0 {
		msg := "No tasks. Press 'a' to add one."
		if m.searchQuery != "" {
			msg = "No tasks match your search. Press 'c' to clear."
		}
		return styles.Empty.Width(width).Render(msg)
	}

	var b strings.Builder

	// Calculate available height
	totalHeight := height - 2
	if totalHeight < 8 {
		totalHeight = 8
	}

	// Reserve space for done section header (always visible when there are done tasks)
	doneHeaderLines := 0
	if len(m.doneTasksList) > 0 {
		doneHeaderLines = 2 // blank line + header
	}

	// Calculate max visible for each section
	activeMaxVisible := totalHeight - doneHeaderLines
	doneMaxVisible := 0

	if !m.doneCollapsed && len(m.doneTasksList) > 0 {
		// Split space more evenly between active and done sections
		availableForTasks := totalHeight - doneHeaderLines
		activeNeeded := min(len(m.activeTasks), availableForTasks/2)
		doneMaxVisible = availableForTasks - activeNeeded
		if doneMaxVisible > len(m.doneTasksList) {
			doneMaxVisible = len(m.doneTasksList)
		}
		activeMaxVisible = availableForTasks - doneMaxVisible
		// Ensure minimum visibility
		if activeMaxVisible < 3 && len(m.activeTasks) > 0 {
			activeMaxVisible = 3
			doneMaxVisible = availableForTasks - activeMaxVisible
		}
	}

	// Render active tasks (todo + doing)
	if len(m.activeTasks) > 0 {
		start := 0
		if !m.inDoneSection && m.cursor >= activeMaxVisible {
			start = m.cursor - activeMaxVisible + 1
		}
		end := start + activeMaxVisible
		if end > len(m.activeTasks) {
			end = len(m.activeTasks)
		}

		for i := start; i < end; i++ {
			selected := !m.inDoneSection && i == m.cursor
			b.WriteString(m.renderTaskItemWithWidth(m.activeTasks[i], selected, width))
			b.WriteString("\n")
		}

		// Scroll indicator for active section
		if len(m.activeTasks) > activeMaxVisible {
			info := fmt.Sprintf("(%d/%d)", m.cursor+1, len(m.activeTasks))
			b.WriteString(styles.MutedStyle.Render(info))
			b.WriteString("\n")
		}
	} else if len(m.doneTasksList) > 0 {
		b.WriteString(styles.MutedStyle.Render("No active tasks"))
		b.WriteString("\n")
	}

	// Render done section
	if len(m.doneTasksList) > 0 {
		b.WriteString("\n")

		// Done section header
		collapseIcon := "▼"
		if m.doneCollapsed {
			collapseIcon = "▶"
		}
		headerStyle := styles.DoneSectionHeader
		if m.inDoneSection {
			headerStyle = styles.DoneSectionHeaderActive
		}
		header := headerStyle.Render(fmt.Sprintf("%s Done (%d)", collapseIcon, len(m.doneTasksList)))
		b.WriteString(header)
		b.WriteString("\n")

		// Done tasks (if expanded)
		if !m.doneCollapsed {
			start := 0
			if m.inDoneSection && m.doneCursor >= doneMaxVisible && doneMaxVisible > 0 {
				start = m.doneCursor - doneMaxVisible + 1
			}
			end := start + doneMaxVisible
			if end > len(m.doneTasksList) {
				end = len(m.doneTasksList)
			}

			for i := start; i < end; i++ {
				selected := m.inDoneSection && i == m.doneCursor
				b.WriteString(m.renderTaskItemWithWidth(m.doneTasksList[i], selected, width))
				b.WriteString("\n")
			}

			// Scroll indicator for done section
			if len(m.doneTasksList) > doneMaxVisible {
				info := fmt.Sprintf("(%d/%d done)", m.doneCursor+1, len(m.doneTasksList))
				b.WriteString(styles.MutedStyle.Render(info))
			}
		}
	}

	return b.String()
}

func (m Model) renderBoardView() string {
	return m.renderBoardViewWithSize(m.width-8, m.height-10)
}

func (m Model) renderBoardViewWithSize(width, height int) string {
	columns := m.tasksByStatus()
	// Account for column margins and gaps: 3 columns + 2 gaps
	colWidth := (width - 4) / 3
	if colWidth < 20 {
		colWidth = 20
	}
	if colWidth > 55 {
		colWidth = 55
	}

	// Column height (subtract column border/padding: 4)
	colHeight := height
	if colHeight < 8 {
		colHeight = 8
	}

	todoCol := m.renderColumnWithHeight("Todo", columns[0], 0, colWidth, colHeight)
	doingCol := m.renderColumnWithHeight("Doing", columns[1], 1, colWidth, colHeight)
	doneCol := m.renderColumnWithHeight("Done", columns[2], 2, colWidth, colHeight)

	board := lipgloss.JoinHorizontal(lipgloss.Top, todoCol, " ", doingCol, " ", doneCol)

	// Force board to fixed height
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(board)
}

func (m Model) renderColumn(title string, tasks []model.Task, colIdx int, width int) string {
	return m.renderColumnWithHeight(title, tasks, colIdx, width, m.height-10)
}

func (m Model) renderColumnWithHeight(title string, tasks []model.Task, colIdx int, width int, height int) string {
	var titleStyle lipgloss.Style
	switch colIdx {
	case 0:
		titleStyle = styles.ColumnTitleTodo
	case 1:
		titleStyle = styles.ColumnTitleDoing
	case 2:
		titleStyle = styles.ColumnTitleDone
	}

	header := titleStyle.Render(fmt.Sprintf(" %s (%d) ", title, len(tasks)))

	// Column inner height (subtract border 2 + padding 2)
	innerHeight := height - 4
	if innerHeight < 5 {
		innerHeight = 5
	}

	// Available lines for items: innerHeight - header(1) - blank(1)
	// Each item = 1 line, maxItems items shown
	maxItems := innerHeight - 2
	if maxItems < 3 {
		maxItems = 3
	}

	cursor := m.boardCursors[colIdx]

	// Calculate scroll position (List-style)
	start := 0
	if cursor >= maxItems {
		start = cursor - maxItems + 1
	}

	end := start + maxItems
	if end > len(tasks) {
		end = len(tasks)
	}

	// Build content lines (fixed height)
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "") // blank after header

	// Render items (1 line each)
	for i := start; i < end; i++ {
		selected := m.boardCol == colIdx && cursor == i
		item := m.renderBoardTaskItem(tasks[i], selected, width-4)
		lines = append(lines, item)
	}

	// Show scroll indicator or empty message (always reserve this line)
	if len(tasks) == 0 {
		emptyMsg := styles.MutedStyle.Render("No tasks")
		lines = append(lines, emptyMsg)
	} else if len(tasks) > maxItems {
		info := fmt.Sprintf("(%d/%d)", cursor+1, len(tasks))
		lines = append(lines, styles.MutedStyle.Render(info))
	} else {
		// Reserve line for consistent height
		lines = append(lines, "")
	}

	// Pad to exact innerHeight
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	// Apply column style with fixed height
	var colStyle lipgloss.Style
	if m.boardCol == colIdx {
		colStyle = styles.ColumnActive.Width(width).Height(height)
	} else {
		colStyle = styles.Column.Width(width).Height(height)
	}

	return colStyle.Render(content)
}

func (m Model) renderTaskItem(task model.Task, selected bool) string {
	return m.renderTaskItemWithWidth(task, selected, m.width-30)
}

func (m Model) renderTaskItemWithWidth(task model.Task, selected bool, width int) string {
	var statusIcon string
	switch task.Status {
	case model.StatusTodo:
		statusIcon = styles.StatusTodoStyle.Render("○")
	case model.StatusDoing:
		statusIcon = styles.StatusDoingStyle.Render("◐")
	case model.StatusDone:
		statusIcon = styles.StatusDoneStyle.Render("●")
	}

	var priority string
	switch task.Priority {
	case model.PriorityLow:
		priority = styles.PriorityLowStyle.Render("↓")
	case model.PriorityMedium:
		priority = styles.PriorityMediumStyle.Render("→")
	case model.PriorityHigh:
		priority = styles.PriorityHighStyle.Render("↑")
	default:
		priority = " "
	}

	titleStyle := styles.TaskTitle
	if task.Status == model.StatusDone {
		titleStyle = styles.TaskTitleDone
	}

	maxTitleLen := width - 20
	if maxTitleLen < 20 {
		maxTitleLen = 20
	}
	title := task.Title
	titleRunes := []rune(title)
	if len(titleRunes) > maxTitleLen {
		title = string(titleRunes[:maxTitleLen-3]) + "..."
	}
	title = titleStyle.Render(title)

	var due string
	if task.DueDate != nil && task.Status != model.StatusDone {
		due = m.formatDue(task)
	}

	var completedAt string
	if task.Status == model.StatusDone {
		completedAt = m.formatCompletedAt(task)
	}

	var tags string
	for _, tag := range task.Tags {
		tags += styles.Tag.Render(tag.Name)
	}

	// Build suffix (due/completed/tags)
	var suffix string
	if due != "" {
		suffix += " " + due
	}
	if completedAt != "" {
		suffix += " " + completedAt
	}
	if tags != "" {
		suffix += " " + tags
	}

	// Rebuild if line too long
	line := fmt.Sprintf("%s %s %s%s", statusIcon, priority, title, suffix)
	if lipgloss.Width(line) > width {
		// Recalculate title length
		prefixLen := 5 // status + priority + spaces
		suffixLen := lipgloss.Width(suffix)
		availableForTitle := width - prefixLen - suffixLen - 3
		if availableForTitle < 10 {
			availableForTitle = 10
		}
		title = task.Title
		for lipgloss.Width(title) > availableForTitle && len([]rune(title)) > 0 {
			title = truncateRunes(title, 1)
		}
		title = titleStyle.Render(title + "...")
		line = fmt.Sprintf("%s %s %s%s", statusIcon, priority, title, suffix)
	}

	if selected {
		return styles.TaskItemSelected.Render(line)
	}
	return styles.TaskItem.Render(line)
}

func (m Model) renderBoardTaskItem(task model.Task, selected bool, maxWidth int) string {
	// Build single line: priority + title + due/completed
	var priority string
	if selected {
		switch task.Priority {
		case model.PriorityLow:
			priority = "↓ "
		case model.PriorityMedium:
			priority = "→ "
		case model.PriorityHigh:
			priority = "↑ "
		}
	} else {
		switch task.Priority {
		case model.PriorityLow:
			priority = styles.PriorityLowStyle.Render("↓") + " "
		case model.PriorityMedium:
			priority = styles.PriorityMediumStyle.Render("→") + " "
		case model.PriorityHigh:
			priority = styles.PriorityHighStyle.Render("↑") + " "
		}
	}

	// Due/completed suffix
	var suffix string
	if task.Status == model.StatusDone {
		if task.CompletedAt != nil {
			completed := *task.CompletedAt
			text := "✓ " + completed.Format("06/1/2")
			// Show difference from due date
			if task.DueDate != nil {
				due := *task.DueDate
				dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location())
				completedDay := time.Date(completed.Year(), completed.Month(), completed.Day(), 0, 0, 0, 0, completed.Location())
				diffDays := int(completedDay.Sub(dueDay).Hours() / 24)
				if diffDays > 0 {
					text += fmt.Sprintf(" +%dd", diffDays)
				} else if diffDays < 0 {
					text += fmt.Sprintf(" %dd", diffDays)
				}
			}
			if selected {
				suffix = " " + text
			} else {
				suffix = " " + styles.CompletedDate.Render(text)
			}
		}
	} else if task.DueDate != nil {
		if selected {
			suffix = " · " + task.DueDate.Format("06/1/2")
		} else {
			suffix = " · " + m.formatDue(task)
		}
	}

	// Calculate available width for title
	// Priority takes ~2 chars, suffix varies
	suffixLen := lipgloss.Width(suffix)
	availableWidth := maxWidth - 3 - suffixLen // 3 for priority + space
	if availableWidth < 10 {
		availableWidth = 10
	}

	title := task.Title
	if lipgloss.Width(title) > availableWidth {
		// Truncate title
		for lipgloss.Width(title) > availableWidth-3 && len([]rune(title)) > 0 {
			title = truncateRunes(title, 1)
		}
		title += "..."
	}

	var titleText string
	if selected {
		titleText = title
	} else if task.Status == model.StatusDone {
		titleText = styles.TaskTitleDone.Render(title)
	} else {
		titleText = styles.TaskTitle.Render(title)
	}

	line := priority + titleText + suffix

	// Force single line - truncate if too long
	if lipgloss.Width(line) > maxWidth {
		// Rebuild with shorter title
		availableForTitle := maxWidth - lipgloss.Width(priority+suffix) - 3
		if availableForTitle < 5 {
			availableForTitle = 5
		}
		title := task.Title
		for lipgloss.Width(title) > availableForTitle && len([]rune(title)) > 0 {
			title = truncateRunes(title, 1)
		}
		title += "..."
		if selected {
			titleText = title
		} else if task.Status == model.StatusDone {
			titleText = styles.TaskTitleDone.Render(title)
		} else {
			titleText = styles.TaskTitle.Render(title)
		}
		line = priority + titleText + suffix
	}

	if selected {
		return styles.TaskItemSelected.Render(line)
	}
	return styles.TaskItem.Render(line)
}

func (m Model) formatDue(task model.Task) string {
	if task.DueDate == nil {
		return ""
	}

	now := time.Now()
	due := *task.DueDate
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location())

	diff := dueDay.Sub(today).Hours() / 24

	var text string
	var dueStyle lipgloss.Style

	switch {
	case diff < 0:
		text = "Overdue"
		dueStyle = styles.DueOverdue
	case diff == 0:
		text = "Today"
		dueStyle = styles.DueToday
	case diff == 1:
		text = "Tomorrow"
		dueStyle = styles.DueNormal
	case diff <= 7:
		text = due.Format("Mon")
		dueStyle = styles.DueNormal
	default:
		text = due.Format("06/1/2")
		dueStyle = styles.DueNormal
	}

	return dueStyle.Render("by " + text)
}

func (m Model) formatCompletedAt(task model.Task) string {
	if task.CompletedAt == nil || task.Status != model.StatusDone {
		return ""
	}

	completed := *task.CompletedAt
	text := "✓ " + completed.Format("06/1/2 15:04")

	// Compare with due date if exists
	if task.DueDate != nil {
		due := *task.DueDate
		dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location())
		completedDay := time.Date(completed.Year(), completed.Month(), completed.Day(), 0, 0, 0, 0, completed.Location())
		diffDays := int(completedDay.Sub(dueDay).Hours() / 24)

		if diffDays > 0 {
			text += fmt.Sprintf(" (+%dd)", diffDays)
		} else if diffDays < 0 {
			text += fmt.Sprintf(" (%dd)", diffDays)
		}
	}

	return styles.CompletedDate.Render(text)
}

func (m Model) renderStatusBar() string {
	help := []string{
		styles.HelpKey.Render("↑↓") + styles.HelpDesc.Render(":move"),
		styles.HelpKey.Render("⏎") + styles.HelpDesc.Render(":status"),
		styles.HelpKey.Render("a") + styles.HelpDesc.Render(":add"),
		styles.HelpKey.Render("e") + styles.HelpDesc.Render(":edit"),
		styles.HelpKey.Render("x") + styles.HelpDesc.Render(":del"),
		styles.HelpKey.Render("p") + styles.HelpDesc.Render(":project"),
		styles.HelpKey.Render("?") + styles.HelpDesc.Render(":help"),
	}
	helpText := strings.Join(help, "  ")

	if m.statusText != "" {
		// Show help on left, status message on right
		var statusStyled string
		if m.statusError {
			statusStyled = styles.StatusBarError.Render(m.statusText)
		} else {
			statusStyled = styles.AccentStyle.Render(m.statusText)
		}
		return styles.StatusBar.Render(helpText) + "  " + statusStyled
	}

	return styles.StatusBar.Render(helpText)
}

// Helper methods

// truncateRunes removes count runes from the end of string, preserving UTF-8 characters
func truncateRunes(s string, count int) string {
	runes := []rune(s)
	if count >= len(runes) {
		return ""
	}
	return string(runes[:len(runes)-count])
}

func (m Model) selectedTask() *model.Task {
	if m.activeView == ViewBoard {
		columns := m.tasksByStatus()
		return m.selectedBoardTask(columns)
	}
	// List view: check which section cursor is in
	if m.inDoneSection {
		if m.doneCursor >= 0 && m.doneCursor < len(m.doneTasksList) {
			return &m.doneTasksList[m.doneCursor]
		}
	} else {
		if m.cursor >= 0 && m.cursor < len(m.activeTasks) {
			return &m.activeTasks[m.cursor]
		}
	}
	return nil
}

func (m Model) selectedBoardTask(columns [3][]model.Task) *model.Task {
	col := columns[m.boardCol]
	cursor := m.boardCursors[m.boardCol]
	if cursor >= 0 && cursor < len(col) {
		return &col[cursor]
	}
	return nil
}

func (m Model) tasksByStatus() [3][]model.Task {
	var columns [3][]model.Task
	for _, task := range m.tasks {
		switch task.Status {
		case model.StatusTodo:
			columns[0] = append(columns[0], task)
		case model.StatusDoing:
			columns[1] = append(columns[1], task)
		case model.StatusDone:
			columns[2] = append(columns[2], task)
		}
	}
	return columns
}

func (m *Model) clampCursor() {
	// For list view, use activeTasks; board view clamps separately
	maxLen := len(m.activeTasks)
	if m.activeView == ViewBoard {
		maxLen = len(m.tasks)
	}
	if m.cursor >= maxLen {
		m.cursor = maxLen - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) clampDoneCursor() {
	if m.doneCursor >= len(m.doneTasksList) {
		m.doneCursor = len(m.doneTasksList) - 1
	}
	if m.doneCursor < 0 {
		m.doneCursor = 0
	}
}

func (m *Model) clampBoardCursor(columns [3][]model.Task) {
	col := columns[m.boardCol]
	if m.boardCursors[m.boardCol] >= len(col) {
		m.boardCursors[m.boardCol] = len(col) - 1
	}
	if m.boardCursors[m.boardCol] < 0 {
		m.boardCursors[m.boardCol] = 0
	}
}

func (m *Model) updateBoardScroll() {
	// Calculate maxItems same as renderColumnWithHeight
	height := m.height - 10
	maxItems := (height - 6) / 4
	if maxItems < 2 {
		maxItems = 2
	}

	col := m.boardCol
	cursor := m.boardCursors[col]
	start := m.boardScrolls[col]

	// Only scroll when cursor goes out of visible area
	if cursor < start {
		m.boardScrolls[col] = cursor
	} else if cursor >= start+maxItems {
		m.boardScrolls[col] = cursor - maxItems + 1
	}
}

func (m Model) reloadTasks() tea.Cmd {
	filter := store.TaskFilter{
		ProjectID: m.currentProject,
		Search:    m.searchQuery,
	}
	return loadTasks(m.store, filter)
}
