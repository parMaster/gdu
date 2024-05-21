package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/jessevdk/go-flags"
	"github.com/parmaster/gdu/fs"
	"github.com/rivo/tview"
)

var Options struct {
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

type bars struct {
	ascii string
	color tcell.Color
}

// colors from light yellow to dark red
var bar map[int]bars = map[int]bars{
	0:  {"          ", tcell.ColorLightYellow},
	1:  {"         ▒", tcell.ColorLightYellow},
	2:  {"        ▒▒", tcell.ColorYellow},
	3:  {"       ▒▒▒", tcell.ColorYellow},
	4:  {"      ▒▒▒▒", tcell.ColorYellow},
	5:  {"     ▒▒▒▒▒", tcell.ColorPaleVioletRed},
	6:  {"    ▒▒▒▒▒▒", tcell.ColorPaleVioletRed},
	7:  {"   ▒▒▒▒▒▒▒", tcell.ColorPaleVioletRed},
	8:  {"  ▒▒▒▒▒▒▒▒", tcell.ColorRed},
	9:  {" ▒▒▒▒▒▒▒▒▒", tcell.ColorDarkRed},
	10: {"▒▒▒▒▒▒▒▒▒▒", tcell.ColorMediumVioletRed},
} // todo: correct colors

func (a *App) Update(list fs.ListView) {
	a.table.Clear()
	a.UpdateTableHeader()
	for i, item := range list.Items {
		color := tcell.ColorWhite
		if item.IsDir {
			color = tcell.ColorGreen
		}
		// Color bar
		barSize := int(math.Ceil(10 * (float64(item.Size) / float64(list.TotalSize))))
		a.table.SetCell(i+1, 0,
			tview.NewTableCell(bar[barSize].ascii).
				SetTextColor(bar[barSize].color).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
		// Size
		size := humanize.Bytes(item.Size)
		if item.Name == ".." {
			size = ".."
		}
		a.table.SetCell(i+1, 1,
			tview.NewTableCell(size).
				SetTextColor(color).
				SetAlign(tview.AlignRight))

		// Name
		// Todo: show icon - file or dir
		var icon string
		if item.IsDir {
			icon = "📁"
		} else {
			icon = "📄"
		}
		a.table.SetCell(i+1, 2,
			tview.NewTableCell(fmt.Sprintf("%s %s", icon, item.Name)).
				SetTextColor(color).
				SetAlign(tview.AlignLeft))
	}
	a.table.ScrollToBeginning()
	a.header.SetText(a.fs.CurrentDir)
}

func (a *App) Run() {
	a.view = tview.NewApplication()

	a.table = tview.NewTable()
	a.table.SetBorders(false)

	a.table.SetSelectable(true, false)

	posRow, posCol := a.table.GetSelection()
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

		newDir := a.table.GetCell(row, 2).Text
		newDir = strings.TrimPrefix(newDir, "📁 ")
		newDir = strings.TrimPrefix(newDir, "📄 ")

		if newDir != ".." {
			// remember cursor position
			posRow, posCol = a.table.GetSelection()
		}

		a.fs.ChangeDir(newDir)

		list := a.fs.List()
		a.Update(*list)
		// todo: restore cursor position through the whole history
		// this is a simple one-level history
		if newDir == ".." {
			a.table.Select(posRow, posCol)
			posRow, posCol = 0, 0
		} else {
			a.table.Select(0, 0)
		}
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
			log.Printf("%e", err)
			return
		}
	}
	dir = filepath.Clean(dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Directory does not exist: %v", dir)
		return
	}

	fmt.Println("Reading directory content:", dir) // todo: add progress indicator

	app := NewApp(dir)
	app.fs.Scan()
	app.Run()
}
