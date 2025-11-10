package tui

import (
	"cmp"
	"slices"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	table            table.Model
	assoc            map[string]map[string]bool
	marked           map[string]string
	horizontalCursor int
}

func NewTable(title string, data map[string][]string) *Table {
	assoc := make(map[string]map[string]bool, len(data))

	cols := make([]table.Column, 0, len(data)+1)
	cols = append(cols, table.Column{
		Title: title,
		Width: max(minCellWidth, len(title)),
	})

	rowData := make(map[string]struct{})
	for key, values := range data {
		cols = append(cols, table.Column{
			Title: key,
			Width: max(minCellWidth, len(key)),
		})

		assoc[key] = make(map[string]bool, len(values))
		for _, v := range values {
			rowData[v] = struct{}{}
			assoc[key][v] = true
		}
	}

	slices.SortFunc(cols[1:], func(a, b table.Column) int {
		return cmp.Compare(a.Title, b.Title)
	})

	rows := make([]table.Row, 0, len(rowData))
	for name := range rowData {
		row := make(table.Row, len(data)+1)
		row[0] = name
		rows = append(rows, row)
	}
	slices.SortFunc(rows, func(a, b table.Row) int {
		return cmp.Compare(a[0], b[0])
	})

	styles := table.DefaultStyles()
	styles.Selected = styles.Selected.Foreground(selectedForeground)

	t := Table{
		table: table.New(
			table.WithRows(rows),
			table.WithColumns(cols),
			table.WithFocused(true),
			table.WithStyles(styles),
		),
		assoc:  assoc,
		marked: make(map[string]string),
	}
	t.updateTable()

	return &t
}

func (t *Table) View() string {
	return t.table.View()
}

func (t *Table) SetHeight(windowHeight int) {
	height := windowHeight
	count := len(t.table.Rows())

	if height > count+1 {
		height = count + 1
	} else {
		height--
	}

	t.table.SetHeight(height)
}

func (t *Table) Marked() map[string]string {
	return t.marked
}

func (t *Table) HandleKey(km tea.KeyMsg) {
	switch k := km.String(); k {
	case "left":
		t.moveCursor(-1)
	case "right":
		t.moveCursor(1)
	case " ":
		t.toggleMark()
	case "ctrl+a":
		t.toggleSelectedRow()
	default:
		t.table, _ = t.table.Update(km)
	}

	t.updateTable()
}

func (t *Table) updateTable() {
	for i, col := range t.table.Columns()[1:] {
		colName := col.Title

		for j, row := range t.table.Rows() {
			rowName := row[0]

			if !t.assoc[colName][rowName] {
				continue
			}

			marked := t.marked[colName] == rowName
			selected := j == t.table.Cursor() && t.selectedCell() == i

			if marked && selected {
				row[i+1] = cellSelectedMarked
			} else if marked {
				row[i+1] = cellMarked
			} else if selected {
				row[i+1] = cellSelected
			} else {
				row[i+1] = cellEmpty
			}
		}
	}

	t.table.UpdateViewport()
}

func (t *Table) moveCursor(delta int) {
	t.horizontalCursor = max(0, min(t.horizontalCursor+delta, len(t.assoc)-1))
	t.horizontalCursor = t.selectedCell()
}

func (t *Table) selectedCell() int {
	name := t.table.SelectedRow()[0]

	for i := range t.table.Columns() {
		after := t.horizontalCursor + i
		if after < len(t.assoc) {
			if t.assoc[t.table.Columns()[after+1].Title][name] {
				return after
			}
		}

		before := t.horizontalCursor - i
		if before >= 0 {
			if t.assoc[t.table.Columns()[before+1].Title][name] {
				return before
			}
		}
	}

	return t.horizontalCursor
}

func (t *Table) toggleMark() {
	col := t.table.Columns()[t.selectedCell()+1].Title
	row := t.table.SelectedRow()[0]

	if !t.assoc[col][row] {
		return
	}

	if t.marked[col] == row {
		delete(t.marked, col)
	} else {
		t.marked[col] = row
	}
}

func (t *Table) toggleSelectedRow() {
	var (
		rowName = t.table.SelectedRow()[0]

		count    int
		marked   []string
		unmarked []string
	)

	for _, col := range t.table.Columns()[1:] {
		colName := col.Title

		if !t.assoc[colName][rowName] {
			continue
		}

		count++
		if t.marked[colName] == rowName {
			marked = append(marked, colName)
		} else {
			unmarked = append(unmarked, colName)
		}
	}

	if count == len(marked) {
		for _, col := range marked {
			delete(t.marked, col)
		}
	} else {
		for _, col := range unmarked {
			t.marked[col] = rowName
		}
	}
}

const (
	cellEmpty          = " [ ] "
	cellSelected       = ">[ ]<"
	cellMarked         = " [x] "
	cellSelectedMarked = ">[x]<"

	selectedForeground = lipgloss.Color("10")
)

var minCellWidth = max(
	len(cellEmpty),
	len(cellSelected),
	len(cellMarked),
	len(cellSelectedMarked),
)
