package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	if m.err != nil {
		return m.renderError()
	}
	if m.height < m.headerHeight()+footerHeight+3 || m.width < 20 {
		return "window too small"
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return strings.Join([]string{header, body, footer}, "\n")
}

func (m Model) renderError() string {
	msg := fmt.Sprintf("error: %v", m.err)
	return msg
}

func (m Model) renderHeader() string {
	search := m.searchInput.View()
	return m.styles.SearchBox.Width(m.width - 2*searchBoxPadding).Render(search)
}

func (m Model) renderBody() string {
	height := m.listHeight()
	listContentHeight := m.panelContentHeight(height)
	listContentWidth := m.panelContentWidth(m.listWidth)
	listContent := m.renderList(listContentHeight, listContentWidth)
	listPanel := m.panelStyle(m.listWidth, height).Render(listContent)

	if m.previewWidth <= 0 {
		return listPanel
	}

	previewContentHeight := m.panelContentHeight(height)
	previewContentWidth := m.panelContentWidth(m.previewWidth)
	previewContent := m.renderPreview(previewContentHeight, previewContentWidth)
	previewPanel := m.panelStyle(m.previewWidth, height).Render(previewContent)

	return joinColumns(listPanel, previewPanel, "", height, m.listWidth, m.previewWidth)
}

func (m Model) panelStyle(width, height int) lipgloss.Style {
	innerWidth := max(1, width-(2*panelPaddingX))
	innerHeight := max(1, height-(2*panelBorderWidth)-(2*panelPaddingY))
	return m.styles.Panel.Width(innerWidth).Height(innerHeight)
}

func (m Model) renderFooter() string {
	key := m.styles.FooterKey.Render
	txt := m.styles.Footer.Render
	sep := txt(" · ")
	help := key("q") + txt(" quit") + sep + key("/") + txt(" search") + sep + key("↵") + txt(" open") + sep + key("y") + txt(" copy") + sep + key("r") + txt(" refresh") + sep + key("s") + txt(" sort") + sep + key("j/k") + txt(" move") + sep + key("g/G") + txt(" top/bottom")

	status := strings.TrimSpace(m.status)
	if status == "" {
		status = "ready"
	}
	if m.loading {
		status = m.spinner.View() + " " + status
	}

	countText := m.countText()
	if countText != "" {
		status = status + "  " + countText
	}

	if m.sortMode != "default" {
		status = status + "  [" + m.sortMode + "]"
	}

	left := help
	rightStyle := m.styles.Footer
	if m.statusIsError {
		rightStyle = m.styles.FooterError
	}
	right := rightStyle.Render(status)

	return m.renderLine(left, right)
}

func (m Model) renderList(height, width int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := []string{}
	bodyHeight := max(0, height)
	if bodyHeight == 0 {
		return strings.Join(lines, "\n")
	}

	if len(m.filtered) == 0 {
		lines = append(lines, padRight(m.styles.Muted.Render("no matches"), width))
		for len(lines) < height {
			lines = append(lines, strings.Repeat(" ", width))
		}
		return strings.Join(lines, "\n")
	}

	start := m.offset
	rowsVisible := max(1, (bodyHeight+1)/listRowHeight)
	end := min(start+rowsVisible, len(m.filtered))
	for i := start; i < end; i++ {
		idx := m.filtered[i]
		if idx < 0 || idx >= len(m.repos) {
			continue
		}
		repo := m.repos[idx]
		selected := i == m.cursor

		metaParts := []string{}
		if repo.IsFork {
			metaParts = append(metaParts, "fork 🍴")
		}
		if repo.PrimaryLanguage != "" {
			metaParts = append(metaParts, repo.PrimaryLanguage+" 🧪")
		}
		metaParts = append(metaParts, fmt.Sprintf("%6d ⭐", repo.Stars))
		name := repo.NameWithOwner
		meta := strings.Join(metaParts, "  ")
		line1 := renderLineWithWidth(truncate(name, width), meta, width)

		desc := repo.Description
		if desc == "" {
			desc = "-"
		}
		desc = truncate(desc, width)
		line2 := padRight(desc, width)

		if selected {
			line1 = m.styles.ListRowSelected.Render(line1)
			line2 = m.styles.ListRowSelectedSecondary.Render(line2)
		} else {
			line1 = m.styles.ListRow.Render(line1)
			line2 = m.styles.ListRowSecondary.Render(line2)
		}

		lines = append(lines, line1, line2)
		if i < end-1 {
			divider := strings.Repeat("─", width)
			lines = append(lines, m.styles.Divider.Render(divider))
		}
	}

	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPreview(height, width int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := []string{}
	bodyHeight := max(0, height)
	if bodyHeight == 0 {
		return strings.Join(lines, "\n")
	}

	if len(m.filtered) == 0 {
		lines = append(lines, padRight(m.styles.Muted.Render("no selection"), width))
		for len(lines) < height {
			lines = append(lines, strings.Repeat(" ", width))
		}
		return strings.Join(lines, "\n")
	}

	repo := m.selectedRepo()
	if repo == nil {
		lines = append(lines, padRight(m.styles.Muted.Render("no selection"), width))
		for len(lines) < height {
			lines = append(lines, strings.Repeat(" ", width))
		}
		return strings.Join(lines, "\n")
	}

	// Repository name
	nameLines := wrapLines([]string{repo.NameWithOwner}, width)
	for _, line := range nameLines {
		lines = append(lines, m.styles.PreviewTitle.Render(line))
	}
	lines = append(lines, "")

	// Meta info (language, stars, date, fork)
	metaParts := make([]string, 0, 4)
	if repo.PrimaryLanguage != "" {
		metaParts = append(metaParts, "🧪 "+repo.PrimaryLanguage)
	}
	metaParts = append(metaParts, fmt.Sprintf("⭐ %d", repo.Stars))
	if !repo.UpdatedAt.IsZero() {
		metaParts = append(metaParts, "🕒 "+repo.UpdatedAt.Format("2006-01-02"))
	}
	if repo.IsFork {
		metaParts = append(metaParts, "🍴 fork")
	}
	if len(metaParts) > 0 {
		metaLines := wrapLines([]string{strings.Join(metaParts, "  ")}, width)
		for _, line := range metaLines {
			lines = append(lines, m.styles.Muted.Render(line))
		}
	}
	lines = append(lines, "")

	// Description
	desc := repo.Description
	if strings.TrimSpace(desc) != "" {
		descLines := wrapLines([]string{desc}, width)
		lines = append(lines, descLines...)
		lines = append(lines, "")
	}

	// Topics
	if len(repo.Topics) > 0 {
		topicLines := wrapLines([]string{strings.Join(repo.Topics, ", ")}, width)
		for _, line := range topicLines {
			lines = append(lines, m.styles.Muted.Render(line))
		}
		lines = append(lines, "")
	}

	// URL
	lines = append(lines, m.styles.Muted.Render(repo.URL))

	for i := range lines {
		lines[i] = padRight(lines[i], width)
	}

	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderLine(left, right string) string {
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	return padRight(line, m.width)
}

func (m Model) countText() string {
	if m.totalCount == 0 {
		if len(m.repos) == 0 {
			return ""
		}
		return fmt.Sprintf("%d loaded", len(m.repos))
	}
	if strings.TrimSpace(m.searchInput.Value()) != "" {
		return fmt.Sprintf("%d/%d match", len(m.filtered), m.totalCount)
	}
	return fmt.Sprintf("%d total", m.totalCount)
}

func joinColumns(left, right, sep string, height, leftWidth, rightWidth int) string {
	leftLines := splitLines(left, height, leftWidth)
	rightLines := splitLines(right, height, rightWidth)

	lines := make([]string, 0, height)
	for i := 0; i < height; i++ {
		lines = append(lines, leftLines[i]+sep+rightLines[i])
	}
	return strings.Join(lines, "\n")
}

func splitLines(text string, height, width int) []string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = padRight(ansi.Truncate(lines[i], width, ""), width)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return lines
}
