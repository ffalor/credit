/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package root

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
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
	Body   string
	Title  string
	Closed bool
}

type MergedPr struct {
	Title        string
	CreatedAt    githubv4.DateTime
	MergedAt     githubv4.DateTime
	ClosedIssues []Issue
}

type Query struct {
	Search struct {
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage githubv4.Boolean
		}
		Edges []struct {
			Node struct {
				PullRequest struct {
					Title                   githubv4.String
					CreatedAt               githubv4.DateTime
					MergedAt                githubv4.DateTime
					ClosingIssuesReferences struct {
						Nodes []struct {
							Body   githubv4.String
							Closed githubv4.Boolean
							Title  githubv4.String
						}
					} `graphql:"closingIssuesReferences(first: 100)"`
				} `graphql:"... on PullRequest"`
			}
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $searchCursor)"`
}

// NewCmdRoot represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	opts := &RootOptions{}

	cmd := &cobra.Command{
		Use:     "credit [from]",
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

	variables := map[string]interface{}{
		"query":        githubv4.String(fmt.Sprintf("is:pr is:merged author:%s merged:>%s", opts.User, opts.FromDate)),
		"searchCursor": (*githubv4.String)(nil),
	}

	var allMergedPrs []MergedPr
	for {
		var query Query

		err := opts.Client.Query(context.Background(), &query, variables)
		if err != nil {
			return err
		}

		for _, edge := range query.Search.Edges {
			node := edge.Node.PullRequest
			var closedIssues []Issue

			for _, issue := range node.ClosingIssuesReferences.Nodes {
				closedIssues = append(closedIssues, Issue{
					Body:   string(issue.Body),
					Closed: bool(issue.Closed),
					Title:  string(issue.Title),
				})
			}

			allMergedPrs = append(allMergedPrs, MergedPr{
				Title:        string(node.Title),
				CreatedAt:    node.CreatedAt,
				MergedAt:     node.MergedAt,
				ClosedIssues: closedIssues,
			})
		}

		if !bool(query.Search.PageInfo.HasNextPage) {
			break
		}

		variables["searchCursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}

	file, err := os.Create("issues.csv")
	defer file.Close()

	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"title", "description", "type"})

	for _, pr := range allMergedPrs {
		writer.Write([]string{pr.Title, "", "pr"})

		for _, issue := range pr.ClosedIssues {
			writer.Write([]string{issue.Title, issue.Body, "issue"})
		}
	}

	return nil
}

// validDate checks if a date is in the format YYYY-MM-DD
func validDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
