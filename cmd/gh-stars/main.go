package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/cli/go-gh/v2/pkg/api"

	"github.com/vini/gh-stars/internal/data"
	"github.com/vini/gh-stars/internal/ui"
)

func main() {
	pageSize := flag.Int("page-size", 100, "Stars to fetch per request (max 100)")
	defaultPath := defaultCachePath()
	cachePath := flag.String("cache", defaultPath, "Cache file path")
	refresh := flag.Bool("refresh", false, "Force refresh on startup")
	syncInterval := flag.Duration("sync-interval", 48*time.Hour, "Background refresh interval")
	flag.Parse()

	if *pageSize <= 0 || *pageSize > 100 {
		fmt.Fprintln(os.Stderr, "page-size must be between 1 and 100")
		os.Exit(2)
	}

	client, err := gh.DefaultGraphQLClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create GitHub client:", err)
		fmt.Fprintln(os.Stderr, "make sure `gh auth login` has been run")
		os.Exit(1)
	}

	cache, err := data.LoadCache(*cachePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: cache load failed:", err)
	}
	if *cachePath == defaultPath && cache.SavedAt.IsZero() {
		legacyPath := ".cache/gh-stars.json"
		legacy, err := data.LoadCache(legacyPath)
		if err == nil && len(legacy.Repos) > 0 {
			if err := data.SaveCache(*cachePath, legacy.Repos); err != nil {
				fmt.Fprintln(os.Stderr, "warning: cache migration failed:", err)
			} else {
				cache = legacy
			}
		}
	}

	backgroundSync := false
	if *cachePath != "" && len(cache.Repos) > 0 && data.IsStale(cache, *syncInterval) && !*refresh {
		backgroundSync = true
		go func() {
			if err := data.RefreshCache(client, *pageSize, *cachePath, cache); err != nil {
				fmt.Fprintln(os.Stderr, "background refresh failed:", err)
			}
		}()
	}

	fetchOnStart := *refresh || len(cache.Repos) == 0
	model := ui.NewModel(client, *pageSize, *cachePath, cache.Repos, fetchOnStart, backgroundSync)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to run UI:", err)
		os.Exit(1)
	}
}

func defaultCachePath() string {
	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, "gh-stars", "cache.json")
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "gh-stars", "cache.json")
	}

	return ".cache/gh-stars.json"
}
