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

type profilesLoadedMsg struct {
	profiles []string
	err      error
}

type profileActivatedMsg struct {
	cfg     *config.Config
	profile string
	err     error
}

type Model struct {
	cfg            *config.Config
	allModels      []string
	filteredModels []string
	loading        bool
	loadErr        string

	profiles        []string
	profileCursor   int
	showProfiles    bool
	activeProfile   string
	creatingProfile bool
	profileInput    textinput.Model

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

func fetchProfilesCmd() tea.Msg {
	p, err := config.ListProfiles()
	return profilesLoadedMsg{profiles: p, err: err}
}

func activateProfileCmd(profile string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadProfile(profile)
		if err != nil {
			return profileActivatedMsg{profile: profile, err: err}
		}
		if err := cfg.Save(); err != nil {
			return profileActivatedMsg{profile: profile, err: err}
		}
		return profileActivatedMsg{cfg: cfg, profile: profile}
	}
}

func createProfileCmd(cfg *config.Config, name string) tea.Cmd {
	return func() tea.Msg {
		filename, err := cfg.CloneToNewProfile(name)
		if err != nil {
			return profileActivatedMsg{err: err}
		}
		return profileActivatedMsg{cfg: cfg, profile: filename}
	}
}

func fetchModelsCmd() tea.Msg {
	m, err := models.Fetch()
	return modelsLoadedMsg{models: m, err: err}
}

func New(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "filter models..."
	ti.CharLimit = 80

	pi := textinput.New()
	pi.Placeholder = "profile name (a-z0-9_-)"
	pi.CharLimit = 32

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
		cfg:           cfg,
		loading:       true,
		leftItems:     items,
		leftCursor:    cursor,
		filter:        ti,
		profileInput:  pi,
		spinner:       sp,
		activeProfile: cfg.ActiveProfileFile,
		width:         120,
		height:        40,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchModelsCmd, fetchProfilesCmd)
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

func (m Model) displayedModels() []string {
	return m.sortModelsCurrentFirst(m.filteredModels)
}

func (m Model) sortModelsCurrentFirst(models []string) []string {
	current := ""
	if entry := m.currentEntry(); entry != nil {
		current = entry.Model
	}
	if current == "" {
		return models
	}

	idx := -1
	for i, model := range models {
		if model == current {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return models
	}

	sorted := make([]string, 0, len(models))
	sorted = append(sorted, current)
	sorted = append(sorted, models[:idx]...)
	sorted = append(sorted, models[idx+1:]...)
	return sorted
}

func (m *Model) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		m.filteredModels = m.allModels
	} else {
		var filtered []string
		for _, model := range m.allModels {
			if strings.Contains(strings.ToLower(model), q) {
				filtered = append(filtered, model)
			}
		}
		m.filteredModels = filtered
	}

	visible := m.displayedModels()
	if m.midCursor >= len(visible) {
		m.midCursor = max(0, len(visible)-1)
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

	case profilesLoadedMsg:
		if msg.err != nil {
			m.message = "failed to load profiles: " + msg.err.Error()
		} else {
			m.profiles = msg.profiles
			for i, p := range m.profiles {
				if p == m.activeProfile {
					m.profileCursor = i
					break
				}
			}
		}
		return m, nil

	case profileActivatedMsg:
		if msg.err != nil {
			m.message = "error: " + msg.err.Error()
		} else {
			m.cfg = msg.cfg
			m.activeProfile = msg.profile
			m.message = "profile activated: " + msg.profile
			m.showProfiles = false
			m.creatingProfile = false
			m.profileInput.SetValue("")
		}
		return m, nil

	case tea.KeyMsg:
		if m.creatingProfile {
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.profileInput.Value())
				if name == "" {
					m.message = "profile name required"
					return m, nil
				}
				// createProfileCmd is missing
				newProfile, err := m.cfg.CloneToNewProfile(name)
				if err != nil {
					m.message = "error: " + err.Error()
					return m, nil
				}
				// Refresh profiles to show the new one
				return m, tea.Batch(
					func() tea.Msg {
						return profileActivatedMsg{cfg: m.cfg, profile: newProfile}
					},
					fetchProfilesCmd,
				)

			case "esc":
				m.creatingProfile = false
				m.profileInput.SetValue("")
				m.message = "cancelled"
				return m, nil
			}

			var cmd tea.Cmd
			m.profileInput, cmd = m.profileInput.Update(msg)
			return m, cmd
		}

		if m.showProfiles {
			switch msg.String() {
			case "esc", "p":
				m.showProfiles = false
				return m, nil
			case "n":
				m.creatingProfile = true
				m.profileInput.SetValue("")
				m.profileInput.Focus()
				m.message = "enter new profile name"
				return m, nil
			case "j", "down":
				if m.profileCursor < len(m.profiles)-1 {
					m.profileCursor++
				}
			case "k", "up":
				if m.profileCursor > 0 {
					m.profileCursor--
				}
			case "enter", "a":
				if m.profileCursor >= 0 && m.profileCursor < len(m.profiles) {
					p := m.profiles[m.profileCursor]
					return m, activateProfileCmd(p)
				}
			}
			return m, nil
		}

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
				if m.midCursor < len(m.displayedModels())-1 {
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
			if m.focus == paneMiddle {
				entry := m.currentEntry()
				visible := m.displayedModels()
				if entry != nil && len(visible) > 0 && m.midCursor >= 0 && m.midCursor < len(visible) {
					entry.Model = visible[m.midCursor]
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

		case "p":
			m.showProfiles = true
			return m, fetchProfilesCmd

		case "a":
			if len(m.profiles) == 0 {
				m.message = "no profiles available"
				return m, nil
			}
			if m.profileCursor >= 0 && m.profileCursor < len(m.profiles) {
				p := m.profiles[m.profileCursor]
				return m, activateProfileCmd(p)
			}
			m.message = "invalid profile selection"
			return m, nil
		}
	}

	return m, nil
}
