package interactive

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
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
	selectedServer config.SSHConfig // cached when server is selected (ggh pattern)
	sshTarget      config.SSHConfig
	portsServer    config.SSHConfig // server whose ports we're scanning
	forwardTarget  *ForwardTarget
	localPortInput string
	localPortMode  bool
	// Server table state — saved before port table overwrites m.table
	serverRows     []table.Row
	serverCursor   int
	windowWidth    int
	windowHeight   int
	tunnels        []tunnel.Tunnel
	tunnelIdx      int
	actionIdx      int
	err            error
	L              *i18n.Translator
	// Tunnel creation state
	tunnelCreating   bool
	tunnelCreated    bool
	tunnelLocalPort  int
	tunnelRemotePort int
	tunnelPID        int
	tunnelErr        error
	locales          []string
	localeIdx        int
}

// ForwardTarget holds the info needed for port forwarding.
type ForwardTarget struct {
	Server     config.SSHConfig
	LocalPort  int
	RemotePort int
}

// --- Key bindings ---

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
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
	Lang      key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k")),
		Down:      key.NewBinding(key.WithKeys("down", "j")),
		Left:      key.NewBinding(key.WithKeys("left", "h")),
		Right:     key.NewBinding(key.WithKeys("right", "l")),
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
		Lang:      key.NewBinding(key.WithKeys("L")),
	}
}

// --- Initialization ---

func initialModel(keyword string, mode Mode) model {
	s := spinner.New()
	s.Style = theme.SpinnerStyle

	m := model{
		spinner:     s,
		mode:        mode,
		keyword:     keyword,
		L:           i18n.Default(),
		locales:     []string{"en", "ko", "zh"},
		localeIdx:   0,
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
		// Ignore null bytes from spinner/input device
		if msg.String() == "" || msg.String() == "\x00" {
			return m, nil
		}
		// Consume Enter in non-table modes to prevent table from processing it
		if msg.String() == "enter" {
			switch m.mode {
			case ModeConfig, ModePorts, ModeActionSelect, ModeTunnelCreate, ModeTunnels:
				return m.updateKeys(msg)
			}
		}
		return m.updateKeys(msg)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case portScanMsg:
		if m.loading {
			m.loading = false
			m.ports = msg.ports
			sort.Slice(m.ports, func(i, j int) bool {
				return m.ports[i].Port < m.ports[j].Port
			})
			if msg.err != nil {
				m.err = msg.err
			} else {
				// Initialize port table
				m.table = table.New(
					table.WithColumns([]table.Column{
						{Title: "Port", Width: 8},
						{Title: "Tunnel", Width: 9},
						{Title: "Proto", Width: 6},
						{Title: "Addr", Width: 14},
						{Title: "Process", Width: 18},
					}),
					table.WithRows(m.portRows()),
					table.WithFocused(true),
				)
				m.table.SetWidth(m.windowWidth - 6)
				m.table.SetHeight(min(15, len(m.ports)+2))
			}
		}
		return m, nil

	case tunnelMsg:
		m.tunnelCreating = false
		if msg.err != nil {
			m.tunnelErr = msg.err
			m.tunnelCreated = false
		} else {
			m.tunnelLocalPort = msg.localPort
			m.tunnelPID = msg.pid
			m.tunnelRemotePort = msg.remotePort
			m.tunnelCreated = true
			m.tunnelErr = nil
			// Reload tunnels list so indicators show
			m.loadTunnels()
		}
		return m, nil

	case tunnelResetMsg:
		m.tunnelCreated = false
		m.tunnelErr = nil
		// Re-initialize port table
		if len(m.ports) > 0 {
			m.table = table.New(
				table.WithColumns([]table.Column{
					{Title: "Port", Width: 8},
					{Title: "Tunnel", Width: 9},
					{Title: "Proto", Width: 6},
					{Title: "Addr", Width: 14},
					{Title: "Process", Width: 18},
				}),
				table.WithRows(m.portRows()),
				table.WithFocused(true),
			)
			m.table.SetWidth(m.windowWidth - 6)
			m.table.SetHeight(min(15, len(m.ports)+2))
		}
		return m, nil
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
		var cmd tea.Cmd
		if m.mode == ModeConfig || m.mode == ModePorts || m.mode == ModeTunnels {
			m.table, cmd = m.table.Update(msg)
		} else if m.mode == ModeActionSelect {
			if m.actionIdx > 0 {
				m.actionIdx--
			}
		}
		return m, cmd

	case key.Matches(msg, k.Down):
		var cmd tea.Cmd
		if m.mode == ModeConfig || m.mode == ModePorts || m.mode == ModeTunnels {
			m.table, cmd = m.table.Update(msg)
		} else if m.mode == ModeActionSelect {
			if m.actionIdx < 1 {
				m.actionIdx++
			}
		}
		return m, cmd

	case key.Matches(msg, k.Left):
		if m.mode == ModeActionSelect || m.mode == ModePorts {
			return m.handleBack()
		}
		return m, nil

	case key.Matches(msg, k.Right):
		return m, nil

	case key.Matches(msg, k.Filter):
		m.table.Focus()
		return m, nil

	case key.Matches(msg, k.Enter):
		if m.mode == ModePorts {
			// Don't create tunnel on Enter when tunnel created/error screen is showing
			if m.tunnelCreated || m.tunnelErr != nil {
				return m, nil
			}
			if len(m.ports) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.ports) {
					m.tunnelCreating = true
					m.tunnelCreated = false
					m.tunnelErr = nil
					port := m.ports[idx].Port
					return m, m.createTunnelCmd(m.portsServer, port, port)
				}
			}
			return m, nil
		}
		return m.handleEnter()

	case key.Matches(msg, k.Forward):
		if m.mode == ModePorts {
			if len(m.ports) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.ports) {
					m.mode = ModeTunnelCreate
					m.localPortMode = true
					m.localPortInput = strconv.Itoa(m.ports[idx].Port)
				}
			}
		}
		return m, nil

	case key.Matches(msg, k.SSH):
		if m.mode == ModeConfig && len(m.servers) > 0 {
			server := m.serverFromRow()
			if server.Name != "" {
				m.sshTarget = server
				return m, tea.Quit
			}
		}

	case key.Matches(msg, k.Stop):
		if m.mode == ModeTunnels && len(m.tunnels) > 0 {
			row := m.table.SelectedRow()
			if len(row) > 0 {
				for _, t := range m.tunnels {
					if strconv.Itoa(t.PID) == row[3] {
						tunnel.Stop(t.ID)
						m.loadTunnels()
						break
					}
				}
			}
		}

	case key.Matches(msg, k.StopAll):
		m.stopAllTunnels()

	case key.Matches(msg, k.Refresh):
		if m.mode == ModeTunnels {
			m.loadTunnels()
		}

	case key.Matches(msg, k.LocalPort):
		if m.mode == ModePorts {
			if len(m.ports) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.ports) {
					m.localPortMode = true
					m.localPortInput = strconv.Itoa(m.ports[idx].Port)
				}
			}
		}

	case key.Matches(msg, k.Lang):
		m.localeIdx = (m.localeIdx + 1) % len(m.locales)
		locale := m.locales[m.localeIdx]
		m.L = i18n.For(locale)
		i18n.Save(locale)
		m.rebuildColumns()

	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "esc"))):
		if m.mode == ModePorts && (m.tunnelCreated || m.tunnelErr != nil) {
			// Left/Esc → stop tunnel and return to list
			if m.tunnelPID > 0 {
				m.stopTunnelForPort(m.ports[m.table.Cursor()].Port)
			}
			m.tunnelCreated = false
			m.tunnelErr = nil
			m.tunnelPID = 0
			m.tunnelLocalPort = 0
			m.tunnelRemotePort = 0
			m.loadTunnels()
			m.table = table.New(
				table.WithColumns([]table.Column{
					{Title: "Port", Width: 8},
					{Title: "Tunnel", Width: 9},
					{Title: "Proto", Width: 6},
					{Title: "Addr", Width: 14},
					{Title: "Process", Width: 18},
				}),
				table.WithRows(m.portRows()),
				table.WithFocused(true),
			)
			m.table.SetWidth(m.windowWidth - 6)
			m.table.SetHeight(min(15, len(m.ports)+2))
			return m, nil
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if m.mode == ModePorts {
			// Always kill by port — works even if state file is out of sync
			port := m.tunnelLocalPort
			if port == 0 && len(m.ports) > 0 {
				port = m.ports[m.table.Cursor()].Port
			}
			if port > 0 {
				if err := tunnel.KillProcessByPort(port); err != nil {
					fmt.Fprintf(os.Stderr, "No SSH tunnel on port %d: %v\n", port, err)
				}
			}
			// Stay on screen for debugging — auto-return disabled
		}
	}

	return m, nil
}

func (m *model) rebuildColumns() {
	switch m.mode {
	case ModeConfig:
		m.table.SetColumns([]table.Column{
			{Title: m.L.T("server_list"), Width: 30},
			{Title: "Host", Width: 25},
			{Title: "Port", Width: 6},
			{Title: "User", Width: 10},
		})
	case ModePorts:
		m.table.SetColumns([]table.Column{
			{Title: "Port", Width: 8},
			{Title: "Tunnel", Width: 9},
			{Title: "Proto", Width: 6},
			{Title: "Addr", Width: 14},
			{Title: "Process", Width: 18},
		})
	case ModeTunnels:
		m.table.SetColumns([]table.Column{
			{Title: "Local", Width: 12},
			{Title: "Remote", Width: 12},
			{Title: "Server", Width: 20},
			{Title: "PID", Width: 8},
		})
	}
}

func (m *model) handleBack() (tea.Model, tea.Cmd) {
	if m.mode == ModeTunnelCreate || m.mode == ModePorts || m.mode == ModeActionSelect {
		m.mode = ModeConfig
		m.ports = nil
		m.actionIdx = 0
		m.forwardTarget = nil
		m.sshTarget = config.SSHConfig{}
		m.selectedServer = config.SSHConfig{}
		// Restore server table (was overwritten by port table)
		if len(m.serverRows) > 0 {
			m.table.SetRows(m.serverRows)
			m.table.SetCursor(m.serverCursor)
		}
	}
	return m, nil
}

func (m *model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeConfig:
		if len(m.servers) > 0 {
			// Cache the selected server immediately (ggh pattern)
			m.selectedServer = m.serverFromRow()
			m.mode = ModeActionSelect
			m.actionIdx = 0
		}
	case ModeActionSelect:
		// Use cached server, not table lookup
		server := m.selectedServer
		if server.Name == "" {
			return m, nil
		}
		if m.actionIdx == 0 {
			// SSH Connect
			m.sshTarget = server
			return m, tea.Quit
		}
		// Port Forward - scan ports
		// Save server table state before it gets overwritten by port table
		m.serverRows = m.table.Rows()
		m.serverCursor = m.table.Cursor()
		m.mode = ModePorts
		m.portsServer = server
		m.loading = true
		return m, m.scanPorts(server)
	case ModePorts:
		// Enter on port → start forwarding
		if len(m.ports) > 0 {
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.ports) {
				m.forwardTarget = &ForwardTarget{
					Server:     m.portsServer,
					LocalPort:  m.ports[idx].Port,
					RemotePort: m.ports[idx].Port,
				}
				return m, tea.Quit
			}
		}
	case ModeTunnelCreate:
		// Enter on local port input → start forwarding
		if m.localPortMode && m.localPortInput != "" {
			if len(m.ports) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.ports) {
					localPort, err := strconv.Atoi(m.localPortInput)
					if err == nil {
						m.forwardTarget = &ForwardTarget{
							Server:     m.portsServer,
							LocalPort:  localPort,
							RemotePort: m.ports[idx].Port,
						}
						return m, tea.Quit
					}
				}
			}
		}
	case ModeTunnels:
		// Show tunnel details
	}
	return m, nil
}

func (m *model) serverFromRow() config.SSHConfig {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return config.SSHConfig{}
	}
	for _, s := range m.servers {
		if s.Name == row[0] {
			return s
		}
	}
	return config.SSHConfig{}
}

func (m *model) scanPorts(server config.SSHConfig) tea.Cmd {
	args := []string{
		"-p", server.Port,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	return func() tea.Msg {
		// Test 1: lsof (macOS)
		c1 := exec.Command("ssh", append(args, "lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n")...)
		out1, _ := c1.CombinedOutput()

		// Test 2: lsof without -sTCP:LISTEN (broader)
		c2 := exec.Command("ssh", append(args, "lsof", "-iTCP", "-P", "-n")...)
		out2, _ := c2.CombinedOutput()

		// Test 3: ss (Linux)
		c3 := exec.Command("ssh", append(args, "ss", "-tlnp")...)
		out3, _ := c3.CombinedOutput()

		// Test 4: netstat (older Linux)
		c4 := exec.Command("ssh", append(args, "netstat", "-tlnp")...)
		out4, _ := c4.CombinedOutput()

		// Try in order: lsof-listen, lsof-broad, ss, netstat
		var output []byte
		if len(out1) > 0 {
			output = out1
		} else if len(out2) > 0 {
			output = out2
		} else if len(out3) > 0 {
			output = out3
		} else if len(out4) > 0 {
			output = out4
		}

		ports := ssh.ParseSSOutput(string(output))
		return portScanMsg{ports: ports}
	}
}

type portScanMsg struct {
	ports []ssh.PortInfo
	err   error
}

type tunnelMsg struct {
	pid        int
	localPort  int
	remotePort int
	err        error
}

type tunnelResetMsg struct{}

func (m *model) createTunnelCmd(server config.SSHConfig, localPort, remotePort int) tea.Cmd {
	// Auto-assign local port if in use
	actualLocalPort := localPort
	if !tunnel.IsPortAvailable(localPort) {
		actualLocalPort = tunnel.FindNextPort()
	}

	args := []string{
		"-N",
		"-L", fmt.Sprintf(":%d:localhost:%d", actualLocalPort, remotePort),
		"-o", "ExitOnForwardFailure=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=9999",
		"-o", "TCPKeepAlive=yes",
		"-p", server.Port,
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	return func() tea.Msg {
		cmd := exec.Command("ssh", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return tunnelMsg{err: fmt.Errorf("failed to start: %w", err)}
		}
		// Record the tunnel
		tunnels, _ := tunnel.LoadState()
		tunnels = append(tunnels, tunnel.Tunnel{
			ID:         fmt.Sprintf("%d", cmd.Process.Pid),
			ServerName: server.Name,
			LocalPort:  actualLocalPort,
			RemotePort: remotePort,
			PID:        cmd.Process.Pid,
		})
		tunnel.SaveState(tunnels)
		// Detach and forget — tunnel runs in background
		go func() {
			_ = cmd.Wait()
		}()
		fmt.Fprintf(os.Stderr, "[TUNNEL] created PID=%d local=%d remote=%d\n", cmd.Process.Pid, actualLocalPort, remotePort)
		return tunnelMsg{pid: cmd.Process.Pid, localPort: actualLocalPort, remotePort: remotePort}
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

func (m *model) stopTunnelForPort(port int) error {
	m.loadTunnels()
	for _, t := range m.tunnels {
		if t.ServerName == m.portsServer.Name && t.RemotePort == port {
			// Kill SSH process by port (more reliable on Windows)
			if err := tunnel.KillProcessByPort(t.LocalPort); err != nil {
				// Fallback to PID-based kill
				fmt.Fprintf(os.Stderr, "[KILL-PORT] %v, falling back to PID=%d\n", err, t.PID)
				tunnel.KillProcess(t.PID)
			}
			// Remove from state (keep other tunnels)
			var remaining []tunnel.Tunnel
			for _, tt := range m.tunnels {
				if tt.PID != t.PID {
					remaining = append(remaining, tt)
				}
			}
			tunnel.SaveState(remaining)
			return nil
		}
	}
	return fmt.Errorf("tunnel not found for port %d", port)
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
		if m.tunnelCreating {
			s.WriteString(m.viewTunnelCreating())
		} else if m.tunnelCreated {
			s.WriteString(m.viewTunnelCreated())
		} else if m.tunnelErr != nil {
			s.WriteString(m.viewTunnelError())
		} else {
			s.WriteString(m.viewPortList())
		}
	case ModeTunnelCreate:
		s.WriteString(m.viewTunnelCreate())
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

	// Render table with border
	tableHeight := min(15, len(m.servers)+2)
	m.table.SetWidth(m.windowWidth - 4)
	m.table.SetHeight(tableHeight)
	s.WriteString(theme.TableBorder.Render(m.table.View()) + "\n")

	s.WriteString(theme.MutedStyle.Render(
		"  ↑↓ "+m.L.T("navigate")+"  / "+m.L.T("filter")+"  enter:"+m.L.T("action")+"  q:"+m.L.T("quit")))

	return s.String()
}

func (m model) viewActionSelect() string {
	server := m.selectedServer
	if server.Name == "" {
		return m.L.T("no_servers") + "\n"
	}
	var s strings.Builder
	s.WriteString(fmt.Sprintf("%s\n\n", theme.HeaderStyle.Render("Server: "+server.Name)))

	// SSH Connect option
	if m.actionIdx == 0 {
		s.WriteString(fmt.Sprintf("  %s\n", theme.SelectedAction.Render("[S] "+m.L.T("ssh_connect")+" — "+m.L.T("ssh_connect_desc"))))
	} else {
		s.WriteString(fmt.Sprintf("  %s\n", theme.UnselectedAction.Render("[S] "+m.L.T("ssh_connect")+" — "+m.L.T("ssh_connect_desc"))))
	}

	// Port Forward option
	if m.actionIdx == 1 {
		s.WriteString(fmt.Sprintf("  %s\n", theme.SelectedAction.Render("[P] "+m.L.T("port_forward")+" — "+m.L.T("port_forward_desc"))))
	} else {
		s.WriteString(fmt.Sprintf("  %s\n", theme.UnselectedAction.Render("[P] "+m.L.T("port_forward")+" — "+m.L.T("port_forward_desc"))))
	}

	s.WriteString("\n")
	s.WriteString(theme.MutedStyle.Render("← back  ↑↓ navigate  enter:execute"))
	return s.String()
}

func (m *model) viewPortList() string {
	server := m.portsServer
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

	s.WriteString(theme.TableBorder.Render(m.table.View()) + "\n\n")

	s.WriteString(theme.MutedStyle.Render(
		"↑↓ "+m.L.T("navigate")+"  enter:"+m.L.T("forward")+"  f:"+m.L.T("forward")+"  esc:"+m.L.T("back")+"  x:stop"))

	return s.String()
}

func (m model) portRows() []table.Row {
	var rows []table.Row
	for _, p := range m.ports {
		tunnelStatus := "-"
		if m.hasTunnelForRemote(p.Port) {
			tunnelStatus = "✓"
		}
		rows = append(rows, table.Row{
			strconv.Itoa(p.Port),
			tunnelStatus,
			p.Protocol,
			p.LocalAddr,
			p.Process,
		})
	}
	return rows
}

func (m model) hasTunnelForRemote(port int) bool {
	for _, t := range m.tunnels {
		if t.ServerName == m.portsServer.Name && t.RemotePort == port {
			return true
		}
	}
	return false
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

func (m model) viewTunnelCreating() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("Server: "+m.portsServer.Name) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")) + "\n\n")
	s.WriteString("  " + m.spinner.View() + " Creating tunnel...\n")
	return s.String()
}

func (m model) viewTunnelCreated() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("Server: "+m.portsServer.Name) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")) + "\n\n")
	s.WriteString(fmt.Sprintf("  %s localhost:%d -> %s:%d\n\n",
		theme.SuccessStyle.Render("✓ Tunnel created"),
		m.tunnelLocalPort, m.portsServer.Name, m.tunnelRemotePort))
	s.WriteString(theme.MutedStyle.Render("←/esc: "+m.L.T("back_to_list")+"  x:stop"))
	return s.String()
}

func (m model) viewTunnelError() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("Server: "+m.portsServer.Name) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")) + "\n\n")
	s.WriteString(fmt.Sprintf("  %s: %v\n\n",
		theme.ErrorStyle.Render("✗ Failed"), m.tunnelErr))
	s.WriteString(theme.MutedStyle.Render("←/esc: "+m.L.T("back_to_list")+"  x:stop"))
	return s.String()
}

func (m *model) viewTunnelManager() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render(m.L.T("active_tunnels")) + "  " +
		theme.MutedStyle.Render("◀ "+m.L.T("back")+"  r:"+m.L.T("refresh")) + "\n\n")

	m.loadTunnels()

	if len(m.tunnels) == 0 {
		s.WriteString("  "+m.L.T("no_tunnels")+"\n")
		return s.String()
	}

	m.table = table.New(
		table.WithColumns([]table.Column{
			{Title: "Local", Width: 12},
			{Title: "Remote", Width: 12},
			{Title: "Server", Width: 20},
			{Title: "PID", Width: 8},
		}),
		table.WithRows(m.tunnelRows()),
		table.WithFocused(true),
	)
	m.table.SetWidth(m.windowWidth - 6)
	m.table.SetHeight(min(15, len(m.tunnels)+2))
	s.WriteString(theme.TableBorder.Render(m.table.View()) + "\n\n")

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
		tunnelCount := 0
		for _, t := range m.tunnels {
			if t.ServerName == m.portsServer.Name {
				tunnelCount++
			}
		}
		status = fmt.Sprintf("%s: %d  %s: %d", m.L.T("ports_count"), len(m.ports), m.L.T("active_tunnels"), tunnelCount)
	case ModeTunnels:
		status = fmt.Sprintf("%s: %d", m.L.T("tunnels_count"), len(m.tunnels))
	default:
		status = ""
	}
	if status != "" {
		status += "  "
	}
	status += m.L.T("locale") + ": " + m.locales[m.localeIdx] + "  L:" + m.L.T("change_locale")
	return theme.StatusBarStyle.Render(status)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
