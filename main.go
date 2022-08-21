package main

import (
	"fmt"
	"os"

	ms "github.com/ybalexdp/go-minesweeper/minesweeper"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	m, err := ms.InitialModel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	err = p.Start()
	if err != nil {
		fmt.Println(err)
	}

}
