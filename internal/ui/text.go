package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func padRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 3 {
		return trimToWidth(s, width)
	}
	ellipsis := "..."
	return trimToWidth(s, width-lipgloss.Width(ellipsis)) + ellipsis
}

func trimToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	current := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if current+rw > width {
			break
		}
		b.WriteRune(r)
		current += rw
	}
	return b.String()
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	words := strings.Fields(text)
	lines := make([]string, 0, len(words))
	line := words[0]
	for _, word := range words[1:] {
		if lipgloss.Width(line)+1+lipgloss.Width(word) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)

	return strings.Join(lines, "\n")
}

func wrapLines(lines []string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			wrapped = append(wrapped, "")
			continue
		}
		w := wrapText(line, width)
		wrapped = append(wrapped, strings.Split(w, "\n")...)
	}
	return wrapped
}

func renderLineWithWidth(left, right string, width int) string {
	if width <= 0 {
		return ""
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return padRight(left, width)
	}

	rightWidth := lipgloss.Width(right)
	if rightWidth >= width {
		return padRight(truncate(right, width), width)
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		maxLeft := width - lipgloss.Width(right) - 1
		if maxLeft < 0 {
			maxLeft = 0
		}
		left = truncate(left, maxLeft)
		gap = width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
	}

	return padRight(left+strings.Repeat(" ", gap)+right, width)
}
