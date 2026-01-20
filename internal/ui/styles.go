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
}

func DefaultStyles() Styles {
	accent := lipgloss.Color("#BD93F9")
	accentAlt := lipgloss.Color("#FF79C6")
	muted := lipgloss.Color("#6272A4")
	info := lipgloss.Color("#8BE9FD")
	success := lipgloss.Color("#50FA7B")
	border := lipgloss.AdaptiveColor{Light: "250", Dark: "238"}

	return Styles{
		HeaderBar:                lipgloss.NewStyle(),
		HeaderTitle:              lipgloss.NewStyle().Bold(true).Foreground(accent),
		HeaderMeta:               lipgloss.NewStyle().Foreground(muted),
		SearchPrompt:             lipgloss.NewStyle().Foreground(muted),
		SearchText:               lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}),
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
		FooterError:              lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}),
	}
}
