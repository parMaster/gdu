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

type App struct {
	fs     *fs.FS
	view   *tview.Application
	table  *tview.Table
	header *tview.TextView
}

func NewApp(dir string) *App {
	return &App{
		fs: fs.NewFS(dir),
	}
}

func (a *App) Update(list fs.ListView) {
	a.table.Clear()
	a.UpdateTableHeader()
	for i, item := range list.Items {
		color := tcell.ColorWhite
		if item.IsDir {
			color = tcell.ColorGreen
		}
		// Todo: show real bar, and colorize it
		a.table.SetCell(i+1, 0,
			tview.NewTableCell("░░░░░█████").
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
		// Todo: humanize
		a.table.SetCell(i+1, 1,
			tview.NewTableCell(fmt.Sprintf("%d b", item.Size)).
				SetTextColor(color).
				SetAlign(tview.AlignRight))

		// Name
		// Todo: show icon - file or dir
		a.table.SetCell(i+1, 2,
			tview.NewTableCell(fmt.Sprintf("%v", item.Name)).
				SetTextColor(color).
				SetAlign(tview.AlignLeft))
	}
	a.header.SetText(a.fs.CurrentDir)
}

func (a *App) Run() {
	a.view = tview.NewApplication()

	a.table = tview.NewTable()
	a.table.SetBorders(false)

	a.table.SetSelectable(true, false)

	a.header = tview.NewTextView()

	a.table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			a.view.Stop()
		}
		if key == tcell.KeyEnter {
			a.table.SetSelectable(true, false)
		}
	}).SetSelectedFunc(func(row int, column int) {
		a.table.GetCell(row, column).SetTextColor(tcell.ColorRed)

		selectedText := a.table.GetCell(row, 2).Text
		// fmt.Println("ChangeDir:", selectedText)

		a.fs.ChangeDir(selectedText)

		list := a.fs.List()
		a.Update(*list)

		a.table.Select(0, 0)
	})

	grid := tview.NewGrid().SetColumns(0).SetRows(1, 0, 1)
	grid.AddItem(a.header, 0, 0, 1, 1, 10, 0, false)
	a.header.SetText("Header").SetTextColor(tcell.ColorYellow)

	grid.AddItem(a.table, 1, 0, 1, 1, 0, 0, true)

	// status bar
	status := tview.NewTextView()
	status.SetText("Status Bar")
	grid.AddItem(status, 2, 0, 1, 1, 0, 0, false)

	list := a.fs.List()
	a.Update(*list)

	if err := a.view.SetRoot(grid, true).Run(); err != nil {
		log.Fatal(err)
	}

}

func (a *App) UpdateTableHeader() {
	a.table.SetCell(0, 0,
		tview.NewTableCell("").
			SetTextColor(tcell.ColorYellow).
			// SetBackgroundColor(tcell.ColorSkyblue).
			SetAlign(tview.AlignLeft).SetSelectable(false))
	a.table.SetCell(0, 1,
		tview.NewTableCell("   Size   ").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).SetSelectable(false).SetMaxWidth(10)).SetSeparator(tview.BoxDrawingsLightVertical)
	a.table.SetCell(0, 2,
		tview.NewTableCell("Name").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetExpansion(1))
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

	app := NewApp(dir)
	app.fs.Scan()
	app.Run()
}
