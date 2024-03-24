package models

import tea "github.com/charmbracelet/bubbletea"

// Screen Indexes
const (
	Display = iota
	Form
)

var RegisteredModels = map[int]tea.Model{
	Display: NewDisplayModel(),
	Form:    NewFormModel(),
}
