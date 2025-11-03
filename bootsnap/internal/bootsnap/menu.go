package bootsnap

import (
	"log"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevwargo/bootsnap/internal/btrfs"
)

type menu struct {
	pool            *btrfs.Pool
	table           table.Model
	selectedVolume  int
	markedSnapshots map[string]string
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
			Width: len(v.Name),
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
	for ri, row := range m.table.Rows() {
		snapshot := m.pool.AllSnapshotNames[ri]
		selectedRow := ri == m.table.Cursor()

		for vi, subvol := range m.pool.Subvols {
			prefix := " "
			if selectedRow && vi == m.selectedVolume {
				prefix = ">"
			}

			if _, ok := subvol.SnapshotPaths[snapshot]; !ok {
				row[vi+1] = prefix
			} else if m.markedSnapshots[subvol.Name] == snapshot {
				row[vi+1] = prefix + "[v]"
			} else {
				row[vi+1] = prefix + "[ ]"
			}
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
		m.selectedVolume = max(0, m.selectedVolume-1)
	case "right":
		m.selectedVolume = min(m.selectedVolume+1, len(m.pool.Subvols)-1)
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
	subvol := m.pool.Subvols[m.selectedVolume]
	snapshot := m.pool.AllSnapshotNames[m.table.Cursor()]

	if _, ok := subvol.SnapshotPaths[snapshot]; ok {
		if m.markedSnapshots[subvol.Name] == snapshot {
			delete(m.markedSnapshots, subvol.Name)
		} else {
			m.markedSnapshots[subvol.Name] = snapshot
		}
	}
}

const (
	timestampTitle     = "Timestamp"
	selectedForeground = lipgloss.Color("10")
)
