package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josesojo2828-tech/docker-tui/internal/docker"
)

type activeTab int

const (
	tabContainers activeTab = iota
	tabVolumes
	tabNetworks
)

type tickMsg time.Time

type dataLoadedMsg struct {
	summary    *docker.SystemSummary
	containers []docker.ContainerItem
	volumes    []docker.VolumeItem
	networks   []docker.NetworkItem
	err        error
}

type actionResultMsg struct {
	message string
	err     error
}

// Model represents the Bubbletea UI state
type Model struct {
	dockerSvc       *docker.Service
	currentTab      activeTab
	showAll         bool
	containersTable table.Model
	volumesTable    table.Model
	networksTable   table.Model

	summary        *docker.SystemSummary
	containersData []docker.ContainerItem
	volumesData    []docker.VolumeItem
	networksData   []docker.NetworkItem

	width         int
	height        int
	statusMsg     string
	isErrorStatus bool
	loading       bool
	webPort       int
}

// NewModel creates an initialized TUI model
func NewModel(dockerSvc *docker.Service, webPort int) Model {
	// Containers Table
	ctCols := []table.Column{
		{Title: "ESTADO", Width: 12},
		{Title: "NOMBRE", Width: 26},
		{Title: "ID", Width: 14},
		{Title: "IMAGEN", Width: 28},
		{Title: "PUERTOS", Width: 28},
		{Title: "STATUS", Width: 24},
	}
	ct := table.New(
		table.WithColumns(ctCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Volumes Table
	vtCols := []table.Column{
		{Title: "NOMBRE", Width: 32},
		{Title: "DRIVER", Width: 12},
		{Title: "SCOPE", Width: 10},
		{Title: "PUNTO DE MONTAJE", Width: 60},
	}
	vt := table.New(
		table.WithColumns(vtCols),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	// Networks Table
	ntCols := []table.Column{
		{Title: "ID", Width: 14},
		{Title: "NOMBRE", Width: 28},
		{Title: "DRIVER", Width: 14},
		{Title: "SCOPE", Width: 12},
		{Title: "CONTENEDORES", Width: 16},
	}
	nt := table.New(
		table.WithColumns(ntCols),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	// Table styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#93c5fd"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#1d4ed8")).
		Bold(true)

	ct.SetStyles(s)
	vt.SetStyles(s)
	nt.SetStyles(s)

	return Model{
		dockerSvc:       dockerSvc,
		currentTab:      tabContainers,
		showAll:         true,
		containersTable: ct,
		volumesTable:    vt,
		networksTable:   nt,
		loading:         true,
		webPort:         webPort,
	}
}

// Init sets up initial commands (data fetching and periodic tick)
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchDataCmd(),
		m.tickCmd(),
	)
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sum, err := m.dockerSvc.GetSummary(ctx)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		conts, err := m.dockerSvc.ListContainers(ctx, m.showAll)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		vols, err := m.dockerSvc.ListVolumes(ctx)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		nets, err := m.dockerSvc.ListNetworks(ctx)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		return dataLoadedMsg{
			summary:    sum,
			containers: conts,
			volumes:    vols,
			networks:   nets,
		}
	}
}

func (m Model) containerActionCmd(action, containerID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		switch action {
		case "start":
			err = m.dockerSvc.StartContainer(ctx, containerID)
		case "stop":
			err = m.dockerSvc.StopContainer(ctx, containerID)
		case "restart":
			err = m.dockerSvc.RestartContainer(ctx, containerID)
		}

		if err != nil {
			return actionResultMsg{
				message: fmt.Sprintf("Error ejecutando %s: %v", action, err),
				err:     err,
			}
		}
		return actionResultMsg{
			message: fmt.Sprintf("Contenedor %s ejecutado: %s", containerID, action),
		}
	}
}

// Update handles incoming messages and keypresses
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		tableHeight := msg.Height - 12
		if tableHeight < 5 {
			tableHeight = 5
		}
		m.containersTable.SetHeight(tableHeight)
		m.volumesTable.SetHeight(tableHeight)
		m.networksTable.SetHeight(tableHeight)

	case tickMsg:
		cmds = append(cmds, m.fetchDataCmd(), m.tickCmd())

	case dataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error actualizando: %v", msg.err)
			m.isErrorStatus = true
			return m, nil
		}
		m.summary = msg.summary
		m.containersData = msg.containers
		m.volumesData = msg.volumes
		m.networksData = msg.networks

		// Update rows
		var cRows []table.Row
		for _, c := range msg.containers {
			stateEmoji := "🟢"
			if strings.ToLower(c.State) != "running" {
				stateEmoji = "⚪"
			}
			cRows = append(cRows, table.Row{
				fmt.Sprintf("%s %s", stateEmoji, c.State),
				c.Name,
				c.ID,
				c.Image,
				c.Ports,
				c.Status,
			})
		}
		m.containersTable.SetRows(cRows)

		var vRows []table.Row
		for _, v := range msg.volumes {
			vRows = append(vRows, table.Row{
				v.Name,
				v.Driver,
				v.Scope,
				v.Mountpoint,
			})
		}
		m.volumesTable.SetRows(vRows)

		var nRows []table.Row
		for _, n := range msg.networks {
			nRows = append(nRows, table.Row{
				n.ID,
				n.Name,
				n.Driver,
				n.Scope,
				fmt.Sprintf("%d contenedores", n.Containers),
			})
		}
		m.networksTable.SetRows(nRows)

	case actionResultMsg:
		m.statusMsg = msg.message
		m.isErrorStatus = msg.err != nil
		// Refresh list after action
		cmds = append(cmds, m.fetchDataCmd())

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.currentTab = (m.currentTab + 1) % 3
			m.updateTableFocus()

		case "shift+tab":
			m.currentTab = (m.currentTab + 2) % 3
			m.updateTableFocus()

		case "1":
			m.currentTab = tabContainers
			m.updateTableFocus()

		case "2":
			m.currentTab = tabVolumes
			m.updateTableFocus()

		case "3":
			m.currentTab = tabNetworks
			m.updateTableFocus()

		case "a":
			if m.currentTab == tabContainers {
				m.showAll = !m.showAll
				m.loading = true
				cmds = append(cmds, m.fetchDataCmd())
			}

		case "R":
			m.loading = true
			m.statusMsg = "Actualizando datos..."
			m.isErrorStatus = false
			cmds = append(cmds, m.fetchDataCmd())

		case "s": // Start
			if m.currentTab == tabContainers && len(m.containersData) > 0 {
				cursor := m.containersTable.Cursor()
				if cursor >= 0 && cursor < len(m.containersData) {
					c := m.containersData[cursor]
					m.statusMsg = fmt.Sprintf("Iniciando contenedor %s...", c.Name)
					m.isErrorStatus = false
					cmds = append(cmds, m.containerActionCmd("start", c.FullID))
				}
			}

		case "x": // Stop
			if m.currentTab == tabContainers && len(m.containersData) > 0 {
				cursor := m.containersTable.Cursor()
				if cursor >= 0 && cursor < len(m.containersData) {
					c := m.containersData[cursor]
					m.statusMsg = fmt.Sprintf("Deteniendo contenedor %s...", c.Name)
					m.isErrorStatus = false
					cmds = append(cmds, m.containerActionCmd("stop", c.FullID))
				}
			}

		case "r": // Restart
			if m.currentTab == tabContainers && len(m.containersData) > 0 {
				cursor := m.containersTable.Cursor()
				if cursor >= 0 && cursor < len(m.containersData) {
					c := m.containersData[cursor]
					m.statusMsg = fmt.Sprintf("Reiniciando contenedor %s...", c.Name)
					m.isErrorStatus = false
					cmds = append(cmds, m.containerActionCmd("restart", c.FullID))
				}
			}
		}
	}

	// Route keyboard events to active table
	switch m.currentTab {
	case tabContainers:
		m.containersTable, cmd = m.containersTable.Update(msg)
		cmds = append(cmds, cmd)
	case tabVolumes:
		m.volumesTable, cmd = m.volumesTable.Update(msg)
		cmds = append(cmds, cmd)
	case tabNetworks:
		m.networksTable, cmd = m.networksTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateTableFocus() {
	m.containersTable.Blur()
	m.volumesTable.Blur()
	m.networksTable.Blur()

	switch m.currentTab {
	case tabContainers:
		m.containersTable.Focus()
	case tabVolumes:
		m.volumesTable.Focus()
	case tabNetworks:
		m.networksTable.Focus()
	}
}

// View renders the terminal interface
func (m Model) View() string {
	var s strings.Builder

	// Header Bar
	headerTitle := titleStyle.Render(" 🐳 DOCKER TUI & DASHBOARD ")
	webInfo := ""
	if m.webPort > 0 {
		webInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#10b981")).Render(fmt.Sprintf(" Web: http://localhost:%d ", m.webPort))
	}
	daemonVer := ""
	if m.summary != nil {
		daemonVer = lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf(" Docker %s ", m.summary.DockerVersion))
	}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, headerTitle, daemonVer, webInfo) + "\n\n")

	// Summary Cards
	if m.summary != nil {
		cardRunning := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("EN EJECUCIÓN"), lipgloss.NewStyle().Foreground(successColor).Bold(true).Render(fmt.Sprintf("%d", m.summary.RunningContainers))))
		cardStopped := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("DETENIDOS"), lipgloss.NewStyle().Foreground(warningColor).Bold(true).Render(fmt.Sprintf("%d", m.summary.StoppedContainers+m.summary.PausedContainers))))
		cardTotal := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("TOTAL CONT."), statValueStyle.Render(fmt.Sprintf("%d", m.summary.TotalContainers))))
		cardImages := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("IMÁGENES"), lipgloss.NewStyle().Foreground(secondaryColor).Bold(true).Render(fmt.Sprintf("%d", m.summary.TotalImages))))
		cardVolumes := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("VOLÚMENES"), statValueStyle.Render(fmt.Sprintf("%d", m.summary.TotalVolumes))))
		cardNetworks := statBoxStyle.Render(fmt.Sprintf("%s\n%s", statLabelStyle.Render("REDES"), statValueStyle.Render(fmt.Sprintf("%d", m.summary.TotalNetworks))))

		cards := lipgloss.JoinHorizontal(lipgloss.Top, cardRunning, " ", cardStopped, " ", cardTotal, " ", cardImages, " ", cardVolumes, " ", cardNetworks)
		s.WriteString(cards + "\n\n")
	}

	// Tabs Header
	tab1 := tabInactiveStyle.Render("1. Contenedores")
	tab2 := tabInactiveStyle.Render("2. Volúmenes")
	tab3 := tabInactiveStyle.Render("3. Redes")

	switch m.currentTab {
	case tabContainers:
		tab1 = tabActiveStyle.Render("1. Contenedores")
	case tabVolumes:
		tab2 = tabActiveStyle.Render("2. Volúmenes")
	case tabNetworks:
		tab3 = tabActiveStyle.Render("3. Redes")
	}

	filterMode := ""
	if m.currentTab == tabContainers {
		if m.showAll {
			filterMode = lipgloss.NewStyle().Foreground(mutedColor).Render(" [Mostrando: Todos (a para filtrar)]")
		} else {
			filterMode = lipgloss.NewStyle().Foreground(successColor).Render(" [Mostrando: Sólo Activos (a para filtrar)]")
		}
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Center, tab1, tab2, tab3, filterMode)
	s.WriteString(tabsRow + "\n\n")

	// Table Body
	switch m.currentTab {
	case tabContainers:
		s.WriteString(m.containersTable.View())
	case tabVolumes:
		s.WriteString(m.volumesTable.View())
	case tabNetworks:
		s.WriteString(m.networksTable.View())
	}

	// Status Message
	s.WriteString("\n")
	if m.statusMsg != "" {
		if m.isErrorStatus {
			s.WriteString(statusErrStyle.Render("● "+m.statusMsg) + "\n")
		} else {
			s.WriteString(statusMsgStyle.Render("✔ "+m.statusMsg) + "\n")
		}
	} else if m.loading {
		s.WriteString(lipgloss.NewStyle().Foreground(warningColor).Render("⟳ Cargando datos de Docker...") + "\n")
	}

	// Help Bar
	help := fmt.Sprintf(
		"%s: Navegar | %s: Cambiar pestaña | %s: Iniciar | %s: Detener | %s: Reiniciar | %s: Alternar filtro | %s: Refrescar | %s: Salir",
		helpKeyStyle.Render("↑/↓"),
		helpKeyStyle.Render("Tab"),
		helpKeyStyle.Render("s"),
		helpKeyStyle.Render("x"),
		helpKeyStyle.Render("r"),
		helpKeyStyle.Render("a"),
		helpKeyStyle.Render("R"),
		helpKeyStyle.Render("q"),
	)
	s.WriteString(helpBarStyle.Render(help))

	return s.String()
}
