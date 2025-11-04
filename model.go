package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	title, desc, content string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list     list.Model
	form     *huh.Form
	ready    bool
	width    int
	height   int
	mode     string
	selected item
	content  string
	vp       viewport.Model
	items    []list.Item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			switch m.mode {
			case "form_add", "post":
				m.mode = "list"
				return m, nil
			}
		case "ctrl+a":
			if m.mode == "list" {
				m.selected = item{}
				m.mode = "form_add"
				m.initform()
				return m, m.form.Init()
			}
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			m.width = msg.Width
			m.height = msg.Height
			listWidth := 60
			listHeight := 15
			m.list.SetSize(listWidth, listHeight)

			m.vp = viewport.New(msg.Width, msg.Height)
			m.ready = true
		}
	}

	switch m.mode {
	case "list":
		return m.updatelist(msg)
	case "form_add":
		return m.updateform(msg)
	case "post":
		return m.updateviewport(msg)
	}

	return m, nil
}

func (m model) updatelist(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			m.mode = "post"
			if selected, ok := m.list.SelectedItem().(item); ok {
				p := new(post)
				db.Where("title", selected.title).First(p)
				m.content, _ = glamour.Render(p.Content, "dark")
				m.vp.SetContent(m.content)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateform(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		title := m.form.GetString("title")
		description := m.form.GetString("description")
		content := m.form.GetString("content")

		m.save(title, description, content)
		newitem := item{title: title, desc: description, content: content}
		m.items = append(m.items, newitem)
		m.list.SetItems(m.items)
		m.mode = "list"
	}

	return m, cmd
}

func (m model) updateviewport(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *model) initform() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("Enter title...").
				Value(&m.selected.title).
				Key("title").
				Validate(func(str string) error {
					if len(str) == 0 {
						return fmt.Errorf("title cannot be empty")
					}

					if len(str) > 20 {
						return fmt.Errorf("title is too long")
					}

					var existingPost post
					if err := db.Where("title = ?", str).First(&existingPost).Error; err == nil {
						return fmt.Errorf("title already exists")
					}

					return nil
				}),

			huh.NewInput().
				Title("Description").
				Placeholder("Enter description...").
				Value(&m.selected.desc).
				Key("description").
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("description cannot be empty")
					}

					if len(s) > 30 {
						return fmt.Errorf("description is too long")
					}

					return nil
				}),

			huh.NewText().
				Title("Content").
				Placeholder("Enter content...").
				Key("content").
				Lines(10).
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("content cannot be empty")
					}

					if len(s) > 500 {
						return fmt.Errorf("content is too long")
					}

					return nil
				}),
		),
	)
}

func (m model) save(title string, desc string, content string) error {
	result := db.Create(&post{
		Title:   title,
		Desc:    desc,
		Content: content,
	})
	return result.Error
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var content string

	switch m.mode {
	case "list":
		content = m.list.View()
	case "post":
		content = m.vp.View()
	case "form_add":
		formStyle := lipgloss.NewStyle().
			Width(60).
			MaxWidth(60)
		content = formStyle.Render(m.form.View())
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
