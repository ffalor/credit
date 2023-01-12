/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package root

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type RootOptions struct {
	Client   *githubv4.Client
	FromDate string
	User     string
}

type Issue struct {
	Id     string
	Body   string
	Title  string
	Url    string
	Labels []string
}

type MergedPr struct {
	Title     string
	Body      string
	Url       string
	CreatedAt string
	MergedAt  string
}

type MergedPrQuery struct {
	Search struct {
		PageInfo struct {
			EndCursor   string
			HasNextPage bool
		}
		Edges []struct {
			Node struct {
				PullRequest struct {
					Title                   string
					Body                    string
					CreatedAt               string
					MergedAt                string
					Url                     string
					ClosingIssuesReferences struct {
						Nodes []struct {
							Id     string
							Body   string
							Title  string
							Url    string
							Labels struct {
								Nodes []struct {
									Name string
								}
							} `graphql:"labels(first: 10)"`
						}
					} `graphql:"closingIssuesReferences(first: 100)"`
				} `graphql:"... on PullRequest"`
			}
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $searchCursor)"`
}

type IssueQuery struct {
	Search struct {
		PageInfo struct {
			EndCursor   string
			HasNextPage bool
		}
		Nodes []struct {
			Issue struct {
				Id     string
				Title  string
				Body   string
				Url    string
				Labels struct {
					Nodes []struct {
						Name string
					}
				} `graphql:"labels(first: 10)"`
			} `graphql:"... on Issue"`
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $searchCursor)"`
}

// NewCmdRoot represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	opts := &RootOptions{}

	cmd := &cobra.Command{
		Use:     "credit [user] -f <YYYY-MM-DD>",
		Short:   "Export all github issues into a csv file for Jira import",
		Long:    "Export all github issues from a start date into a csv file for Jira import.",
		Example: "$ credit ffalor -f 2020-01-01",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.FromDate != "" && !validDate(opts.FromDate) {
				return fmt.Errorf("invalid date format for --from, please use YYYY-MM-DD")
			} else if opts.FromDate == "" {
				opts.FromDate = time.Now().AddDate(0, 0, -90).Format("2006-01-02")
			}

			var user string

			if len(args) == 0 {
				prompt := &survey.Input{
					Message: "Please enter a user to export issues for",
				}
				survey.AskOne(prompt, &user)
			} else {
				user = args[0]
			}

			githubToken, ok := os.LookupEnv("GITHUB_TOKEN")

			if !ok {
				prompt := &survey.Password{
					Message: "Please enter your github token",
				}
				survey.AskOne(prompt, &githubToken)
			}

			src := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: githubToken},
			)
			httpClient := oauth2.NewClient(context.Background(), src)
			opts.Client = githubv4.NewClient(httpClient)
			opts.User = user

			return runRoot(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.FromDate, "from", "f", "", "Start date for issues to export (YYYY-MM-DD) (default 90 days ago")

	return cmd
}

func runRoot(opts *RootOptions) error {

	var allMergedPrs []MergedPr
	allIssues := make(map[string]Issue)

	for {
		variables := map[string]interface{}{
			"query":        githubv4.String(fmt.Sprintf("is:pr is:merged author:%s merged:>%s", opts.User, opts.FromDate)),
			"searchCursor": (*githubv4.String)(nil),
		}

		var query MergedPrQuery

		err := opts.Client.Query(context.Background(), &query, variables)
		if err != nil {
			return err
		}

		for _, edge := range query.Search.Edges {
			node := edge.Node.PullRequest

			for _, issue := range node.ClosingIssuesReferences.Nodes {

				var labels []string

				for _, label := range issue.Labels.Nodes {
					labels = append(labels, label.Name)
				}

				allIssues[issue.Id] = Issue{
					Id:     issue.Id,
					Body:   issue.Body,
					Title:  issue.Title,
					Labels: labels,
				}
			}

			allMergedPrs = append(allMergedPrs, MergedPr{
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
			"query":        githubv4.String(fmt.Sprintf("is:issue is:closed author:%s closed:>%s", opts.User, opts.FromDate)),
			"searchCursor": (*githubv4.String)(nil),
		}
		var query IssueQuery

		err := opts.Client.Query(context.Background(), &query, variables)
		if err != nil {
			return err
		}

		for _, node := range query.Search.Nodes {
			issue := node.Issue

			var labels []string

			for _, label := range issue.Labels.Nodes {
				labels = append(labels, label.Name)
			}

			allIssues[issue.Id] = Issue{
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

	file, err := os.Create("issues.csv")
	defer file.Close()

	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"title", "description", "assignee", "type"})

	for _, pr := range allMergedPrs {
		body := fmt.Sprintf("%s\nURL: %s", pr.Body, pr.Url)
		writer.Write([]string{pr.Title, body, opts.User, "pr"})
	}

	for _, issue := range allIssues {
		body := fmt.Sprintf("%s\nURL: %s", issue.Body, issue.Url)
		if len(issue.Labels) > 0 {
			body = fmt.Sprintf("%s\nLabels: %s", body, strings.Join(issue.Labels, ", "))
		}

		writer.Write([]string{issue.Title, body, opts.User, "issue"})
	}

	return nil
}

// validDate checks if a date is in the format YYYY-MM-DD
func validDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
