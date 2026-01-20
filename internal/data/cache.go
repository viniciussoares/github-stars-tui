package data

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	gh "github.com/cli/go-gh/v2/pkg/api"
)

type Cache struct {
	SavedAt time.Time `json:"saved_at"`
	Repos   []Repo    `json:"repos"`
}

func LoadCache(path string) (Cache, error) {
	if path == "" {
		return Cache{}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cache{}, nil
		}
		return Cache{}, err
	}

	var file Cache
	if err := json.Unmarshal(content, &file); err != nil {
		return Cache{}, err
	}
	return file, nil
}

func SaveCache(path string, repos []Repo) error {
	if path == "" {
		return nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	payload := Cache{
		SavedAt: time.Now().UTC(),
		Repos:   repos,
	}

	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o644)
}

func IsStale(cache Cache, interval time.Duration) bool {
	if interval <= 0 {
		return false
	}
	if cache.SavedAt.IsZero() {
		return true
	}
	return time.Since(cache.SavedAt) >= interval
}

func RefreshCache(client *gh.GraphQLClient, pageSize int, path string, cache Cache) error {
	if path == "" || client == nil {
		return nil
	}

	cacheIndex := make(map[string]struct{}, len(cache.Repos))
	for _, repo := range cache.Repos {
		cacheIndex[repo.NameWithOwner] = struct{}{}
	}

	repos := cache.Repos
	var cursor *string

	for {
		page, err := FetchStarsPage(context.Background(), client, pageSize, cursor)
		if err != nil {
			return err
		}

		newRepos := make([]Repo, 0, len(page.Repos))
		foundCached := false
		for _, repo := range page.Repos {
			if _, exists := cacheIndex[repo.NameWithOwner]; exists {
				foundCached = true
				continue
			}
			newRepos = append(newRepos, repo)
			cacheIndex[repo.NameWithOwner] = struct{}{}
		}

		if len(newRepos) > 0 {
			repos = append(newRepos, repos...)
		}

		if foundCached || !page.HasNext {
			break
		}

		if page.EndCursor != "" {
			next := page.EndCursor
			cursor = &next
		} else {
			break
		}
	}

	return SaveCache(path, repos)
}
