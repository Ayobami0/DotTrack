package models

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

var home, _ = os.UserHomeDir()

const (
	defaultJSONFile = "dots.json"
)

func getDefaultZIPPath() string { return fmt.Sprintf("%s/dotfiles.zip", home) }

type DotFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (d DotFile) FilterValue() string {
	return d.Name
}

func (d DotFile) Title() string {
	return d.Name
}

func (d DotFile) Description() string {
	return d.Path
}

func loadJSON() []DotFile {
	var dots []DotFile
	j, err := os.ReadFile(defaultJSONFile)
	if err != nil {
		if os.IsNotExist(err) {
			fs, _ := os.Create(defaultJSONFile)
			fs.Close()
			return dots
		}
		panic(err)
	}
	err = json.Unmarshal(j, &dots)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return dots
		}
		panic(err)
	}
	return dots
}

func saveJSON(dotFiles []DotFile) error {
	dots := loadJSON()
	for _, d := range dotFiles {
		if !slices.Contains(dots, d) {
			dots = append(dots, d)
		}
	}
	data, err := json.Marshal(dots)
	if err != nil {
		return err
	}
	err = os.WriteFile(defaultJSONFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func removeJSON(dotName string) error {
	dots := loadJSON()
	if !slices.ContainsFunc(dots, func(d DotFile) bool {
		return d.Name == dotName
	}) {
		return fmt.Errorf("No Such dotfile")
	}
	dots = slices.DeleteFunc(dots, func(d DotFile) bool {
		return d.Name == dotName
	})
	data, err := json.Marshal(dots)
	if err != nil {
		return err
	}
	err = os.WriteFile(defaultJSONFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func zipDotFiles() (tea.Cmd, error) {
	dots := loadJSON()

	tmpPath := os.TempDir()
	tmpDir, err := os.MkdirTemp(tmpPath, "dotzip-*")
	if err != nil {
		return nil, err
	}
	for _, d := range dots {
		err = os.Symlink(d.Path, fmt.Sprintf("%s/%s", tmpDir, d.Name))
		if err != nil {
			return nil, err
		}
	}
	zip := exec.Command("zip", "-r", getDefaultZIPPath(), tmpDir)
	zip.Env = os.Environ()

	return tea.ExecProcess(zip, func(err error) tea.Msg {
		return compressionFinishedMsg{err}
	}), nil

}
