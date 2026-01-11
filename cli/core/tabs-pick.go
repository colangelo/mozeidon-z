package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/egovelox/mozeidon/browser/core/models"
	"github.com/sahilm/fuzzy"
)

// Styles
var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#ffffff"})

	activeMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#22c55e", Dark: "#4ade80"}).
				Bold(true)

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#60a5fa"}).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"})

	titleStyle = lipgloss.NewStyle()

	domainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"})
)

// pickerModel is the bubbletea model for the tab picker
type pickerModel struct {
	app           *App
	tabs          []models.Tab
	filtered      []models.Tab
	matches       []fuzzy.Match
	cursor        int
	textInput     textinput.Model
	loopMode      bool
	demoMode      bool
	width         int
	height        int
	err           error
	selected      *models.Tab
	shouldQuit    bool
	needsRefresh  bool
}

// tabMatchSource implements fuzzy.Source for tab matching
type tabMatchSource []models.Tab

func (t tabMatchSource) String(i int) string {
	return t[i].Title + " " + t[i].Domain
}

func (t tabMatchSource) Len() int {
	return len(t)
}

// Init initializes the model
func (m pickerModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchTabs)
}

// fetchTabs fetches tabs from the browser
func (m pickerModel) fetchTabs() tea.Msg {
	if m.demoMode {
		return tabsLoadedMsg{tabs: generateDemoTabs()}
	}

	tabs := []models.Tab{}
	for result := range m.app.TabsGet(false, true) {
		tabs = append(tabs, result.Items...)
	}
	// Sort by lastAccessed descending (most recent first)
	sort.Slice(tabs, func(i, j int) bool {
		return tabs[i].LastAccessed > tabs[j].LastAccessed
	})
	return tabsLoadedMsg{tabs: tabs}
}

// generateDemoTabs creates fake tabs for demo/testing
func generateDemoTabs() []models.Tab {
	return []models.Tab{
		{Id: 1, WindowId: 1, Title: "GitHub - charmbracelet/bubbletea", Domain: "github.com", Active: true, LastAccessed: 1000},
		{Id: 2, WindowId: 1, Title: "Google Search - golang tui", Domain: "google.com", Active: false, LastAccessed: 999},
		{Id: 3, WindowId: 1, Title: "Stack Overflow - How to build CLI apps", Domain: "stackoverflow.com", Active: false, LastAccessed: 998},
		{Id: 4, WindowId: 1, Title: "Hacker News", Domain: "news.ycombinator.com", Active: false, LastAccessed: 997},
		{Id: 5, WindowId: 1, Title: "Reddit - r/golang", Domain: "reddit.com", Active: false, LastAccessed: 996},
		{Id: 6, WindowId: 2, Title: "YouTube - Charm CLI Tools Tutorial", Domain: "youtube.com", Active: false, LastAccessed: 995},
		{Id: 7, WindowId: 2, Title: "Twitter / X - @chaborel", Domain: "x.com", Active: false, LastAccessed: 994},
		{Id: 8, WindowId: 2, Title: "Gmail - Inbox", Domain: "mail.google.com", Active: false, LastAccessed: 993},
		{Id: 9, WindowId: 2, Title: "Notion - Project Notes", Domain: "notion.so", Active: false, LastAccessed: 992},
		{Id: 10, WindowId: 2, Title: "Figma - UI Design", Domain: "figma.com", Active: false, LastAccessed: 991},
		{Id: 11, WindowId: 1, Title: "MDN Web Docs - JavaScript", Domain: "developer.mozilla.org", Active: false, LastAccessed: 990},
		{Id: 12, WindowId: 1, Title: "Go Documentation", Domain: "go.dev", Active: false, LastAccessed: 989},
	}
}

type tabsLoadedMsg struct {
	tabs []models.Tab
}

type tabActivatedMsg struct {
	err error
}

// Update handles messages
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tabsLoadedMsg:
		m.tabs = msg.tabs
		m.filterTabs()
		m.needsRefresh = false
		return m, nil

	case tabActivatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.loopMode {
			// Stay open, refresh tabs
			m.needsRefresh = true
			return m, m.fetchTabs
		}
		m.shouldQuit = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.shouldQuit = true
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				tab := m.filtered[m.cursor]
				m.selected = &tab
				return m, m.activateTab(tab)
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "r", "R":
			if !m.textInput.Focused() || msg.String() == "R" {
				m.needsRefresh = true
				return m, m.fetchTabs
			}
		}
	}

	// Handle text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.filterTabs()
	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m, cmd
}

// filterTabs filters tabs based on the current query
func (m *pickerModel) filterTabs() {
	query := strings.TrimSpace(m.textInput.Value())

	var newFiltered []models.Tab
	var newMatches []fuzzy.Match

	if query == "" {
		newFiltered = m.tabs
		newMatches = nil
	} else {
		// Use fuzzy matching
		matches := fuzzy.FindFrom(query, tabMatchSource(m.tabs))
		newFiltered = make([]models.Tab, len(matches))
		newMatches = make([]fuzzy.Match, len(matches))
		for i, match := range matches {
			newFiltered[i] = m.tabs[match.Index]
			newMatches[i] = match
		}
	}

	// Only reset cursor if the filtered list actually changed
	if len(newFiltered) != len(m.filtered) {
		m.cursor = 0
	}

	m.filtered = newFiltered
	m.matches = newMatches
}

// activateTab activates the selected tab
func (m pickerModel) activateTab(tab models.Tab) tea.Cmd {
	return func() tea.Msg {
		tabId := fmt.Sprintf("%d:%d", tab.WindowId, tab.Id)
		m.app.TabsActivate(tabId)
		return tabActivatedMsg{}
	}
}

// View renders the UI
func (m pickerModel) View() string {
	if m.shouldQuit {
		return ""
	}

	var b strings.Builder

	// Search input
	b.WriteString("🔍 ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render(
			fmt.Sprintf("Error: %v", m.err),
		))
		b.WriteString("\n\n")
	}

	// Loading/refresh indicator
	if m.needsRefresh {
		b.WriteString(dimStyle.Render("Refreshing..."))
		b.WriteString("\n\n")
	}

	// Empty state
	if len(m.tabs) == 0 && !m.needsRefresh {
		b.WriteString(dimStyle.Render("No tabs found. Press Esc to exit."))
		b.WriteString("\n")
		return b.String()
	}

	// Tab list
	maxVisible := m.height - 6 // Account for header and margins
	if maxVisible < 5 {
		maxVisible = 10
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := min(start+maxVisible, len(m.filtered))

	for i := start; i < end; i++ {
		tab := m.filtered[i]
		isSelected := i == m.cursor
		isActive := tab.Active

		// Build the line
		var line strings.Builder

		// Active marker
		if isActive {
			line.WriteString(activeMarkerStyle.Render("● "))
		} else {
			line.WriteString("  ")
		}

		// Title (with match highlighting if applicable)
		title := tab.Title
		maxTitleLen := 50
		truncated := len(title) > maxTitleLen
		if truncated {
			title = title[:maxTitleLen-1] + "…"
		}

		if m.matches != nil && i < len(m.matches) {
			// Only use indexes that fall within the visible title
			var titleIndexes []int
			for _, idx := range m.matches[i].MatchedIndexes {
				if idx < len(tab.Title) && idx < maxTitleLen-1 {
					titleIndexes = append(titleIndexes, idx)
				}
			}
			title = highlightMatches(title, titleIndexes)
		}
		line.WriteString(title)

		// Padding - use original title length for calculation
		displayLen := len(tab.Title)
		if displayLen > maxTitleLen {
			displayLen = maxTitleLen
		}
		padding := 55 - displayLen
		if padding < 2 {
			padding = 2
		}
		line.WriteString(strings.Repeat(" ", padding))

		// Domain
		domain := truncate(tab.Domain, 30)
		line.WriteString(domainStyle.Render(domain))

		// Apply selection style
		lineStr := line.String()
		if isSelected {
			lineStr = selectedStyle.Render(lineStr)
		}

		b.WriteString(lineStr)
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf(
		"%d/%d tabs • ↑↓/jk navigate • Enter select • R refresh • Esc quit",
		len(m.filtered), len(m.tabs),
	)))
	if m.loopMode {
		b.WriteString(dimStyle.Render(" • loop mode"))
	}
	b.WriteString("\n")

	return b.String()
}

// highlightMatches highlights matched characters in a string
func highlightMatches(s string, indexes []int) string {
	if len(indexes) == 0 {
		return s
	}

	// Create a set of matched indexes
	matchSet := make(map[int]bool)
	for _, idx := range indexes {
		matchSet[idx] = true
	}

	var result strings.Builder
	for i, r := range s {
		if matchSet[i] {
			result.WriteString(matchStyle.Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TabsPick launches the interactive tab picker
func (a *App) TabsPick(loopMode bool, demoMode bool) error {
	ti := textinput.New()
	ti.Placeholder = "Type to search tabs..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	m := pickerModel{
		app:       a,
		textInput: ti,
		loopMode:  loopMode,
		demoMode:  demoMode,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if a tab was selected
	if fm, ok := finalModel.(pickerModel); ok && fm.selected != nil {
		// Tab was activated successfully
		return nil
	}

	return nil
}
