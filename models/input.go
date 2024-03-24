package models

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Help   key.Binding
	Quit   key.Binding
	Delete key.Binding
	Save   key.Binding
	Close  key.Binding
	Tab    key.Binding
	STab   key.Binding
	Next   key.Binding
	Prev   key.Binding
	Comp   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Close, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Save},
		{k.Next, k.Prev, k.Comp},
		{k.Tab, k.STab},
		{k.Help, k.Close, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit program"),
	),
	Close: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "cancel addition"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save dotfiles"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "delete entry"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next page"),
	),
	STab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev page"),
	),
	Next: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "next suggestion"),
	),
	Prev: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "previous suggestion"),
	),
	Comp: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "complete suggestion"),
	),
}

var winSize = struct {
	h int
	w int
}{}

var suggestions = map[string]string{}

type FormModel struct {
	SelectionList list.Model
	LastIndex     int
	ConfigDir     string
	Input         textinput.Model
	Help          help.Model
}

func (f FormModel) Init() tea.Cmd {
	return nil
}

func (f FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	f.Input.Placeholder = ""
	f.Input.PlaceholderStyle = lipgloss.NewStyle().Foreground(PLACEHOLDER_DEFAULT)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		winSize.h = msg.Height
		winSize.w = msg.Width
		f.SelectionList = list.New([]list.Item{}, list.NewDefaultDelegate(), winSize.w/2, winSize.h/2)
		f.SelectionList.Help = help.New()
		f.SelectionList.SetShowHelp(false)
		f.SelectionList.DisableQuitKeybindings()
		f.SelectionList.SetShowFilter(false)
		f.SelectionList.KeyMap.CursorUp = keys.Up
		f.SelectionList.KeyMap.CursorDown = keys.Down
		f.SelectionList.KeyMap.NextPage = keys.Tab
		f.SelectionList.KeyMap.PrevPage = keys.STab
		f.SelectionList.Title = "Configs"
		f.SelectionList.SetShowStatusBar(false)
		f.SelectionList.SetShowTitle(false)
		f.Help.Width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return f, tea.Quit
		case "ctrl+x":
			f.SelectionList.RemoveItem(f.SelectionList.Index())
		case "ctrl+s":
			var dots []DotFile
			for _, v := range f.SelectionList.Items() {
				item, ok := v.(DotFile)
				log.Println(ok)
				if !ok {
					goto update
				}
				dots = append(dots, item)
			}
			err := saveJSON(dots)
			if err != nil {
				goto update
			}
			log.Println(dots)
			return RegisteredModels[Display].Update(
				tea.WindowSizeMsg{
					Width:  winSize.w,
					Height: winSize.h})
		case "ctrl+q":
			return RegisteredModels[Display].Update(
				tea.WindowSizeMsg{
					Width:  winSize.w,
					Height: winSize.h})
		case "ctrl+h":
			f.Help.ShowAll = !f.Help.ShowAll
		case "enter":
			if f.Input.Focused() {
				dotName := f.Input.Value()
				if name, ok := suggestions[strings.ToLower(dotName)]; ok {
					if slices.ContainsFunc(f.SelectionList.Items(), func(i list.Item) bool {
						return i.FilterValue() == name
					}) {
						f.Input.Reset()
						f.Input.Placeholder = "Duplicate Entry"
						f.Input.PlaceholderStyle = lipgloss.NewStyle().Foreground(PLACEHOLDER_ERROR)
						goto update

					}

					f.SelectionList.InsertItem(
						f.LastIndex,
						DotFile{
							Name: name,
							Path: fmt.Sprintf("%s/%s", f.ConfigDir, name)})
					f.Input.Reset()
					f.LastIndex++
				} else {
					f.Input.Reset()
					f.Input.Placeholder = "Invalid Dotfile"
					f.Input.PlaceholderStyle = lipgloss.NewStyle().Foreground(PLACEHOLDER_ERROR)
				}
			}
		default:
			if !f.Input.Focused() {
				if reg, _ := regexp.MatchString("[0-9a-zA-Z.]", msg.String()); reg {
					f.Input.Focus()
				}
			}
		}
	}
update:
	f.SelectionList, cmd = f.SelectionList.Update(msg)
	f.Input, cmd = f.Input.Update(msg)
	f.Help, cmd = f.Help.Update(msg)
	return f, cmd
}

var dialogBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(VIOLET).
	Padding(1, 0).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)

func (f FormModel) View() string {
	inputBox := lipgloss.NewStyle().
		Width(winSize.w/2).
		Margin(0, 1, 2, 1).
		Render(f.Input.View())
	listBox := lipgloss.NewStyle().
		Margin(0, 1).
		Width(winSize.w / 2).
		Height(winSize.h / 2).
		Render(f.SelectionList.View())

	helpView := lipgloss.NewStyle().Margin(0, 1).
		Render(f.Help.View(keys))

	ui := lipgloss.JoinVertical(lipgloss.Left, inputBox, listBox, helpView)
	return lipgloss.Place(winSize.w, winSize.h, lipgloss.Center, lipgloss.Center, dialogBoxStyle.Render(ui))
}

func NewFormModel() *FormModel {

	input := textinput.New()
	input.Focus()
	input.Prompt = "SEARCH: "

	input.PromptStyle = lipgloss.NewStyle().Foreground(VIOLET)
	input.CharLimit = 15
	var suggestion []string
	input.KeyMap.NextSuggestion = keys.Next
	input.KeyMap.PrevSuggestion = keys.Prev
	input.KeyMap.AcceptSuggestion = keys.Comp
	config, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	dots, err := os.ReadDir(config)
	if err != nil {
		panic(err)
	}

	for _, v := range dots {
		suggestion = append(suggestion, strings.ToLower(v.Name()))
		suggestions[strings.ToLower(v.Name())] = v.Name()
	}
	input.SetSuggestions(suggestion)
	input.ShowSuggestions = true

	help := help.New()
	return &FormModel{
		Input:     input,
		ConfigDir: config,
		Help:      help,
	}
}
