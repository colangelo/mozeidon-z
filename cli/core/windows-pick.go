package core

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/egovelox/mozeidon/browser/core/models"
	"github.com/sahilm/fuzzy"
)

// windowPickerModel is the bubbletea model for the window picker. It mirrors
// pickerModel (tabs-pick.go) and reuses its shared styles/helpers.
type windowPickerModel struct {
	app          *App
	windows      []models.Window
	filtered     []models.Window
	matches      []fuzzy.Match
	cursor       int
	textInput    textinput.Model
	loopMode     bool
	demoMode     bool
	width        int
	height       int
	err          error
	selected     *models.Window
	shouldQuit   bool
	needsRefresh bool
}

// windowMatchSource implements fuzzy.Source, matching on the active tab title,
// its host, and the window id (so "3289" or "github" both find a window).
type windowMatchSource []models.Window

func (w windowMatchSource) String(i int) string {
	return w[i].ActiveTabTitle + " " + hostOf(w[i].ActiveTabUrl) + " " + fmt.Sprintf("window %d", w[i].Id)
}

func (w windowMatchSource) Len() int {
	return len(w)
}

func hostOf(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	if u, err := url.Parse(rawURL); err == nil {
		return u.Hostname()
	}
	return ""
}

func (m windowPickerModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchWindows)
}

func (m windowPickerModel) fetchWindows() tea.Msg {
	if m.demoMode {
		return windowsLoadedMsg{windows: generateDemoWindows()}
	}

	windows := []models.Window{}
	for result := range m.app.WindowsGet() {
		windows = append(windows, result.Items...)
	}
	// Last-focused window first, then by id (windows have no lastAccessed).
	sort.Slice(windows, func(i, j int) bool {
		if windows[i].IsLastFocused != windows[j].IsLastFocused {
			return windows[i].IsLastFocused
		}
		return windows[i].Id < windows[j].Id
	})
	return windowsLoadedMsg{windows: windows}
}

func generateDemoWindows() []models.Window {
	return []models.Window{
		{Id: 1, IsLastFocused: true, TabCount: 12, ActiveTabTitle: "GitHub - charmbracelet/bubbletea", ActiveTabUrl: "https://github.com/charmbracelet/bubbletea", State: "normal", Type: "normal"},
		{Id: 2, TabCount: 5, ActiveTabTitle: "Gmail - Inbox", ActiveTabUrl: "https://mail.google.com", State: "normal", Type: "normal"},
		{Id: 3, TabCount: 1, Incognito: true, ActiveTabTitle: "DuckDuckGo — Privacy", ActiveTabUrl: "https://duckduckgo.com", State: "minimized", Type: "normal"},
		{Id: 4, TabCount: 8, ActiveTabTitle: "YouTube", ActiveTabUrl: "https://youtube.com", State: "maximized", Type: "normal"},
	}
}

type windowsLoadedMsg struct {
	windows []models.Window
}

type windowFocusedMsg struct {
	err error
}

func (m windowPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case windowsLoadedMsg:
		m.windows = msg.windows
		m.filterWindows()
		m.needsRefresh = false
		return m, nil

	case windowFocusedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.loopMode {
			m.needsRefresh = true
			return m, m.fetchWindows
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
				win := m.filtered[m.cursor]
				m.selected = &win
				return m, m.focusWindow(win)
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
				return m, m.fetchWindows
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.filterWindows()
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m, cmd
}

func (m *windowPickerModel) filterWindows() {
	query := strings.TrimSpace(m.textInput.Value())

	var newFiltered []models.Window
	var newMatches []fuzzy.Match

	if query == "" {
		newFiltered = m.windows
		newMatches = nil
	} else {
		matches := fuzzy.FindFrom(query, windowMatchSource(m.windows))
		newFiltered = make([]models.Window, len(matches))
		newMatches = make([]fuzzy.Match, len(matches))
		for i, match := range matches {
			newFiltered[i] = m.windows[match.Index]
			newMatches[i] = match
		}
	}

	if len(newFiltered) != len(m.filtered) {
		m.cursor = 0
	}

	m.filtered = newFiltered
	m.matches = newMatches
}

func (m windowPickerModel) focusWindow(win models.Window) tea.Cmd {
	return func() tea.Msg {
		m.app.WindowFocus(fmt.Sprintf("%d", win.Id))
		return windowFocusedMsg{}
	}
}

func (m windowPickerModel) View() string {
	if m.shouldQuit {
		return ""
	}

	var b strings.Builder

	b.WriteString("🔍 ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")).Render(
			fmt.Sprintf("Error: %v", m.err),
		))
		b.WriteString("\n\n")
	}

	if m.needsRefresh {
		b.WriteString(dimStyle.Render("Refreshing..."))
		b.WriteString("\n\n")
	}

	if len(m.windows) == 0 && !m.needsRefresh {
		b.WriteString(dimStyle.Render("No windows found. Press Esc to exit."))
		b.WriteString("\n")
		return b.String()
	}

	maxVisible := m.height - 6
	if maxVisible < 5 {
		maxVisible = 10
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := min(start+maxVisible, len(m.filtered))

	for i := start; i < end; i++ {
		win := m.filtered[i]
		isSelected := i == m.cursor

		var line strings.Builder

		// Last-focused marker
		if win.IsLastFocused {
			line.WriteString(activeMarkerStyle.Render("● "))
		} else {
			line.WriteString("  ")
		}

		// Active tab title (with match highlighting)
		title := win.ActiveTabTitle
		if title == "" {
			title = "(no active tab)"
		}
		maxTitleLen := 50
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-1] + "…"
		}
		if m.matches != nil && i < len(m.matches) {
			var titleIndexes []int
			for _, idx := range m.matches[i].MatchedIndexes {
				if idx < len(win.ActiveTabTitle) && idx < maxTitleLen-1 {
					titleIndexes = append(titleIndexes, idx)
				}
			}
			title = highlightMatches(title, titleIndexes)
		}
		line.WriteString(title)

		// Padding based on the untruncated title length
		displayLen := len(win.ActiveTabTitle)
		if win.ActiveTabTitle == "" {
			displayLen = len("(no active tab)")
		}
		if displayLen > maxTitleLen {
			displayLen = maxTitleLen
		}
		padding := 55 - displayLen
		if padding < 2 {
			padding = 2
		}
		line.WriteString(strings.Repeat(" ", padding))

		// Meta: tab count · window id · flags
		meta := fmt.Sprintf("%d tabs · win:%d", win.TabCount, win.Id)
		if win.Incognito {
			meta += " · private"
		}
		if win.State == "minimized" || win.State == "maximized" || win.State == "fullscreen" {
			meta += " · " + win.State
		}
		line.WriteString(domainStyle.Render(meta))

		lineStr := line.String()
		if isSelected {
			lineStr = selectedStyle.Render(lineStr)
		}

		b.WriteString(lineStr)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf(
		"%d/%d windows • ↑↓/jk navigate • Enter focus • R refresh • Esc quit",
		len(m.filtered), len(m.windows),
	)))
	if m.loopMode {
		b.WriteString(dimStyle.Render(" • loop mode"))
	}
	b.WriteString("\n")

	return b.String()
}

// WindowsPick launches the interactive window picker.
func (a *App) WindowsPick(loopMode bool, demoMode bool) error {
	ti := textinput.New()
	ti.Placeholder = "Type to search windows..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	m := windowPickerModel{
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

	if fm, ok := finalModel.(windowPickerModel); ok && fm.selected != nil {
		return nil
	}

	return nil
}
