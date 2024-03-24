package models

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type DisplayModel struct {
	list            list.Model
	err             error
	altscreenActive bool
}

var displayShortKeys = []key.Binding{}
var displayLongKeys = []key.Binding{
	key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add dotfile"),
	),
	key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "remove dotfile"),
	),
	key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "zip dotfiles"),
	),
}

func (m *DisplayModel) initList(w, h int) {
	var dotItems []list.Item
	for _, v := range loadJSON() {
		dotItems = append(dotItems, v)
	}
	m.list = list.New(dotItems, list.NewDefaultDelegate(), w, h)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return displayShortKeys
	}
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return displayLongKeys
	}
	m.list.Title = "Dots..."
}

func (m DisplayModel) Init() tea.Cmd {
	return nil
}

func (m DisplayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			return RegisteredModels[Form].Update(
				tea.WindowSizeMsg{Width: m.list.Width(), Height: m.list.Height()})
		case "x":
			itemSelected := m.list.Index()
			if err := removeJSON(m.list.SelectedItem().FilterValue()); err == nil {
				m.list.RemoveItem(itemSelected)
			}
		case "z":
			cmd, _ := zipDotFiles()
			return m, cmd

		}
	case compressionFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.ExitAltScreen
		}
	}
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

type compressionFinishedMsg struct{ err error }

func (m DisplayModel) View() string {
	return m.list.View()
}

func NewDisplayModel() *DisplayModel {
	return &DisplayModel{}
}
