package gncdu

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Page interface {
	SetNavigator(nav *Navigator)
	SetPrevious(previous Page)
	Previous() Page
	Show()
	Dispose()
}

type BasePage struct {
	app       *tview.Application
	previous  Page
	navigator *Navigator
}

func (p *BasePage) SetNavigator(nav *Navigator) {
	p.navigator = nav
}

func (p *BasePage) SetPrevious(previous Page) {
	p.previous = previous
}

func (p *BasePage) Previous() Page {
	return p.previous
}

func (p *BasePage) Dispose() {
}

type ScanningPage struct {
	BasePage
	done chan bool
}

func NewScanningPage(app *tview.Application) *ScanningPage {
	done := make(chan bool)
	return &ScanningPage{BasePage: BasePage{app: app}, done: done}
}

func (page *ScanningPage) Show() {
	modal := tview.NewModal().
		SetText("Scanning       \n\nTime 0s")

	info := tview.NewTextView().
		SetText("[ctrl+c] close")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(modal, 0, 1, true).
		AddItem(info, 1, 1, false)

	go func() {
		tick := time.Tick(time.Millisecond * 500)
		dots := []byte{'.', '.', '.', '.', '.', '.'}
		spaces := []byte{' ', ' ', ' ', ' ', ' ', ' '}
		count := 0
		for {
			select {
			case <-tick:
				count++
				p := count % 7
				s := string(dots[0:p])
				b := string(spaces[0:(6 - p)])
				page.app.QueueUpdateDraw(func() {
					modal.SetText(fmt.Sprintf("Scanning %s%s\n\nTime %ds", s, b, count/2))
				})
			case <-page.done:
				return
			}
		}
	}()

	page.app.SetRoot(layout, true).SetFocus(layout)
}

func (p *ScanningPage) Dispose() {
	close(p.done)
}

type ResultPage struct {
	BasePage
	files  []*FileData
	parent *FileData
}

func NewResultPage(app *tview.Application, files []*FileData, parent *FileData) *ResultPage {
	return &ResultPage{
		BasePage: BasePage{app: app},
		files:    files,
		parent:   parent,
	}
}

func (p *ResultPage) Show() {
	sort.Slice(p.files, func(i, j int) bool {
		return p.files[i].Size() > p.files[j].Size()
	})

	offset := 1
	var title string
	if p.parent != nil {
		offset = 2
		title = p.parent.Path()
	}

	table := tview.NewTable().
		SetFixed(1, 1).
		SetSelectable(true, false).
		SetSelectedFunc(func(row, column int) {
			if row == 0 {
				return
			}

			if row == offset-1 {
				if p.parent.parent.root() {
					page := NewResultPage(p.app, p.parent.parent.Children, nil)
					navigator.Push(page)
				} else {
					page := NewResultPage(p.app, p.parent.parent.Children, p.parent.parent)
					navigator.Push(page)
				}
				return
			}

			file := p.files[row-offset]
			if !file.info.IsDir() {
				return
			}
			page := NewResultPage(p.app, file.Children, file)
			navigator.Push(page)
		})

	color := tcell.ColorYellow
	table.SetCell(0, 0, tview.NewTableCell("Name").SetTextColor(color).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("Size").SetTextColor(color).SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("Items").SetTextColor(color).SetSelectable(false))

	if p.parent != nil {
		table.SetCellSimple(1, 0, "...")
	}

	for i, file := range p.files {
		table.SetCellSimple(i+offset, 0, file.info.Name())
		table.SetCell(i+offset, 1,
			tview.NewTableCell(ToHumanSize(file.Size())).
				SetAlign(tview.AlignRight))
		table.SetCell(i+offset, 2,
			tview.NewTableCell(strconv.Itoa((file.Count()))).
				SetAlign(tview.AlignRight))
	}

	layout := newLayout(title, table)
	p.app.SetRoot(layout, true).SetFocus(layout)
}

type HelpPage struct {
	BasePage
}

func NewHelpPage(app *tview.Application) *HelpPage {
	return &HelpPage{BasePage: BasePage{app: app}}
}

func (p *HelpPage) Show() {
	modal := tview.NewModal().
		SetText("Help").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(i int, l string) {
			if i == 0 {
				p.navigator.Pop()
			}
		})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(modal, 0, 1, true).
		AddItem(newInfoView(), 1, 1, false)

	p.app.SetRoot(layout, true).SetFocus(layout)
}