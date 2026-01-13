package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hwanchang/tsk/internal/model"
	"github.com/hwanchang/tsk/internal/store"
)

func loadTasks(st *store.SQLiteStore, filter store.TaskFilter) tea.Cmd {
	return func() tea.Msg {
		tasks, err := st.ListTasks(filter)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return TasksLoadedMsg{Tasks: tasks}
	}
}

func loadProjects(st *store.SQLiteStore) tea.Cmd {
	return func() tea.Msg {
		projects, err := st.ListProjects()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return ProjectsLoadedMsg{Projects: projects}
	}
}

func loadTags(st *store.SQLiteStore) tea.Cmd {
	return func() tea.Msg {
		tags, err := st.ListTags()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return TagsLoadedMsg{Tags: tags}
	}
}

func createTask(st *store.SQLiteStore, title string, projectID *int64) tea.Cmd {
	return func() tea.Msg {
		task := model.NewTask(title)
		task.ProjectID = projectID
		if err := st.CreateTask(task); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskCreatedMsg{Task: task}
	}
}

func updateTask(st *store.SQLiteStore, task *model.Task) tea.Cmd {
	return func() tea.Msg {
		if err := st.UpdateTask(task); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskUpdatedMsg{Task: task}
	}
}

func completeTask(st *store.SQLiteStore, taskID int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.CompleteTaskWithRecurrence(taskID); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskUpdatedMsg{Task: nil}
	}
}

func deleteTask(st *store.SQLiteStore, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.DeleteTask(id); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskDeletedMsg{ID: id}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

func createProject(st *store.SQLiteStore, name string) tea.Cmd {
	return func() tea.Msg {
		project := model.NewProject(name)
		if err := st.CreateProject(project); err != nil {
			return ErrorMsg{Err: err}
		}
		return ProjectCreatedMsg{Project: project}
	}
}

func createProjectWithDesc(st *store.SQLiteStore, name, description string) tea.Cmd {
	return func() tea.Msg {
		project := model.NewProject(name)
		project.Description = description
		if err := st.CreateProject(project); err != nil {
			return ErrorMsg{Err: err}
		}
		return ProjectCreatedMsg{Project: project}
	}
}

func addTagToTask(st *store.SQLiteStore, taskID, tagID int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.AddTagToTask(taskID, tagID); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskUpdatedMsg{Task: nil}
	}
}

func removeTagFromTask(st *store.SQLiteStore, taskID, tagID int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.RemoveTagFromTask(taskID, tagID); err != nil {
			return ErrorMsg{Err: err}
		}
		return TaskUpdatedMsg{Task: nil}
	}
}

func createTag(st *store.SQLiteStore, name string) tea.Cmd {
	return func() tea.Msg {
		tag := model.NewTag(name)
		if err := st.CreateTag(tag); err != nil {
			return ErrorMsg{Err: err}
		}
		return TagCreatedMsg{Tag: tag}
	}
}

func deleteTag(st *store.SQLiteStore, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.DeleteTag(id); err != nil {
			return ErrorMsg{Err: err}
		}
		return TagDeletedMsg{ID: id}
	}
}

func deleteProject(st *store.SQLiteStore, id int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.DeleteProject(id); err != nil {
			return ErrorMsg{Err: err}
		}
		return ProjectDeletedMsg{ID: id}
	}
}

func setRecurrence(st *store.SQLiteStore, taskID int64, pattern model.RecurrencePattern, interval int) tea.Cmd {
	return func() tea.Msg {
		rec := model.NewRecurrence(taskID, pattern, interval)
		// Get task to determine next due date
		task, err := st.GetTask(taskID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		if task.DueDate != nil {
			rec.NextDue = rec.CalculateNextDue(*task.DueDate)
		} else {
			rec.NextDue = rec.CalculateNextDue(time.Now())
		}
		if err := st.SetRecurrence(rec); err != nil {
			return ErrorMsg{Err: err}
		}
		return RecurrenceSetMsg{TaskID: taskID}
	}
}

func deleteRecurrence(st *store.SQLiteStore, taskID int64) tea.Cmd {
	return func() tea.Msg {
		if err := st.DeleteRecurrence(taskID); err != nil {
			return ErrorMsg{Err: err}
		}
		return RecurrenceDeletedMsg{TaskID: taskID}
	}
}
