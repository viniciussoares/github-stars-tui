package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	HeaderBar                lipgloss.Style
	HeaderTitle              lipgloss.Style
	HeaderMeta               lipgloss.Style
	SearchPrompt             lipgloss.Style
	SearchText               lipgloss.Style
	SearchInactive           lipgloss.Style
	SearchBox                lipgloss.Style
	Panel                    lipgloss.Style
	PanelTitle               lipgloss.Style
	ListRow                  lipgloss.Style
	ListRowSecondary         lipgloss.Style
	ListRowSelected          lipgloss.Style
	ListRowSelectedSecondary lipgloss.Style
	PreviewTitle             lipgloss.Style
	Muted                    lipgloss.Style
	Footer                   lipgloss.Style
	FooterKey                lipgloss.Style
	FooterError              lipgloss.Style
	Divider                  lipgloss.Style
}

func DefaultStyles() Styles {
	accent := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#BD93F9"}    // purple
	accentAlt := lipgloss.AdaptiveColor{Light: "#DB2777", Dark: "#FF79C6"} // pink
	muted := lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6272A4"}     // gray
	info := lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#8BE9FD"}      // cyan
	success := lipgloss.AdaptiveColor{Light: "#059669", Dark: "#50FA7B"}   // green
	border := lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#44475A"}
	text := lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F8F8F2"}

	return Styles{
		HeaderBar:                lipgloss.NewStyle(),
		HeaderTitle:              lipgloss.NewStyle().Bold(true).Foreground(accent),
		HeaderMeta:               lipgloss.NewStyle().Foreground(muted),
		SearchPrompt:             lipgloss.NewStyle().Foreground(muted),
		SearchText:               lipgloss.NewStyle().Foreground(text),
		SearchInactive:           lipgloss.NewStyle().Foreground(muted),
		SearchBox:                lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(0, searchBoxPadding),
		Panel:                    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(panelPaddingY, panelPaddingX),
		PanelTitle:               lipgloss.NewStyle().Bold(true).Foreground(accentAlt),
		ListRow:                  lipgloss.NewStyle(),
		ListRowSecondary:         lipgloss.NewStyle().Foreground(muted),
		ListRowSelected:          lipgloss.NewStyle().Foreground(success).Bold(true),
		ListRowSelectedSecondary: lipgloss.NewStyle().Foreground(info),
		PreviewTitle:             lipgloss.NewStyle().Bold(true).Foreground(accent),
		Muted:                    lipgloss.NewStyle().Foreground(muted),
		Footer:                   lipgloss.NewStyle().Foreground(muted),
		FooterKey:                lipgloss.NewStyle().Foreground(accent).Bold(true),
		FooterError:              lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FF5555"}),
		Divider:                  lipgloss.NewStyle().Foreground(border),
	}
}
