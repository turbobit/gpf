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
	"github.com/turbobit/gpf/internal/config"
	"github.com/turbobit/gpf/internal/i18n"
	"github.com/turbobit/gpf/internal/ssh"
	"github.com/turbobit/gpf/internal/theme"
	"github.com/turbobit/gpf/internal/tunnel"
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
	L              *i18n.Translator
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
		L:       i18n.Default(),
	}

	// Load servers
	servers, err := config.GetConfigWithSearch(keyword)
	if err != nil {
		m.err = err
	}
	m.servers = servers

	// Setup table columns
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: m.L.T("server_list"), Width: 30},
			{Title: "Host", Width: 25},
			{Title: "Port", Width: 6},
			{Title: "User", Width: 10},
		}),
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

	s.WriteString(theme.TitleStyle.Render(m.L.T("title")) + "\n\n")

	if m.keyword != "" {
		s.WriteString(theme.MutedStyle.Render(m.L.T("filter")+": "+m.keyword) + "\n\n")
	}

	if m.loading {
		s.WriteString("  " + m.spinner.View() + " "+m.L.T("loading_servers")+"\n")
		return s.String()
	}

	if len(m.servers) == 0 {
		s.WriteString("  "+m.L.T("no_servers")+"\n")
		return s.String()
	}

	// Render table
	s.WriteString(m.table.View() + "\n")

	s.WriteString(theme.MutedStyle.Render(
		"  ↑↓ "+m.L.T("navigate")+"  / "+m.L.T("filter")+"  enter:"+m.L.T("action")+"  q:"+m.L.T("quit")))

	return s.String()
}

func (m model) viewActionSelect() string {
	if m.selectedIdx >= len(m.servers) {
		return m.L.T("no_servers") + "\n"
	}
	server := m.servers[m.selectedIdx]
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Server: %s\n\n", server.Name))
	s.WriteString(fmt.Sprintf("  %s [P] %s — %s\n", theme.ActionStyle.Render("▶"), m.L.T("port_forward"), m.L.T("port_forward_desc")))
	s.WriteString(fmt.Sprintf("  %s [S] %s — %s\n\n", theme.MutedStyle.Render(" "), m.L.T("ssh_connect"), m.L.T("ssh_connect_desc")))
	s.WriteString(theme.MutedStyle.Render("↑↓ "+m.L.T("navigate")+"  enter:execute  esc:"+m.L.T("back")))
	return s.String()
}

func (m model) viewPortList() string {
	server := m.servers[m.selectedIdx]
	var s strings.Builder

	s.WriteString(theme.HeaderStyle.Render("Server: "+server.Name) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")) + "\n\n")

	if m.loading {
		s.WriteString("  " + m.spinner.View() + " "+m.L.T("scanning_ports")+"\n")
		return s.String()
	}

	if len(m.ports) == 0 {
		s.WriteString("  "+m.L.T("no_ports")+"\n")
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
		"↑↓ "+m.L.T("navigate")+"  enter:"+m.L.T("forward")+"  f:"+m.L.T("forward")+"  esc:"+m.L.T("back")+"  q:"+m.L.T("quit")))

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
	s.WriteString(m.L.T("creating_tunnel") + "\n\n")

	if m.localPortMode {
		cursor := " "
		if m.localPortInput != "" {
			cursor = "█"
		}
		s.WriteString(fmt.Sprintf("  %s: [%s%s]\n\n", m.L.T("local_port"), m.localPortInput, cursor))
		s.WriteString(theme.MutedStyle.Render("Enter: "+m.L.T("create_tunnel")+"  esc: "+m.L.T("cancel")))
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
	s.WriteString(theme.HeaderStyle.Render(m.L.T("active_tunnels")) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")+"  r:"+m.L.T("refresh")) + "\n\n")

	m.loadTunnels()

	if len(m.tunnels) == 0 {
		s.WriteString("  "+m.L.T("no_tunnels")+"\n")
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
		"↑↓ "+m.L.T("navigate")+"  k:"+m.L.T("kill")+"  ctrl+u:"+m.L.T("stop_all")+"  r:"+m.L.T("refresh")))

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
		status = fmt.Sprintf("%s: %d", m.L.T("servers_count"), len(m.servers))
	case ModePorts:
		status = fmt.Sprintf("%s: %d", m.L.T("ports_count"), len(m.ports))
	case ModeTunnels:
		status = fmt.Sprintf("%s: %d", m.L.T("tunnels_count"), len(m.tunnels))
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
