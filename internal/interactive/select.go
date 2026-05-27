package interactive

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/user/port-forwarding/internal/config"
	"github.com/user/port-forwarding/internal/ssh"
	"github.com/user/port-forwarding/internal/theme"
	"github.com/user/port-forwarding/internal/tunnel"
)

// --- Model ---

type model struct {
	servers        []config.SSHConfig
	table          table.Model
	selectedIdx    int
	ports          []ssh.PortInfo
	mode           Mode
	keyword        string
	spinner        spinner.Model
	loading        bool
	sshCmd         *exec.Cmd
	localPortInput string
	localPortMode  bool
	windowWidth    int
	windowHeight   int
	tunnels        []tunnel.Tunnel
	tunnelIdx      int
	err            error
}

// --- Key bindings ---

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Quit      key.Binding
	Back      key.Binding
	Filter    key.Binding
	Forward   key.Binding
	SSH       key.Binding
	Stop      key.Binding
	StopAll   key.Binding
	Refresh   key.Binding
	LocalPort key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k")),
		Down:      key.NewBinding(key.WithKeys("down", "j")),
		Enter:     key.NewBinding(key.WithKeys("enter")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c")),
		Back:      key.NewBinding(key.WithKeys("esc")),
		Filter:    key.NewBinding(key.WithKeys("/")),
		Forward:   key.NewBinding(key.WithKeys("f")),
		SSH:       key.NewBinding(key.WithKeys("s")),
		Stop:      key.NewBinding(key.WithKeys("k")),
		StopAll:   key.NewBinding(key.WithKeys("ctrl+u")),
		Refresh:   key.NewBinding(key.WithKeys("r")),
		LocalPort: key.NewBinding(key.WithKeys("l")),
	}
}

// --- Initialization ---

func initialModel(keyword string, mode Mode) model {
	s := spinner.New()
	s.Style = theme.SpinnerStyle

	m := model{
		spinner: s,
		mode:    mode,
		keyword: keyword,
	}

	// Load servers
	servers, err := config.GetConfigWithSearch(keyword)
	if err != nil {
		m.err = err
	}
	m.servers = servers

	// Setup table columns
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Host", Width: 25},
		{Title: "Port", Width: 6},
		{Title: "User", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(m.rows()),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	m.table = t

	return m
}

func (m model) rows() []table.Row {
	var rows []table.Row
	for _, s := range m.servers {
		rows = append(rows, table.Row{s.Name, s.Host, s.Port, s.User})
	}
	return rows
}

// --- Bubble Tea interface ---

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(min(15, msg.Height-10))
		return m, nil

	case tea.KeyMsg:
		return m.updateKeys(msg)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := newKeyMap()

	switch {
	case key.Matches(msg, k.Quit):
		m.stopAllTunnels()
		return m, tea.Quit

	case key.Matches(msg, k.Back):
		return m.handleBack()

	case key.Matches(msg, k.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, k.Down):
		if m.mode == ModeConfig && m.selectedIdx < len(m.servers)-1 {
			m.selectedIdx++
		} else if m.mode == ModePorts && m.selectedIdx < len(m.ports)-1 {
			m.selectedIdx++
		} else if m.mode == ModeTunnels && m.tunnelIdx < len(m.tunnels)-1 {
			m.tunnelIdx++
		}
		return m, nil

	case key.Matches(msg, k.Filter):
		m.table.Focus()
		return m, nil

	case key.Matches(msg, k.Enter):
		return m.handleEnter()

	case key.Matches(msg, k.Forward):
		if m.mode == ModePorts && len(m.ports) > m.selectedIdx {
			m.mode = ModeTunnelCreate
			m.localPortMode = true
			m.localPortInput = strconv.Itoa(m.ports[m.selectedIdx].Port)
		}
		return m, nil

	case key.Matches(msg, k.SSH):
		if m.mode == ModeConfig && m.selectedIdx < len(m.servers) {
			return m, m.launchSSH(m.servers[m.selectedIdx])
		}

	case key.Matches(msg, k.Stop):
		if m.mode == ModeTunnels && m.tunnelIdx < len(m.tunnels) {
			t := m.tunnels[m.tunnelIdx]
			tunnel.Stop(t.ID)
			m.loadTunnels()
		}

	case key.Matches(msg, k.StopAll):
		m.stopAllTunnels()

	case key.Matches(msg, k.Refresh):
		if m.mode == ModeTunnels {
			m.loadTunnels()
		}

	case key.Matches(msg, k.LocalPort):
		if m.mode == ModePorts && len(m.ports) > m.selectedIdx {
			m.localPortMode = true
			m.localPortInput = strconv.Itoa(m.ports[m.selectedIdx].Port)
		}
	}

	return m, nil
}

func (m *model) handleBack() (tea.Model, tea.Cmd) {
	if m.sshCmd != nil && m.sshCmd.Process != nil {
		m.sshCmd.Process.Signal(os.Interrupt)
		m.sshCmd = nil
		return m, nil
	}
	if m.mode == ModeTunnelCreate || m.mode == ModePorts {
		m.mode = ModeConfig
		m.ports = nil
	} else if m.mode == ModeSSH {
		m.mode = ModeConfig
	}
	return m, nil
}

func (m *model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeConfig:
		if m.selectedIdx < len(m.servers) {
			m.mode = ModeActionSelect
		}
	case ModePorts:
		if m.selectedIdx < len(m.ports) {
			m.localPortMode = true
			m.localPortInput = strconv.Itoa(m.ports[m.selectedIdx].Port)
		}
	case ModeTunnels:
		// Show tunnel details
	}
	return m, nil
}

func (m *model) launchSSH(server config.SSHConfig) tea.Cmd {
	args := []string{
		"-p", server.Port,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	m.mode = ModeSSH
	m.sshCmd = cmd

	return func() tea.Msg {
		_ = cmd.Run()
		m.sshCmd = nil
		m.mode = ModeConfig
		return nil
	}
}

func (m *model) loadTunnels() {
	tunnels, err := tunnel.LoadState()
	if err != nil {
		m.err = err
		return
	}
	m.tunnels = tunnels
}

func (m *model) stopAllTunnels() {
	tunnel.StopAll()
	m.loadTunnels()
}

// --- View ---

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit.\n", m.err)
	}

	var s strings.Builder

	switch m.mode {
	case ModeConfig:
		s.WriteString(m.viewServerList())
	case ModePorts:
		s.WriteString(m.viewPortList())
	case ModeTunnelCreate:
		s.WriteString(m.viewTunnelCreate())
	case ModeSSH:
		s.WriteString(m.viewSSH())
	case ModeTunnels:
		s.WriteString(m.viewTunnelManager())
	case ModeActionSelect:
		s.WriteString(m.viewActionSelect())
	}

	s.WriteString("\n")
	s.WriteString(m.statusBar())

	return s.String()
}

func (m model) viewServerList() string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render("gpf — Greenfield Port Forwarding") + "\n\n")

	if m.keyword != "" {
		s.WriteString(theme.MutedStyle.Render("Filter: " + m.keyword) + "\n\n")
	}

	if m.loading {
		s.WriteString("  " + m.spinner.View() + " Loading servers...\n")
		return s.String()
	}

	if len(m.servers) == 0 {
		s.WriteString("  No servers found.\n")
		return s.String()
	}

	// Render table
	s.WriteString(m.table.View() + "\n")

	s.WriteString(theme.MutedStyle.Render(
		"  ↑↓ navigate  / filter  enter:action  q:quit"))

	return s.String()
}

func (m model) viewActionSelect() string {
	if m.selectedIdx >= len(m.servers) {
		return "No server selected.\n"
	}
	server := m.servers[m.selectedIdx]
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Server: %s\n\n", server.Name))
	s.WriteString(fmt.Sprintf("  %s [P]ort Forward — 포트 포워딩으로 연결\n", theme.ActionStyle.Render("▶")))
	s.WriteString(fmt.Sprintf("  %s [S]SH Connect  — SSH 직접 접속\n\n", theme.MutedStyle.Render(" ")))
	s.WriteString(theme.MutedStyle.Render("↑↓ navigate  enter:execute  esc:back"))
	return s.String()
}

func (m model) viewPortList() string {
	server := m.servers[m.selectedIdx]
	var s strings.Builder

	s.WriteString(theme.HeaderStyle.Render("Server: " + server.Name) + "  " +
		theme.MutedStyle.Render("◀ Back") + "\n\n")

	if m.loading {
		s.WriteString("  " + m.spinner.View() + " Scanning ports...\n")
		return s.String()
	}

	if len(m.ports) == 0 {
		s.WriteString("  No listening ports found.\n")
		return s.String()
	}

	// Port table
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 8},
		{Title: "LocalAddr", Width: 14},
		{Title: "Process", Width: 20},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(m.portRows()),
		table.WithFocused(true),
		table.WithHeight(min(15, len(m.ports)+2)),
	)
	t.SetWidth(m.windowWidth - 6)
	s.WriteString(t.View() + "\n\n")

	s.WriteString(theme.MutedStyle.Render(
		"↑↓ navigate  enter:forward  f:forward  esc:back  q:quit"))

	return s.String()
}

func (m model) portRows() []table.Row {
	var rows []table.Row
	for _, p := range m.ports {
		rows = append(rows, table.Row{
			strconv.Itoa(p.Port),
			p.Protocol,
			p.LocalAddr,
			p.Process,
		})
	}
	return rows
}

func (m model) viewTunnelCreate() string {
	var s strings.Builder
	s.WriteString("Creating tunnel...\n\n")

	if m.localPortMode {
		cursor := " "
		if m.localPortInput != "" {
			cursor = "█"
		}
		s.WriteString(fmt.Sprintf("  Local port: [%s%s]\n\n", m.localPortInput, cursor))
		s.WriteString(theme.MutedStyle.Render("Enter: create tunnel  esc: cancel"))
	} else {
		s.WriteString("  " + m.spinner.View() + " Starting tunnel...\n")
	}

	return s.String()
}

func (m model) viewSSH() string {
	var s strings.Builder
	s.WriteString("SSH Connection (ctrl+c to detach)\n")
	if m.sshCmd != nil && m.sshCmd.Process != nil {
		s.WriteString(fmt.Sprintf("  PID: %d\n", m.sshCmd.Process.Pid))
	}
	return s.String()
}

func (m model) viewTunnelManager() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("Active Tunnels") + "  " +
		theme.MutedStyle.Render("◀ Back  r:refresh") + "\n\n")

	m.loadTunnels()

	if len(m.tunnels) == 0 {
		s.WriteString("  No active tunnels.\n")
		return s.String()
	}

	columns := []table.Column{
		{Title: "Local", Width: 12},
		{Title: "Remote", Width: 12},
		{Title: "Server", Width: 20},
		{Title: "PID", Width: 8},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(m.tunnelRows()),
		table.WithFocused(true),
		table.WithHeight(min(15, len(m.tunnels)+2)),
	)
	t.SetWidth(m.windowWidth - 6)
	s.WriteString(t.View() + "\n\n")

	s.WriteString(theme.MutedStyle.Render(
		"↑↓ navigate  k:kill  ctrl+u:stop-all  r:refresh"))

	return s.String()
}

func (m model) tunnelRows() []table.Row {
	var rows []table.Row
	for _, t := range m.tunnels {
		rows = append(rows, table.Row{
			fmt.Sprintf(":%d", t.LocalPort),
			fmt.Sprintf(":%d", t.RemotePort),
			t.ServerName,
			strconv.Itoa(t.PID),
		})
	}
	return rows
}

func (m model) statusBar() string {
	var status string
	switch m.mode {
	case ModeConfig:
		status = fmt.Sprintf("Servers: %d", len(m.servers))
	case ModePorts:
		status = fmt.Sprintf("Ports: %d", len(m.ports))
	case ModeTunnels:
		status = fmt.Sprintf("Tunnels: %d", len(m.tunnels))
	default:
		status = ""
	}
	return theme.StatusBarStyle.Render(status)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
