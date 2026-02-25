package tui

import (
	"fmt"
	"strings"

	"omoc/internal/config"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED"))

	normalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	selectedItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666"))

	modelHighlight = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E"))

	variantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	statusOk = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E"))

	statusErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666"))

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	infoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCC"))

	fallbackStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Italic(true)
)

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	leftWidth := 40
	infoWidth := 38
	midWidth := m.width - leftWidth - infoWidth - 8
	if midWidth < 30 {
		midWidth = 30
	}
	contentHeight := m.height - 5

	header := titleStyle.Render("⚙ omoc — model configurator") + "\n"
	left := m.viewLeft(leftWidth, contentHeight)
	mid := m.viewMiddle(midWidth, contentHeight)
	right := m.viewInfo(infoWidth, contentHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", mid, " ", right)
	footer := m.viewFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) viewLeft(width, height int) string {
	var lines []string

	for i, item := range m.leftItems {
		if item.isHeader {
			centered := lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(item.name))
			lines = append(lines, centered)
			continue
		}

		var entry *config.AgentEntry
		if item.isAgent {
			entry = m.cfg.Agents[item.name]
		} else {
			entry = m.cfg.Categories[item.name]
		}

		cursor := "  "
		style := dimStyle
		if i == m.leftCursor {
			cursor = "▸ "
			style = selectedItem
		}

		label := style.Render(item.name)

		detail := ""
		if entry != nil && entry.Model != "" {
			short := entry.Model
			parts := strings.SplitN(short, "/", 2)
			if len(parts) == 2 {
				short = parts[1]
			}
			variantLen := 0
			if entry.Variant != "" {
				variantLen = len(entry.Variant) + 1
			}
			maxModel := width - len(item.name) - 6 - variantLen
			if maxModel < 6 {
				maxModel = 6
			}
			detail = " " + modelHighlight.Render(truncate(short, maxModel))
			if entry.Variant != "" {
				detail += " " + variantStyle.Render(entry.Variant)
			}
		}

		lines = append(lines, cursor+label+detail)
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	content := strings.Join(lines, "\n")

	border := normalBorder
	if m.focus == paneLeft {
		border = focusedBorder
	}
	return border.Width(width).Height(height).Render(content)
}

func (m Model) viewMiddle(width, height int) string {
	filterLine := "  " + m.filter.View()
	visibleHeight := height - 2

	var lines []string

	currentModel := ""
	if entry := m.currentEntry(); entry != nil {
		currentModel = entry.Model
	}

	sorted := m.sortModelsCurrentFirst(m.filteredModels)

	if len(sorted) == 0 {
		lines = append(lines, dimStyle.Render("  no models match filter"))
	} else {
		scrollTop := m.midCursor - visibleHeight + 1
		if scrollTop < 0 {
			scrollTop = 0
		}
		if scrollTop > m.midCursor {
			scrollTop = m.midCursor
		}

		end := scrollTop + visibleHeight
		if end > len(sorted) {
			end = len(sorted)
		}

		for i := scrollTop; i < end; i++ {
			model := sorted[i]
			cursor := "  "
			style := dimStyle

			if i == m.midCursor && m.focus == paneMiddle {
				cursor = "▸ "
				style = selectedItem
			}

			marker := "  "
			if model == currentModel {
				marker = "★ "
				if i != m.midCursor || m.focus != paneMiddle {
					style = modelHighlight
				}
			}

			lines = append(lines, cursor+style.Render(marker+truncate(model, width-8)))
		}
	}

	for len(lines) < visibleHeight {
		lines = append(lines, "")
	}

	content := filterLine + "\n" + strings.Join(lines, "\n")

	border := normalBorder
	if m.focus == paneMiddle {
		border = focusedBorder
	}
	return border.Width(width).Height(height).Render(content)
}

func (m Model) viewInfo(width, height int) string {
	info := m.currentInfo()
	item := m.currentItem()

	var lines []string

	if info == nil || item.isHeader {
		lines = append(lines, dimStyle.Render("select an item"))
	} else {
		lines = append(lines, infoLabelStyle.Render(item.name))
		lines = append(lines, "")

		lines = append(lines, infoLabelStyle.Render("Role"))
		lines = append(lines, infoValueStyle.Render(info.Role))
		lines = append(lines, "")

		lines = append(lines, infoLabelStyle.Render("Description"))
		for _, l := range wrapText(info.Description, width-2) {
			lines = append(lines, infoValueStyle.Render(l))
		}
		lines = append(lines, "")

		lines = append(lines, infoLabelStyle.Render("Fallback Chain"))
		for i, fb := range info.Fallback {
			prefix := "  "
			if i == 0 {
				prefix = "→ "
			} else {
				prefix = "  → "
			}
			lines = append(lines, fallbackStyle.Render(prefix+fb))
		}
		lines = append(lines, "")

		lines = append(lines, infoLabelStyle.Render("Notes"))
		for _, l := range wrapText(info.Notes, width-2) {
			lines = append(lines, noteStyle.Render(l))
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	content := strings.Join(lines, "\n")
	return normalBorder.Width(width).Height(height).Render(content)
}

func (m Model) viewFooter() string {
	status := ""
	if m.message != "" {
		if m.saved {
			status = statusOk.Render(" " + m.message)
		} else {
			status = statusErr.Render(" " + m.message)
		}
	}

	help := helpStyle.Render("tab/h/l:switch  j/k:move  enter:select  v:variant  d:clear  /:filter  s:save  q:quit")

	return fmt.Sprintf("\n%s\n%s", status, help)
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func wrapText(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if len(current)+1+len(w) > width {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}
