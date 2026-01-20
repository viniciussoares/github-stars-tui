package data

import (
	"context"
	"errors"
	"strings"
	"time"

	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type Repo struct {
	Name            string
	NameWithOwner   string
	Description     string
	URL             string
	Stars           int
	PrimaryLanguage string
	UpdatedAt       time.Time
	IsFork          bool
	Topics          []string
}

type StarsPage struct {
	Repos      []Repo
	TotalCount int
	EndCursor  string
	HasNext    bool
}

type topicNode struct {
	Topic struct {
		Name string
	}
}

func FetchStarsPage(ctx context.Context, client *gh.GraphQLClient, pageSize int, after *string) (StarsPage, error) {
	if client == nil {
		return StarsPage{}, errors.New("nil GraphQL client")
	}

	select {
	case <-ctx.Done():
		return StarsPage{}, ctx.Err()
	default:
	}

	var query struct {
		Viewer struct {
			StarredRepositories struct {
				Nodes []struct {
					Name            string
					NameWithOwner   string `graphql:"nameWithOwner"`
					Description     string
					URL             string
					Stars           int `graphql:"stargazerCount"`
					UpdatedAt       time.Time
					IsFork          bool `graphql:"isFork"`
					PrimaryLanguage *struct {
						Name string
					}
					RepositoryTopics struct {
						Nodes []topicNode
					} `graphql:"repositoryTopics(first: 5)"`
				}
				TotalCount int
				PageInfo   struct {
					HasNextPage bool
					EndCursor   graphql.String
				}
			} `graphql:"starredRepositories(first: $first, after: $after, orderBy: {field: STARRED_AT, direction: DESC})"`
		}
	}

	var afterVar *graphql.String
	if after != nil {
		v := graphql.String(*after)
		afterVar = &v
	}

	variables := map[string]any{
		"first": graphql.Int(pageSize),
		"after": afterVar,
	}

	if err := client.Query("ViewerStars", &query, variables); err != nil {
		return StarsPage{}, err
	}

	repos := make([]Repo, 0, len(query.Viewer.StarredRepositories.Nodes))
	for _, node := range query.Viewer.StarredRepositories.Nodes {
		primaryLanguage := ""
		if node.PrimaryLanguage != nil {
			primaryLanguage = node.PrimaryLanguage.Name
		}

		topics := make([]string, 0, len(node.RepositoryTopics.Nodes))
		for _, topic := range node.RepositoryTopics.Nodes {
			name := strings.TrimSpace(topic.Topic.Name)
			if name == "" {
				continue
			}
			topics = append(topics, name)
		}

		repos = append(repos, Repo{
			Name:            node.Name,
			NameWithOwner:   node.NameWithOwner,
			Description:     strings.TrimSpace(node.Description),
			URL:             node.URL,
			Stars:           node.Stars,
			PrimaryLanguage: primaryLanguage,
			UpdatedAt:       node.UpdatedAt,
			IsFork:          node.IsFork,
			Topics:          topics,
		})
	}

	return StarsPage{
		Repos:      repos,
		TotalCount: query.Viewer.StarredRepositories.TotalCount,
		EndCursor:  string(query.Viewer.StarredRepositories.PageInfo.EndCursor),
		HasNext:    query.Viewer.StarredRepositories.PageInfo.HasNextPage,
	}, nil
}
