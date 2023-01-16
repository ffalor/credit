package gh

import (
	"context"
	"fmt"

	"github.com/ffalor/credit/pkg/util/types"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Gh struct {
	Token  string
	Client *githubv4.Client
}

func NewGh(token string) *Gh {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	return &Gh{
		Token:  token,
		Client: client,
	}
}

// GetIssues returns all merged PRs and closed issues for a given user
func (g *Gh) GetIssues(user string, fromDate string) ([]types.MergedPr, map[string]types.Issue, error) {
	var allMergedPrs []types.MergedPr
	allIssues := make(map[string]types.Issue)

	for {
		variables := map[string]interface{}{
			"query":        githubv4.String(fmt.Sprintf("is:pr is:merged author:%s merged:>%s", user, fromDate)),
			"searchCursor": (*githubv4.String)(nil),
		}

		var query types.MergedPrQuery

		err := g.Client.Query(context.Background(), &query, variables)
		if err != nil {
			return allMergedPrs, allIssues, err
		}

		for _, edge := range query.Search.Edges {
			node := edge.Node.PullRequest

			for _, issue := range node.ClosingIssuesReferences.Nodes {

				var labels []string

				for _, label := range issue.Labels.Nodes {
					labels = append(labels, label.Name)
				}

				allIssues[issue.Id] = types.Issue{
					Id:     issue.Id,
					Body:   issue.Body,
					Title:  issue.Title,
					Labels: labels,
				}
			}

			allMergedPrs = append(allMergedPrs, types.MergedPr{
				Title:     node.Title,
				Body:      node.Body,
				Url:       node.Url,
				CreatedAt: node.CreatedAt,
				MergedAt:  node.MergedAt,
			})
		}

		if !query.Search.PageInfo.HasNextPage {
			break
		}

		variables["searchCursor"] = query.Search.PageInfo.EndCursor
	}

	for {
		variables := map[string]interface{}{
			"query":        githubv4.String(fmt.Sprintf("is:issue is:closed author:%s closed:>%s", user, fromDate)),
			"searchCursor": (*githubv4.String)(nil),
		}
		var query types.IssueQuery

		err := g.Client.Query(context.Background(), &query, variables)
		if err != nil {
			return allMergedPrs, allIssues, err
		}

		for _, node := range query.Search.Nodes {
			issue := node.Issue

			var labels []string

			for _, label := range issue.Labels.Nodes {
				labels = append(labels, label.Name)
			}

			allIssues[issue.Id] = types.Issue{
				Id:     issue.Id,
				Body:   issue.Body,
				Title:  issue.Title,
				Labels: labels,
			}
		}

		if !query.Search.PageInfo.HasNextPage {
			break
		}

		variables["searchCursor"] = query.Search.PageInfo.EndCursor
	}

	return allMergedPrs, allIssues, nil
}
