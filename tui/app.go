package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/noahgorstein/dog-watcher/statusbar"
	stardog "github.com/noahgorstein/go-stardog/stardog"
)

var (
	tableHeaderStyle = lipgloss.NewStyle().Bold(true)
	tableBorder      = table.Border{
		Top:    "─",
		Left:   "│",
		Right:  "│",
		Bottom: "─",

		TopRight:    "╮",
		TopLeft:     "╭",
		BottomRight: "╯",
		BottomLeft:  "╰",

		TopJunction:    "┬",
		LeftJunction:   "├",
		RightJunction:  "┤",
		BottomJunction: "┴",
		InnerJunction:  "┼",

		InnerDivider: "│",
	}
	// Frost
	nord8  = lipgloss.Color("#88C0D0")
	nord10 = lipgloss.Color("#5E81AC")
	nord11 = lipgloss.Color("#BF616A") // red
)

const (
	columnKeyID          = "id"
	columnKeyType        = "type"
	columnKeyDb          = "db"
	columnKeyUser        = "user"
	columnKeyElapsedTime = "elapsedTime"
	columnKeyStatus      = "status"
	columnKeyProgress    = "progress"
)

type Model struct {
	table         table.Model
	updateDelay   time.Duration
	stardogClient *stardog.Client
	statusbar     statusbar.Bubble
	width         int
	height        int
}

func NewModel(stardogClient *stardog.Client) Model {
	c := stardogClient
	sb := statusbar.New(*c)
	sb.StatusMessageLifetime = time.Duration(15 * time.Second)
	t := table.New(generateColumns()).Focused(true).SelectableRows(true).
		Border(tableBorder).WithPageSize(10).HeaderStyle(tableHeaderStyle).
		WithBaseStyle(lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.AdaptiveColor{
				Light: string(lipgloss.Color("16")),
				Dark:  string(lipgloss.Color("15")),
			})).
		Filtered(true)
	return Model{
		table:         t,
		updateDelay:   time.Second * 5,
		stardogClient: c,
		statusbar:     sb,
	}
}

type successMsg struct {
	message string
}

type getServerProcessesMsg struct {
	processes stardog.Processes
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m Model) getServerProcessesCmd(delayed bool) tea.Cmd {
	return func() tea.Msg {

		if delayed {
			time.Sleep(m.updateDelay)
		}
		processes, err := m.stardogClient.ServerAdmin.GetProcesses(context.Background())
		if err != nil {
			return errMsg{
				err: err,
			}
		}
		return getServerProcessesMsg{
			processes: processes,
		}
	}
}

func (m Model) killProcessCmd(processId string, processType string) tea.Cmd {
	return func() tea.Msg {

		_, err := m.stardogClient.ServerAdmin.KillProcess(context.Background(), processId)
		if err != nil {
			return errMsg{
				err: fmt.Errorf("Unable to kill process: %s", err.Error()),
			}
		}

		return successMsg{
			message: fmt.Sprintf("Successfully killed %s process with ID: %s ", processType, processId),
		}

	}
}

func generateColumns() []table.Column {
	return []table.Column{
		table.NewFlexColumn(columnKeyID, "ID", 25).WithFiltered(true),
		table.NewFlexColumn(columnKeyDb, "Database", 16).WithFiltered(true),
		table.NewFlexColumn(columnKeyElapsedTime, "Elapsed Time", 14),
		table.NewFlexColumn(columnKeyUser, "User", 16).WithFiltered(true),
		table.NewColumn(columnKeyStatus, "Status", 12).WithFiltered(true),
		table.NewColumn(columnKeyType, "Type", 20).WithFiltered(true),
		table.NewColumn(columnKeyProgress, "Progress", 30),
	}
}

func (m Model) Init() tea.Cmd {
	return m.getServerProcessesCmd(false)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.statusbar.SetWidth(msg.Width - 4)
		m.table = m.table.WithTargetWidth(msg.Width - 2)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			cmds = append(cmds, tea.Quit)
		case "ctrl+x":
			selectedProccessID := m.table.HighlightedRow().Data[columnKeyID].(string)
			selectedProcessType := m.table.HighlightedRow().Data[columnKeyType].(string)
			killProcessCmd := m.killProcessCmd(selectedProccessID, selectedProcessType)
			cmds = append(cmds, killProcessCmd)
		case "i":
			if !m.table.GetIsFilterInputFocused() {
				m.updateDelay += time.Second
			}
		case "d":
			if !m.table.GetIsFilterInputFocused() {
				if m.updateDelay >= 2*time.Second {
					m.updateDelay -= time.Second
				}
			}
		}
	case getServerProcessesMsg:
		m.table = m.table.WithRows(generateRowsFromData(msg)).WithColumns(generateColumns())

		getServerProccessesCmd := m.getServerProcessesCmd(true)
		cmds = append(cmds, getServerProccessesCmd)
	case successMsg:
		newStatusMsgCmd := m.statusbar.NewStatusMessage(msg.message, true)
		cmds = append(cmds, newStatusMsgCmd)

		getServerProccessesCmd := m.getServerProcessesCmd(false)
		cmds = append(cmds, getServerProccessesCmd)
	case errMsg:
		newStatusMsgCmd := m.statusbar.NewStatusMessage(msg.err.Error(), false)
		cmds = append(cmds, newStatusMsgCmd)
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.statusbar, cmd = m.statusbar.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	logo := strings.Builder{}
	logo.WriteString(" ____   __    ___    _  _   __  ____  ___  _  _  ____  ____ \n")
	logo.WriteString("(    \\ /  \\  / __)  / )( \\ / _\\(_  _)/ __)/ )( \\(  __)(  _ \n")
	logo.WriteString(" ) D ((  O )( (_ \\  \\ /\\ //    \\ )( ( (__ ) __ ( ) _)  )   /\n")
	logo.WriteString("(____/ \\__/  \\___/  (_/\\_)\\_/\\_/(__) \\___)\\_)(_/(____)(__\\_)")

	body := strings.Builder{}
	body.WriteString(lipgloss.NewStyle().Bold(true).PaddingTop(1).PaddingLeft(1).
		Foreground(lipgloss.AdaptiveColor{Light: string(lipgloss.Color("38")), Dark: string(lipgloss.Color("152"))}).
		Render(logo.String()))
	body.WriteRune('\n')
	body.WriteRune('\n')
	body.WriteString(lipgloss.NewStyle().Bold(true).Italic(true).PaddingLeft(1).Render("a Stardog process manager"))
	body.WriteRune('\n')
	body.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("———————————————————————————————————————————————————————————————————————————————"))
	body.WriteRune('\n')
	body.WriteRune('\n')
	body.WriteString(lipgloss.NewStyle().Bold(true).PaddingLeft(1).PaddingTop(1).
		Render(
			fmt.Sprintf(
				"Processes updated every: %s | %s/%s to decrement/increment by 1s\n",
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
					Light: string(nord11),
					Dark:  string(nord11),
				}).Render(m.updateDelay.String()),
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
					Light: string(nord10),
					Dark:  string(nord8),
				}).Render("d"),
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
					Light: string(nord10),
					Dark:  string(nord8),
				}).Render("i"))))
	body.WriteString(m.table.View())
  body.WriteRune('\n')

	footer := strings.Builder{}
	footer.WriteString(m.statusbar.View())

	main := strings.Builder{}
	main.WriteString(body.String())

	return lipgloss.NewStyle().Width(m.width).Height(m.height).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			main.String(),
			footer.String()))
}

const (
	NotStarted string = "NOT_STARTED"
	Running           = "RUNNING"
	Killed            = "KILLED"
	Finished          = "FINISHED"
)

func generateRowsFromData(data getServerProcessesMsg) []table.Row {

	rows := []table.Row{}

	for _, ps := range data.processes {
		row := table.NewRow(table.RowData{
			columnKeyID:          ps.ID,
			columnKeyDb:          ps.Db,
			columnKeyUser:        ps.User,
			columnKeyType:        ps.Type,
			columnKeyElapsedTime: getElapsedTime(ps.StartTime),
			columnKeyStatus:      ps.Status,
			columnKeyProgress:    formatProgress(ps.Progress.Current, ps.Progress.Max, ps.Progress.Stage),
		})

		switch ps.Status {
		case Running:
			row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
				Light: string(lipgloss.Color("2")),
				Dark:  string(lipgloss.Color("10")),
			}))
		case Killed:
			row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
				Light: string(lipgloss.Color("1")),
				Dark:  string(lipgloss.Color("160")),
			}))

		}
		rows = append(rows, row)
	}
	return rows
}

func formatProgress(current int, max int, stage string) string {

	if max != 0 {
		progress := (float64(current) / float64(max)) * 100

		if stage != "" {
			return fmt.Sprintf("%s %.2f%%", stage, progress)
		}
		return fmt.Sprintf("%.2f%%", progress)
	}
	if stage != "" {
		return stage
	} else {
		return "N/A"
	}

}

func getElapsedTime(startTime int64) string {

	now := time.Now()
	start := time.UnixMilli(startTime)
	elapsedTime := now.Sub(start).Round(time.Millisecond)

	return elapsedTime.String()
}
