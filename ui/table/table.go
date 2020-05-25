package table

import (
	"github.com/gdamore/tcell"
	"github.com/jwdevantier/spellbook/ui"
	"github.com/rivo/tview"
)

type Row interface {
	// Return unique id for row
	Id() uint64
	// Number of columns
	Len() int
	// value in cell
	CellValue(col int) interface{}
}

type Renderer interface {
	Render(Row) []string
}

type Filter interface {
	Filter([]Row) []Row
}

type Model struct {
	rows []Row
	// maps from row id to index in rows
	lookup map[uint64]int
	onChanged func()
}

func (tm *Model) SetContents(rows []Row){
	tm.rows = make([]Row, len(rows))
	tm.lookup = make(map[uint64]int)
	for i, row := range rows {
		tm.rows[i] = row
		tm.lookup[row.Id()] = i
	}

	if tm.onChanged != nil {
		tm.onChanged()
	}
}

func (tm *Model) Contents() []Row {
	return tm.rows
}

func (tm *Model) SetOnChanged(handler func()) {
	tm.onChanged = handler
}

func (tm *Model) LookUp(id uint64) (Row, bool) {
	ndx, ok := tm.lookup[id]
	if !ok {
		return nil, false
	}
	return tm.rows[ndx], true
}

func NewTableModel(rows []Row) *Model {
	m := &Model{}
	m.SetContents(rows)
	return m
}

// Table
///////////////////////////////
type Table struct {
	model    *Model
	filter   Filter
	view     *tview.Table
	renderer Renderer
	onTab    func()
	onEsc    func()

	rowIndex map[int]uint64
}

func (t *Table) Model() *Model {
	return t.model
}

func (t *Table) Primitive() tview.Primitive {
	return t.view
}

func (t *Table) SetFilter(filter Filter) {
	t.filter = filter
}

func (t *Table) SetOnSelected(handler func (cell *tview.TableCell)) *Table {
	t.view.SetSelectedFunc(func(row, column int) {
		handler(t.view.GetCell(row, column))
	})
	return t
}

func (t *Table) SetOnTab(handler func()) *Table {
	t.onTab = handler
	return t
}

func (t *Table) SetOnEsc(handler func()) *Table {
	t.onEsc = handler
	return t
}

func (t *Table) SetBackgroundColor(color tcell.Color) *Table {
	t.view.SetBackgroundColor(color)
	return t
}

func (t *Table) SetBordersColor(color tcell.Color) *Table {
	t.view.SetBordersColor(color)
	return t
}


func (t *Table) Render() {
	t.view.Clear()

	t.rowIndex = make(map[int]uint64)
	rows := t.model.Contents()
	if t.filter != nil {
		rows = t.filter.Filter(t.model.Contents())
	}

	for nRow, row := range rows { // for each row...
		outputs := t.renderer.Render(row)
		t.rowIndex[nRow] = row.Id()
		for nCol := 0; nCol < row.Len(); nCol++ { // for each cell in the row...
			// TODO: maybe move styling up
			cell := tview.NewTableCell(outputs[nCol]).
				SetTextColor(tcell.ColorWhite)

			t.view.SetCell(nRow, nCol, cell)
		}
	}

	//select first entry
	t.view.Select(0,0)
}

func (t *Table) SelectionDown() {
	row, col := t.view.GetSelection()
	row += 1
	if row < t.view.GetRowCount() {
		t.view.Select(row, col)
	}
}

func (t *Table) SelectionUp() {
	row, col := t.view.GetSelection()
	if row > 0 {
		t.view.Select(row-1, col)
	}
}

func (t *Table) GetSelection() (row int, col int) {
	if t.view.GetRowCount() == 0 {
		return -1, 0
	}
	return t.view.GetSelection()
}

func (t *Table) GetSelectedRow() (Row, bool) {
	r, _ := t.view.GetSelection()
	if r == -1 {
		return nil, false
	}
	return t.model.LookUp(t.rowIndex[r])
}

func (t *Table) Style(theme *ui.Base16Theme) {
	t.SetBackgroundColor(theme.Background)
	t.SetBordersColor(theme.Cyan)
	t.Render()
}


func NewTable(model *Model, renderer Renderer) *Table {
	if model == nil {
		panic("No model given") // TODO
	}
	uiTable := tview.NewTable().SetBorders(false)
	// selection spans whole row
	uiTable.SetSelectable(true, false)

	t := &Table{
		model: model,
		view: uiTable,
		renderer: renderer,
	}

	uiTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTAB {
			if t.onTab != nil {
				t.onTab()
			}
		} else if key == tcell.KeyEscape {
			if t.onEsc != nil {
				t.onEsc()
			}
		}
	})

	model.SetOnChanged(func() {
		t.Render()
	})
	t.Render()

	return t
}
