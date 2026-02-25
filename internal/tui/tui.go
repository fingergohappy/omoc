package tui

import (
	"strings"

	"omoc/internal/config"
	"omoc/internal/models"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type pane int

const (
	paneLeft pane = iota
	paneMiddle
	paneRight
)

type leftItem struct {
	name     string
	isAgent  bool
	isHeader bool
}

type modelsLoadedMsg struct {
	models []string
	err    error
}

type Model struct {
	cfg            *config.Config
	allModels      []string
	filteredModels []string
	loading        bool
	loadErr        string

	focus pane

	leftItems  []leftItem
	leftCursor int
	midCursor  int

	filter  textinput.Model
	spinner spinner.Model

	width  int
	height int

	saved   bool
	message string
}

func buildLeftItems() []leftItem {
	var items []leftItem
	items = append(items, leftItem{name: "── Agents ──", isHeader: true})
	for _, name := range config.KnownAgents {
		items = append(items, leftItem{name: name, isAgent: true})
	}
	items = append(items, leftItem{name: "── Categories ──", isHeader: true})
	for _, name := range config.KnownCategories {
		items = append(items, leftItem{name: name, isAgent: false})
	}
	return items
}

func fetchModelsCmd() tea.Msg {
	m, err := models.Fetch()
	return modelsLoadedMsg{models: m, err: err}
}

func New(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "filter models..."
	ti.CharLimit = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	items := buildLeftItems()
	cursor := 0
	for i, item := range items {
		if !item.isHeader {
			cursor = i
			break
		}
	}

	return Model{
		cfg:        cfg,
		loading:    true,
		leftItems:  items,
		leftCursor: cursor,
		filter:     ti,
		spinner:    sp,
		width:      120,
		height:     40,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchModelsCmd)
}

func (m Model) currentItem() leftItem {
	if m.leftCursor < len(m.leftItems) {
		return m.leftItems[m.leftCursor]
	}
	return leftItem{}
}

func (m Model) currentEntry() *config.AgentEntry {
	item := m.currentItem()
	if item.isHeader {
		return nil
	}
	if item.isAgent {
		return m.cfg.Agents[item.name]
	}
	return m.cfg.Categories[item.name]
}

func (m Model) currentInfo() *config.ItemInfo {
	item := m.currentItem()
	if item.isHeader {
		return nil
	}
	if item.isAgent {
		if info, ok := config.AgentInfo[item.name]; ok {
			return &info
		}
	} else {
		if info, ok := config.CategoryInfo[item.name]; ok {
			return &info
		}
	}
	return nil
}

func (m *Model) sortModelsCurrentFirst(models []string) []string {
	current := ""
	if entry := m.currentEntry(); entry != nil {
		current = entry.Model
	}
	if current == "" {
		return models
	}
	sorted := make([]string, 0, len(models))
	for _, model := range models {
		if model == current {
			sorted = append([]string{model}, sorted...)
		} else {
			sorted = append(sorted, model)
		}
	}
	return sorted
}

func (m *Model) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		m.filteredModels = m.allModels
		return
	}
	var filtered []string
	for _, model := range m.allModels {
		if strings.Contains(strings.ToLower(model), q) {
			filtered = append(filtered, model)
		}
	}
	m.filteredModels = filtered
	if m.midCursor >= len(m.filteredModels) {
		m.midCursor = max(0, len(m.filteredModels)-1)
	}
}

func (m *Model) moveLeftCursor(delta int) {
	next := m.leftCursor + delta
	for next >= 0 && next < len(m.leftItems) {
		if !m.leftItems[next].isHeader {
			m.leftCursor = next
			return
		}
		next += delta
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case modelsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
		} else {
			m.allModels = msg.models
			m.filteredModels = msg.models
		}
		return m, nil

	case tea.KeyMsg:
		if m.focus == paneMiddle && m.filter.Focused() {
			if msg.String() == "enter" || msg.String() == "esc" {
				m.filter.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			m.applyFilter()
			return m, cmd
		}

		m.saved = false
		m.message = ""

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			switch m.focus {
			case paneLeft:
				m.focus = paneMiddle
			case paneMiddle:
				m.focus = paneLeft
			}

		case "h", "left":
			if m.focus == paneMiddle {
				m.focus = paneLeft
			}

		case "l", "right":
			if m.focus == paneLeft {
				m.focus = paneMiddle
			}

		case "j", "down":
			if m.focus == paneLeft {
				m.moveLeftCursor(1)
			} else if m.focus == paneMiddle {
				if m.midCursor < len(m.filteredModels)-1 {
					m.midCursor++
				}
			}

		case "k", "up":
			if m.focus == paneLeft {
				m.moveLeftCursor(-1)
			} else if m.focus == paneMiddle {
				if m.midCursor > 0 {
					m.midCursor--
				}
			}

		case "enter":
			if m.focus == paneMiddle && len(m.filteredModels) > 0 {
				entry := m.currentEntry()
				if entry != nil {
					entry.Model = m.filteredModels[m.midCursor]
					m.message = "model set: " + entry.Model
				}
			}

		case "v":
			if m.focus == paneLeft {
				entry := m.currentEntry()
				if entry != nil {
					idx := 0
					for i, v := range config.KnownVariants {
						if v == entry.Variant {
							idx = i
							break
						}
					}
					idx = (idx + 1) % len(config.KnownVariants)
					entry.Variant = config.KnownVariants[idx]
					if entry.Variant == "" {
						m.message = "variant cleared"
					} else {
						m.message = "variant set: " + entry.Variant
					}
				}
			}

		case "d":
			if m.focus == paneLeft {
				entry := m.currentEntry()
				if entry != nil {
					entry.Model = ""
					entry.Variant = ""
					m.message = "config cleared"
				}
			}

		case "/":
			if m.focus == paneMiddle {
				cmd := m.filter.Focus()
				return m, cmd
			}

		case "esc":
			m.filter.SetValue("")
			m.applyFilter()
			m.filter.Blur()

		case "s", "ctrl+s":
			if err := m.cfg.Save(); err != nil {
				m.message = "save error: " + err.Error()
			} else {
				m.saved = true
				m.message = "saved!"
			}

		case "r":
			if !m.loading {
				m.loading = true
				m.allModels = nil
				m.filteredModels = nil
				m.midCursor = 0
				m.message = "refreshing models..."
				return m, tea.Batch(m.spinner.Tick, fetchModelsCmd)
			}
		}
	}

	return m, nil
}
