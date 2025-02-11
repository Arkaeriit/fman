package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/76creates/stickers"
	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/nore-dev/fman/message"

	"github.com/nore-dev/fman/model/entryinfo"
	"github.com/nore-dev/fman/model/infobar"
	"github.com/nore-dev/fman/model/list"
	"github.com/nore-dev/fman/model/toolbar"

	"github.com/nore-dev/fman/theme"
)

type App struct {
	list      list.List
	entryInfo entryinfo.EntryInfo
	toolbar   toolbar.Toolbar
	infobar   infobar.Infobar

	flexBox *stickers.FlexBox
}

func (app *App) Init() tea.Cmd {
	return tea.Batch(app.infobar.Init(), app.UpdatePath(), app.list.Init())
}

func (app *App) UpdatePath() tea.Cmd {
	return func() tea.Msg {
		path := args.Path

		absolutePath, _ := filepath.Abs(path)
		return message.PathMsg{Path: absolutePath}
	}
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return app, tea.Quit
		}

	case tea.WindowSizeMsg:
		app.flexBox.SetHeight(msg.Height - lipgloss.Height(app.toolbar.View()) - lipgloss.Height(app.toolbar.View()))
		app.flexBox.SetWidth(msg.Width)

		app.flexBox.ForceRecalculate()

		app.list.SetWidth(app.flexBox.Row(0).Cell(0).GetWidth())
		app.list.SetHeight(app.flexBox.GetHeight())

		app.entryInfo.SetWidth(app.flexBox.Row(0).Cell(1).GetWidth())

	}

	var listCmd, toolbarCmd, entryCmd, infobarCmd tea.Cmd

	app.list, listCmd = app.list.Update(msg)
	app.toolbar, toolbarCmd = app.toolbar.Update(msg)
	app.entryInfo, entryCmd = app.entryInfo.Update(msg)
	app.infobar, infobarCmd = app.infobar.Update(msg)

	return app, tea.Batch(listCmd, toolbarCmd, entryCmd, infobarCmd)
}

func (app *App) View() string {
	app.flexBox.ForceRecalculate()

	row := app.flexBox.Row(0)

	// Set content of list view
	row.Cell(0).SetContent(app.list.View())

	// Set content of entry view
	row.Cell(1).SetContent(app.entryInfo.View())

	return zone.Scan(lipgloss.JoinVertical(
		lipgloss.Top,
		app.toolbar.View(),
		zone.Mark("list", app.flexBox.Render()),
		app.infobar.View(),
	))
}

// Define CLI arguments
var args struct {
	Path  string `arg:"positional" default:"."`
	Theme string `default:"default"`
}

func main() {
	// Initialize Bubblezone
	zone.NewGlobal()
	defer zone.Close()

	arg.MustParse(&args)

	selectedTheme := theme.GetActiveTheme(args.Theme)

	theme.SetTheme(selectedTheme)

	app := App{
		list:      list.New(&selectedTheme),
		entryInfo: entryinfo.New(),
		toolbar:   toolbar.New(),
		infobar:   infobar.New(),
		flexBox:   stickers.NewFlexBox(0, 0),
	}

	rows := []*stickers.FlexBoxRow{
		app.flexBox.NewRow().AddCells(
			[]*stickers.FlexBoxCell{
				stickers.NewFlexBoxCell(7, 1).SetStyle(theme.ListStyle),      // List
				stickers.NewFlexBoxCell(3, 1).SetStyle(theme.ContainerStyle), // Entry Information
			},
		),
	}

	app.flexBox.AddRows(rows)

	p := tea.NewProgram(&app, tea.WithAltScreen(), tea.WithMouseAllMotion())

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
