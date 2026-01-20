package ui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/viniciussoares/github-stars-tui/internal/data"
)

type Model struct {
	client *gh.GraphQLClient

	styles      Styles
	spinner     spinner.Model
	searchInput textinput.Model

	width  int
	height int

	listWidth    int
	previewWidth int

	repos      []data.Repo
	filtered   []int
	cursor     int
	offset     int
	totalCount int
	pageSize   int
	nextCursor *string
	loading    bool
	err        error

	searchFocused bool
	sortMode      string

	status        string
	statusIsError bool

	cachePath  string
	cacheIndex map[string]struct{}
	cacheDirty bool

	deferRefresh bool
	pendingNew   []data.Repo
}

type starsPageMsg struct {
	page data.StarsPage
}

type statusMsg struct {
	text    string
	isError bool
}

type errorMsg struct {
	err error
}

func NewModel(client *gh.GraphQLClient, pageSize int, cachePath string, cachedRepos []data.Repo, fetchOnStart bool, backgroundSync bool) Model {
	styles := DefaultStyles()

	sp := spinner.New(spinner.WithSpinner(spinner.Spinner{
		Frames: []string{"-", "\\", "|", "/"},
		FPS:    time.Second / 8,
	}))
	sp.Style = styles.HeaderMeta

	ti := textinput.New()
	ti.Prompt = "Search: "
	ti.Placeholder = "name, description, repo/name"
	ti.CharLimit = 200
	ti.Blur()
	ti.PromptStyle = styles.SearchPrompt
	ti.TextStyle = styles.SearchInactive
	ti.PlaceholderStyle = styles.SearchInactive

	cacheIndex := make(map[string]struct{}, len(cachedRepos))
	for _, repo := range cachedRepos {
		cacheIndex[repo.NameWithOwner] = struct{}{}
	}

	status := "loading"
	deferRefresh := false
	if len(cachedRepos) > 0 {
		if fetchOnStart {
			status = "refreshing"
			deferRefresh = true
		} else if backgroundSync {
			status = "syncing (bg)"
		} else {
			status = "cached"
		}
	}

	model := Model{
		client:        client,
		styles:        styles,
		spinner:       sp,
		searchInput:   ti,
		pageSize:      pageSize,
		loading:       fetchOnStart,
		status:        status,
		statusIsError: false,
		cachePath:     cachePath,
		cacheIndex:    cacheIndex,
		repos:         cachedRepos,
		deferRefresh:  deferRefresh,
		sortMode:      "default",
	}
	model.applyFilter()
	return model
}

func (m Model) Init() tea.Cmd {
	if !m.loading {
		return nil
	}
	return tea.Batch(m.spinner.Tick, fetchStarsPageCmd(m.client, m.pageSize, m.nextCursor))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		m.applyFilter()
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case starsPageMsg:
		foundCached := false
		newRepos := make([]data.Repo, 0, len(msg.page.Repos))
		for _, repo := range msg.page.Repos {
			if _, exists := m.cacheIndex[repo.NameWithOwner]; exists {
				foundCached = true
				continue
			}
			newRepos = append(newRepos, repo)
			m.cacheIndex[repo.NameWithOwner] = struct{}{}
		}
		if len(newRepos) > 0 {
			if m.deferRefresh {
				m.pendingNew = append(m.pendingNew, newRepos...)
			} else {
				m.repos = append(newRepos, m.repos...)
			}
			m.cacheDirty = true
		}
		if msg.page.TotalCount > 0 {
			m.totalCount = msg.page.TotalCount
		}
		if msg.page.EndCursor != "" {
			next := msg.page.EndCursor
			m.nextCursor = &next
		}
		if foundCached && (len(m.repos) > 0 || len(m.pendingNew) > 0) {
			m.loading = false
		} else {
			m.loading = msg.page.HasNext
		}
		if m.loading {
			if len(m.repos) > 0 {
				m.status = "refreshing"
			} else {
				m.status = "loading"
			}
		} else {
			m.status = "ready"
		}
		m.statusIsError = false
		if !m.deferRefresh || !m.loading {
			if m.deferRefresh && len(m.pendingNew) > 0 {
				m.repos = append(m.pendingNew, m.repos...)
				m.pendingNew = nil
				m.deferRefresh = false
			}
			m.applyFilter()
		}

		if m.cacheDirty && !m.loading {
			if err := data.SaveCache(m.cachePath, m.repos); err != nil {
				m.status = fmt.Sprintf("cache save failed: %v", err)
				m.statusIsError = true
			} else {
				m.cacheDirty = false
			}
		}

		if m.loading {
			return m, fetchStarsPageCmd(m.client, m.pageSize, m.nextCursor)
		}
		return m, nil
	case statusMsg:
		m.status = msg.text
		m.statusIsError = msg.isError
		return m, nil
	case errorMsg:
		m.loading = false
		m.err = msg.err
		m.status = msg.err.Error()
		m.statusIsError = true
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		if m.searchFocused {
			switch key {
			case "esc", "enter":
				m.blurSearch()
				return m, nil
			}
			break
		}

		switch key {
		case "/":
			m.focusSearch()
			return m, nil
		case "up", "k":
			m.moveCursor(-1)
			return m, nil
		case "down", "j":
			m.moveCursor(1)
			return m, nil
		case "pgup":
			m.moveCursor(-m.listBodyRows())
			return m, nil
		case "pgdown":
			m.moveCursor(m.listBodyRows())
			return m, nil
		case "g":
			m.moveToTop()
			return m, nil
		case "G":
			m.moveToBottom()
			return m, nil
		case "enter":
			repo := m.selectedRepo()
			if repo == nil {
				return m, nil
			}
			return m, openRepoCmd(repo.URL)
		case "y":
			repo := m.selectedRepo()
			if repo == nil {
				return m, nil
			}
			return m, copyURLCmd(repo.URL)
		case "r", "R":
			if m.loading {
				return m, nil
			}
			m.repos = nil
			m.filtered = nil
			m.cacheIndex = make(map[string]struct{})
			m.cursor = 0
			m.offset = 0
			m.totalCount = 0
			m.nextCursor = nil
			m.loading = true
			m.status = "refreshing"
			m.applyFilter()
			return m, tea.Batch(m.spinner.Tick, fetchStarsPageCmd(m.client, m.pageSize, m.nextCursor))
		case "s":
			m.cycleSortMode()
			m.applyFilter()
			return m, nil
		}
	}

	if m.searchFocused {
		prev := m.searchInput.Value()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		if m.searchInput.Value() != prev {
			m.applyFilter()
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height

	promptWidth := lipgloss.Width(m.searchInput.Prompt)
	searchInnerWidth := width - (2 * searchBoxBorder) - (2 * searchBoxPadding)
	m.searchInput.Width = max(0, searchInnerWidth-promptWidth)

	if width < 120 {
		m.listWidth = width
		m.previewWidth = 0
		return
	}

	panelsTotalWidth := width
	m.listWidth = int(float64(panelsTotalWidth) * 0.55)
	if m.listWidth < 30 {
		m.listWidth = 30
	}
	m.previewWidth = max(0, panelsTotalWidth-m.listWidth)
}

func (m *Model) focusSearch() {
	m.searchFocused = true
	m.searchInput.Focus()
	m.searchInput.TextStyle = m.styles.SearchText
}

func (m *Model) blurSearch() {
	m.searchFocused = false
	m.searchInput.Blur()
	m.searchInput.TextStyle = m.styles.SearchInactive
}

func (m *Model) listHeight() int {
	return max(1, m.height-m.headerHeight()-footerHeight)
}

func (m *Model) headerHeight() int {
	return lipgloss.Height(m.renderHeader())
}

func (m *Model) panelContentHeight(panelHeight int) int {
	return max(1, panelHeight-(2*panelBorderWidth)-(2*panelPaddingY))
}

func (m *Model) panelContentWidth(panelWidth int) int {
	return max(1, panelWidth-(2*panelBorderWidth)-(2*panelPaddingX))
}

func (m *Model) listBodyRows() int {
	bodyHeight := m.panelContentHeight(m.listHeight()) - 1
	if bodyHeight <= 0 {
		return 1
	}
	return max(1, bodyHeight/listRowHeight)
}

func (m *Model) applyFilter() {
	query := strings.TrimSpace(strings.ToLower(m.searchInput.Value()))
	m.filtered = m.filtered[:0]

	for i, repo := range m.repos {
		if matchesQuery(repo, query) {
			m.filtered = append(m.filtered, i)
		}
	}

	if len(m.filtered) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}

	// Sort filtered results
	m.sortFiltered()

	m.cursor = clamp(m.cursor, 0, len(m.filtered)-1)
	m.ensureCursorVisible()
}

func (m *Model) cycleSortMode() {
	switch m.sortMode {
	case "default":
		m.sortMode = "stars"
	case "stars":
		m.sortMode = "name"
	case "name":
		m.sortMode = "updated"
	case "updated":
		m.sortMode = "default"
	default:
		m.sortMode = "default"
	}
}

func (m *Model) sortFiltered() {
	if len(m.filtered) == 0 {
		return
	}

	sort.Slice(m.filtered, func(i, j int) bool {
		a := m.repos[m.filtered[i]]
		b := m.repos[m.filtered[j]]
		switch m.sortMode {
		case "default":
			return a.StarredAt.After(b.StarredAt)
		case "stars":
			return a.Stars > b.Stars
		case "name":
			return strings.ToLower(a.NameWithOwner) < strings.ToLower(b.NameWithOwner)
		case "updated":
			return a.UpdatedAt.After(b.UpdatedAt)
		}
		return false
	})
}

func matchesQuery(repo data.Repo, query string) bool {
	if query == "" {
		return true
	}

	haystack := strings.ToLower(strings.Join([]string{
		repo.NameWithOwner,
		repo.Name,
		repo.Description,
		repo.PrimaryLanguage,
		strings.Join(repo.Topics, " "),
	}, " "))

	for _, term := range strings.Fields(query) {
		if !strings.Contains(haystack, term) {
			return false
		}
	}

	return true
}

func (m *Model) moveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = clamp(m.cursor+delta, 0, len(m.filtered)-1)
	m.ensureCursorVisible()
}

func (m *Model) ensureCursorVisible() {
	if len(m.filtered) == 0 {
		return
	}
	viewHeight := m.listBodyRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+viewHeight {
		m.offset = m.cursor - viewHeight + 1
	}
	m.offset = max(0, m.offset)
}

func (m *Model) moveToTop() {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = 0
	m.offset = 0
}

func (m *Model) moveToBottom() {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = len(m.filtered) - 1
	m.offset = max(0, m.cursor-m.listBodyRows()+1)
}

func (m *Model) selectedRepo() *data.Repo {
	if len(m.filtered) == 0 {
		return nil
	}
	idx := m.filtered[m.cursor]
	if idx < 0 || idx >= len(m.repos) {
		return nil
	}
	return &m.repos[idx]
}

func fetchStarsPageCmd(client *gh.GraphQLClient, pageSize int, after *string) tea.Cmd {
	return func() tea.Msg {
		page, err := data.FetchStarsPage(context.Background(), client, pageSize, after)
		if err != nil {
			return errorMsg{err: err}
		}
		return starsPageMsg{page: page}
	}
}

func openRepoCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return statusMsg{text: "no URL to open", isError: true}
		}
		b := browser.New("", io.Discard, io.Discard)
		if err := b.Browse(url); err != nil {
			return statusMsg{text: fmt.Sprintf("open failed: %v", err), isError: true}
		}
		return statusMsg{text: "opened in browser", isError: false}
	}
}

func copyURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return statusMsg{text: "no URL to copy", isError: true}
		}
		if err := clipboard.WriteAll(url); err != nil {
			return statusMsg{text: fmt.Sprintf("copy failed: %v", err), isError: true}
		}
		return statusMsg{text: "copied URL", isError: false}
	}
}
