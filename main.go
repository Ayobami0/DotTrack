package main

import (
	"os"

	"github.com/Ayobami0/dottrack/models"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fp, err := tea.LogToFile("./debug.log", "debug")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	m := models.RegisteredModels[models.Display]
	program := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		os.Exit(1)
	}

}
