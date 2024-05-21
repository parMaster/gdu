package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/jessevdk/go-flags"
	"github.com/parmaster/gdu/fs"
	"github.com/rivo/tview"
)

var Options struct {
	// Dir        bool `long:"dir" short:"d" description:"show help message"`
	Help       bool `long:"help" short:"h" description:"show help message"`
	Verbose    bool `long:"verbose" short:"v" env:"VERBOSE" description:"verbose output (default: false)"`
	Positional struct {
		Dir string
	} `positional-args:"yes"`
}

func fillTable(table *tview.Table, text string) {
	for i := 0; i < 10; i++ {
		table.SetCell(i+1, 0,
			tview.NewTableCell("░░░░░█████").
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
		table.SetCell(i+1, 1,
			tview.NewTableCell(fmt.Sprintf("%d GB", i)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignRight))
		table.SetCell(i+1, 2,
			tview.NewTableCell(fmt.Sprintf("%d %v", i, text)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft))
	}
}

func main() {

	if _, err := flags.Parse(&Options); err != nil {
		os.Exit(1)
	}

	var dir string
	var err error
	if len(Options.Positional.Dir) != 0 {
		dir = Options.Positional.Dir
	} else {
		dir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("dir:", dir)

	fs := fs.NewFS(dir)
	fs.Scan()

	app := tview.NewApplication()

	table := tview.NewTable()
	table.SetBorders(false)

	table.SetSelectable(true, false)

	table.SetCell(0, 0,
		tview.NewTableCell("").
			SetTextColor(tcell.ColorYellow).
			// SetBackgroundColor(tcell.ColorSkyblue).
			SetAlign(tview.AlignLeft).SetSelectable(false))
	table.SetCell(0, 1,
		tview.NewTableCell("   Size   ").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).SetSelectable(false).SetMaxWidth(10)).SetSeparator(tview.BoxDrawingsLightVertical)
	table.SetCell(0, 2,
		tview.NewTableCell("Name").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetExpansion(1))

	header := tview.NewTextView()

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, false)
		}
	}).SetSelectedFunc(func(row int, column int) {
		table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		// table.SetSelectable(false, false)

		selectedText := table.GetCell(row, 2).Text

		header.SetText(fmt.Sprintf("Header Selected %d %d, %v", row, column, selectedText))

		fillTable(table, "refilled"+fmt.Sprintf(" %d %d", row, column))

		table.Select(0, 0)
	})

	fillTable(table, "initial")

	grid := tview.NewGrid().SetColumns(0).SetRows(1, 0, 1)
	grid.AddItem(header, 0, 0, 1, 1, 10, 0, false)
	header.SetText("Header").SetTextColor(tcell.ColorYellow)

	grid.AddItem(table, 1, 0, 1, 1, 0, 0, true)

	// status bar
	status := tview.NewTextView()
	status.SetText("Status Bar")
	grid.AddItem(status, 2, 0, 1, 1, 0, 0, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		log.Fatal(err)
	}
}
