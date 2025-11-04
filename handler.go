package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

func handler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	items := []list.Item{}

	posts := []post{}
	if err := db.Find(&posts).Error; err != nil {
		s.Exit(1)
	}

	for _, p := range posts {
		items = append(items, item{title: p.Title, desc: p.Desc})
	}

	m := &model{
		list:  list.New(items, list.NewDefaultDelegate(), 60, 15),
		mode:  "list",
		items: items,
	}

	m.list.Title = "Posts"
	m.list.SetItems(items)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
