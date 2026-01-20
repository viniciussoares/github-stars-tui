# gh-stars

A terminal UI to browse and search your GitHub stars, inspired by gh-dash.

## Requirements
- GitHub CLI (`gh`) authenticated via `gh auth login`
- Go 1.22+

## Run
```
go run ./cmd/gh-stars
```

## Build
```
go build -o gh-stars ./cmd/gh-stars
./gh-stars
```

## Keys
- `j`/`k` or arrows: move
- `/`: focus search
- `esc`: exit search
- `enter`: open selected repo in browser
- `y`: copy selected repo URL
- `q`: quit

## Cache
By default, stars are cached to `~/.config/gh-stars/cache.json` and refreshed in the background every 48h.
Use `-refresh` to force a foreground refresh on startup.
Use `-sync-interval` to change the background refresh cadence or `-sync-interval=0` to disable it.
Use `-cache ''` to disable caching.
