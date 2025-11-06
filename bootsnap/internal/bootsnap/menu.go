package bootsnap

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevwargo/bootsnap/internal/btrfs"
	"github.com/kevwargo/bootsnap/internal/log"
)

type menu struct {
	pool             *btrfs.Pool
	table            table.Model
	horizontalCursor int
	markedSnapshots  map[string]string
}

func runMenu(pool *btrfs.Pool) error {
	m := menu{pool: pool}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		return err
	}

	for _, v := range pool.Subvols {
		if s, ok := m.markedSnapshots[v.Name]; ok {
			if path, ok := v.SnapshotPaths[s]; ok {
				log.Printf("Restoring %s to %s", v.Path, path)
			}
		}
	}

	return nil
}

func (m *menu) Init() tea.Cmd {
	m.markedSnapshots = make(map[string]string)
	m.buildTable()
	m.updateTable()

	return nil
}

func (m *menu) buildTable() {
	styles := table.DefaultStyles()
	styles.Selected = styles.Selected.Foreground(selectedForeground)

	cols := make([]table.Column, 0, len(m.pool.Subvols)+1)
	cols = append(cols, table.Column{
		Title: timestampTitle,
		Width: max(len(timestampTitle), len(btrfs.SnapshotFormat)),
	})
	for _, v := range m.pool.Subvols {
		cols = append(cols, table.Column{
			Title: v.Name,
			Width: max(minCellWidth, len(v.Name)),
		})
	}

	rows := make([]table.Row, 0, len(m.pool.AllSnapshotNames))
	for _, snapshot := range m.pool.AllSnapshotNames {
		row := make(table.Row, len(m.pool.Subvols)+1)
		row[0] = snapshot
		rows = append(rows, row)
	}

	m.table = table.New(
		table.WithRows(rows),
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithStyles(styles),
	)
}

func (m *menu) updateTable() {
	selectedVolume := m.selectedVolume()

	for ri, row := range m.table.Rows() {
		snapshot := m.pool.AllSnapshotNames[ri]
		selectedRow := ri == m.table.Cursor()

		for vi, subvol := range m.pool.Subvols {
			if _, ok := subvol.SnapshotPaths[snapshot]; !ok {
				continue
			}

			cell := " [ ] "
			if m.markedSnapshots[subvol.Name] == snapshot {
				cell = " [x] "
			}
			if selectedRow && vi == selectedVolume {
				cell = ">" + cell[1:len(cell)-1] + "<"
			}

			row[vi+1] = cell
		}
	}

	m.table.UpdateViewport()
}

func (m *menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKey(msg)
	case tea.WindowSizeMsg:
		height := msg.Height
		count := len(m.pool.AllSnapshotNames)
		if height > count+1 {
			height = count + 1
		} else {
			height--
		}

		m.table.SetHeight(height)
	}

	return m, cmd
}

func (m *menu) View() string {
	return m.table.View() + "\n"
}

func (m *menu) handleKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch k := msg.String(); k {
	case "left":
		m.moveCursor(-1)
	case "right":
		m.moveCursor(1)
	case " ":
		m.toggleMark()
	case "q", "ctrl+c":
		m.markedSnapshots = nil
		cmd = tea.Quit
	case "enter":
		cmd = tea.Quit
	default:
		m.table, cmd = m.table.Update(msg)
	}

	m.updateTable()

	return cmd
}

func (m *menu) toggleMark() {
	subvol := m.pool.Subvols[m.selectedVolume()]
	snapshot := m.pool.AllSnapshotNames[m.table.Cursor()]

	if _, ok := subvol.SnapshotPaths[snapshot]; ok {
		if m.markedSnapshots[subvol.Name] == snapshot {
			delete(m.markedSnapshots, subvol.Name)
		} else {
			m.markedSnapshots[subvol.Name] = snapshot
		}
	}
}

func (m *menu) moveCursor(delta int) {
	m.horizontalCursor = max(0, min(m.horizontalCursor+delta, len(m.pool.Subvols)-1))
	m.horizontalCursor = m.selectedVolume()
}

func (m *menu) selectedVolume() int {
	snapshot := m.pool.AllSnapshotNames[m.table.Cursor()]

	for i := range m.pool.Subvols {
		after := m.horizontalCursor + i
		if after < len(m.pool.Subvols) {
			if _, ok := m.pool.Subvols[after].SnapshotPaths[snapshot]; ok {
				return after
			}
		}

		before := m.horizontalCursor - i
		if before >= 0 {
			if _, ok := m.pool.Subvols[before].SnapshotPaths[snapshot]; ok {
				return before
			}
		}
	}

	return m.horizontalCursor
}

const (
	timestampTitle     = "Timestamp"
	selectedForeground = lipgloss.Color("10")

	// possible non-empty options:
	// ">[x]<"
	// ">[ ]<"
	// " [x] "
	// " [ ] "
	minCellWidth = 5
)
