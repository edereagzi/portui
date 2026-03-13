package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/edereagzi/portui/internal/types"
)

type appState int

const (
	stateLoading appState = iota
	stateTable
	stateDetail
	stateSearch
	stateConfirmKill
	stateHelp
)

type Model struct {
	scanner            types.PortScanner
	processService     types.ProcessService
	entries            []types.PortEntry
	filtered           []types.PortEntry
	table              table.Model
	state              appState
	width              int
	height             int
	err                error
	filterText         string
	filterInput        textinput.Model
	selectedEntry      *types.PortEntry
	confirmImpactPorts []uint16
	detailInfo         *types.ProcessInfo
	statusMsg          string
	statusIsErr        bool
}

func New(scanner types.PortScanner, processService types.ProcessService) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter..."

	return Model{
		scanner:        scanner,
		processService: processService,
		table:          buildTable(nil, 0, 0),
		state:          stateLoading,
		filterInput:    ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanPortsCmd(m.scanner), tickCmd())
}

const confirmDialogLines = 6

func (m *Model) rebuildTable() {
	prevRow := m.table.SelectedRow()
	m.filtered = filteredEntries(m.entries, m.filterText)
	h := m.height
	if m.state == stateConfirmKill {
		h -= confirmDialogLines
	}
	m.table = buildTable(m.filtered, m.width, h)
	if prevRow != nil {
		for i, row := range m.table.Rows() {
			if len(row) >= 3 && len(prevRow) >= 3 && row[0] == prevRow[0] && row[2] == prevRow[2] {
				m.table.SetCursor(i)
				break
			}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portsLoadedMsg:
		m.entries = msg.entries
		m.err = nil
		if m.state == stateLoading {
			m.state = stateTable
		}
		m.rebuildTable()
		return m, nil
	case tickMsg:
		return m, tea.Batch(scanPortsCmd(m.scanner), tickCmd())
	case killResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("✗ Failed: %s", msg.err)
			m.statusIsErr = true
		} else {
			m.statusMsg = fmt.Sprintf("✓ Killed %s (PID %d)", msg.entry.ProcessName, msg.entry.PID)
			m.statusIsErr = false
		}
		m.selectedEntry = nil
		m.confirmImpactPorts = nil
		m.state = stateTable
		m.rebuildTable()
		return m, tea.Batch(scanPortsCmd(m.scanner), statusClearCmd())
	case statusClearMsg:
		m.statusMsg = ""
		m.statusIsErr = false
		return m, nil
	case errMsg:
		m.err = msg.err
		if m.state == stateLoading {
			m.state = stateTable
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildTable()
		return m, nil
	case tea.KeyPressMsg:
		if key.Matches(msg, Keys.Help) {
			if m.state == stateHelp {
				m.state = stateTable
			} else {
				m.state = stateHelp
			}
			return m, nil
		}

		if m.state == stateHelp {
			if key.Matches(msg, Keys.Esc) {
				m.state = stateTable
			}
			return m, nil
		}

		if key.Matches(msg, Keys.Quit) {
			return m, tea.Quit
		}

		if m.state == stateDetail {
			if key.Matches(msg, Keys.Esc) || key.Matches(msg, Keys.Enter) {
				m.state = stateTable
				m.detailInfo = nil
				return m, nil
			}
			if key.Matches(msg, Keys.Kill) && m.selectedEntry != nil {
				m.confirmImpactPorts = impactedPortsByPID(m.entries, m.selectedEntry.PID)
				m.state = stateConfirmKill
				m.rebuildTable()
				return m, nil
			}
			return m, nil
		}

		if m.state == stateTable {
			if key.Matches(msg, Keys.Enter) {
				entry := selectedPortEntry(m.filtered, m.table.SelectedRow())
				if entry != nil {
					info, err := m.processService.GetInfo(entry.PID)
					if err != nil {
						m.statusMsg = fmt.Sprintf("✗ Failed to get process info: %s", err.Error())
						m.statusIsErr = true
						return m, statusClearCmd()
					}
					entryCopy := *entry
					m.selectedEntry = &entryCopy
					m.detailInfo = info
					m.state = stateDetail
				}
				return m, nil
			}
			if key.Matches(msg, Keys.Search) {
				m.state = stateSearch
				cmd := m.filterInput.Focus()
				return m, cmd
			}
			if key.Matches(msg, Keys.Kill) {
				entry := selectedPortEntry(m.filtered, m.table.SelectedRow())
				if entry != nil {
					entryCopy := *entry
					m.selectedEntry = &entryCopy
					m.confirmImpactPorts = impactedPortsByPID(m.entries, entryCopy.PID)
					m.state = stateConfirmKill
					m.rebuildTable()
				}
				return m, nil
			}
			if key.Matches(msg, Keys.Refresh) {
				return m, scanPortsCmd(m.scanner)
			}
			if key.Matches(msg, Keys.Up) || key.Matches(msg, Keys.Down) {
				updatedTable, cmd := m.table.Update(msg)
				m.table = updatedTable
				return m, cmd
			}
		}

		if m.state == stateConfirmKill {
			if key.Matches(msg, Keys.Esc) || msg.Text == "n" {
				m.selectedEntry = nil
				m.confirmImpactPorts = nil
				m.state = stateTable
				m.rebuildTable()
				return m, nil
			}
			if msg.Text == "y" {
				entry := *m.selectedEntry
				m.statusMsg = fmt.Sprintf("Killing %s (PID %d)…", entry.ProcessName, entry.PID)
				m.statusIsErr = false
				return m, killProcessCmd(m.processService, entry)
			}
		}

		if m.state == stateSearch {
			if key.Matches(msg, Keys.Esc) {
				m.state = stateTable
				m.filterText = ""
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.rebuildTable()
				return m, nil
			}
			if key.Matches(msg, Keys.Enter) {
				m.state = stateTable
				m.filterInput.Blur()
				return m, nil
			}
			updatedTI, cmd := m.filterInput.Update(msg)
			m.filterInput = updatedTI
			m.filterText = m.filterInput.Value()
			m.rebuildTable()
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() tea.View {
	if m.width > 0 && m.height > 0 && (m.width < 60 || m.height < 15) {
		view := tea.NewView(EmptyStateStyle.Render("Terminal too small. Minimum: 60×15"))
		view.AltScreen = true
		return view
	}

	var content string
	switch m.state {
	case stateTable, stateSearch, stateConfirmKill:
		content = m.renderTableView(m.width)
	case stateDetail:
		content = m.detailView()
	case stateHelp:
		content = m.helpOverlayView()
	default:
		content = m.loadingView()
	}

	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m Model) helpOverlayView() string {
	header := AppTitle
	helpDialog := renderHelpOverlay(0, 0)
	footer := MutedStyle.Render(m.footerHints())
	return lipgloss.JoinVertical(lipgloss.Left, header, helpDialog, footer)
}

func (m Model) detailView() string {
	header := AppTitle
	detail := renderDetailPanel(m.detailInfo, m.selectedEntry, m.width)
	footer := MutedStyle.Render(m.footerHints())

	top := lipgloss.JoinVertical(lipgloss.Left, header, detail)

	if m.height > 0 {
		gap := m.height - lipgloss.Height(top) - lipgloss.Height(footer)
		if gap > 0 {
			top += strings.Repeat("\n", gap)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, top, footer)
}

func (m Model) loadingView() string {
	content := EmptyStateStyle.Render("Loading ports...")
	if m.err != nil {
		content = lipgloss.JoinVertical(lipgloss.Left, content, ErrorStyle.Render(fmt.Sprintf("error: %v", m.err)))
	}

	if m.width <= 0 || m.height <= 0 {
		return content
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func (m Model) renderTableView(width int) string {
	body := m.table.View()
	if len(m.filtered) == 0 && m.filterText != "" {
		body = EmptyStateStyle.Render("No matching ports.")
	} else if len(m.entries) == 0 {
		body = EmptyStateStyle.Render("No listening TCP ports found.\nTip: Start a server and it will appear here.")
	}

	renderWidth := width
	if renderWidth <= 0 {
		renderWidth = m.width
	}

	header := AppTitle
	statusBar := renderStatusBar(m.statusLeft(), "", renderWidth)
	footer := MutedStyle.Render(m.footerHints())

	topParts := []string{header}
	if m.state == stateSearch {
		topParts = append(topParts, SearchInputStyle.Render(m.filterInput.View()))
	}
	topParts = append(topParts, body)
	if m.state == stateConfirmKill {
		topParts = append(topParts, confirmKillView(m.selectedEntry, m.confirmImpactPorts, renderWidth))
	}

	top := lipgloss.JoinVertical(lipgloss.Left, topParts...)
	bottom := lipgloss.JoinVertical(lipgloss.Left, statusBar, footer)

	if m.height > 0 {
		gap := m.height - lipgloss.Height(top) - lipgloss.Height(bottom)
		if gap > 0 {
			top += strings.Repeat("\n", gap)
		}
	}

	view := lipgloss.JoinVertical(lipgloss.Left, top, bottom)
	if width > 0 {
		return lipgloss.NewStyle().Width(width).Render(view)
	}
	return view
}

func (m Model) footerHints() string {
	switch m.state {
	case stateSearch:
		return "enter: apply filter  |  esc: clear & close  |  q: quit"
	case stateConfirmKill:
		return "y: confirm kill  |  n/esc: cancel"
	case stateHelp:
		return "?/esc: close help"
	case stateDetail:
		return "x: kill  |  esc: back  |  ?: help  |  q: quit"
	default:
		return "↑/↓: navigate  |  enter: detail  |  x: kill  |  /: search  |  ?: help  |  q: quit"
	}
}

func (m Model) statusLeft() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("⚠ Scan failed — %d ports (stale)", len(m.entries)))
	}
	if m.statusMsg != "" {
		style := SuccessStyle
		if m.statusIsErr {
			style = ErrorStyle
		}
		return style.Render(m.statusMsg)
	}
	if m.filterText != "" {
		return fmt.Sprintf("Filter: %q (%d of %d)", m.filterText, len(m.filtered), len(m.entries))
	}
	return fmt.Sprintf("%d ports", len(m.filtered))
}

func renderStatusBar(left, right string, width int) string {
	if width <= 0 {
		content := strings.TrimSpace(strings.Join([]string{left, right}, " "))
		return StatusBarStyle.Render(content)
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = width
	}

	rightWidth := lipgloss.Width(right)
	leftLimit := innerWidth - rightWidth - 1
	if leftLimit < 0 {
		leftLimit = 0
	}
	left = truncateStatus(left, leftLimit)

	gap := innerWidth - lipgloss.Width(left) - rightWidth
	if gap < 1 {
		gap = 1
	}

	return StatusBarStyle.Width(width).Render(left + strings.Repeat(" ", gap) + right)
}

func truncateStatus(s string, maxWidth int) string {
	if maxWidth <= 0 || lipgloss.Width(s) <= maxWidth {
		return s
	}

	if maxWidth == 1 {
		return "…"
	}

	runes := []rune(s)
	for len(runes) > 0 {
		candidate := string(runes) + "…"
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}

	return "…"
}
