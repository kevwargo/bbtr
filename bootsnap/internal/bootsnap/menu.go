package bootsnap

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevwargo/bootsnap/internal/btrfs"
	"github.com/kevwargo/bootsnap/internal/log"
	"github.com/kevwargo/bootsnap/internal/tui"
)

func runMenu(pool *btrfs.Pool) error {
	m := menu{pool: pool}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		return err
	}

	if !m.complete {
		log.Println("program interrupted")

		return nil
	}

	for _, subvol := range pool.Subvols {
		if backup, ok := m.backupsTable.Marked()[subvol.Name]; ok {
			log.Printf("btrfs subvolume snapshot -r %s %s~%s", subvol.Path, subvol.Name, backup)
		}

		snapName, ok := m.snapshotsTable.Marked()[subvol.Name]
		if !ok {
			continue
		}

		if snapPath, ok := subvol.SnapshotPaths[snapName]; ok {
			log.Printf("btrfs subvolume snapshot %s %s", snapPath, subvol.Path)
		}
	}

	return nil
}

type menu struct {
	pool           *btrfs.Pool
	snapshotsTable *tui.Table
	backupsTable   *tui.Table
	windowHeight   int
	quitting       bool
	complete       bool
}

func (m *menu) Init() tea.Cmd {
	m.snapshotsTable = tui.NewTable("Snapshot timestamp", m.pool.Table())
	m.backupsTable = nil

	return nil
}

func (m *menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch k := msg.String(); k {
		case "ctrl+c", "q":
			return m, m.quit
		case "enter":
			return m, m.switchOrQuit()
		default:
			m.currentTable().HandleKey(msg)
		}
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.currentTable().SetHeight(m.windowHeight)
	}

	return m, nil
}

func (m *menu) View() string {
	if v := m.currentTable().View(); m.quitting {
		return v + "\n"
	} else {
		return v
	}
}

func (m *menu) currentTable() *tui.Table {
	if m.backupsTable != nil {
		return m.backupsTable
	}

	return m.snapshotsTable
}

func (m *menu) switchOrQuit() tea.Cmd {
	if m.backupsTable != nil {
		m.complete = true

		return m.quit
	}

	now := time.Now().Format(btrfs.SnapshotFormat)
	data := make(map[string][]string)
	for _, subvol := range m.pool.Subvols {
		data[subvol.Name] = []string{now}
	}

	m.backupsTable = tui.NewTable("Create backup(s)?", data)
	if m.windowHeight > 0 {
		m.backupsTable.SetHeight(m.windowHeight)
	}

	return tea.Println(m.snapshotsTable.View())
}

func (m *menu) quit() tea.Msg {
	m.quitting = true

	return tea.QuitMsg{}
}
