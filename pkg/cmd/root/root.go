/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package root

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ffalor/credit/pkg/util/csvwriter"
	"github.com/ffalor/credit/pkg/util/gh"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	gh       *gh.Gh
	FromDate string
	User     string
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

			gh := gh.NewGh(githubToken)

			opts.gh = gh
			opts.User = user

			return runRoot(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.FromDate, "from", "f", "", "Start date for issues to export (YYYY-MM-DD) (default 90 days ago")

	return cmd
}

func runRoot(opts *RootOptions) error {

	// Get all merged PRs and issues
	allMergedPrs, allIssues, err := opts.gh.GetIssues(opts.User, opts.FromDate)
	if err != nil {
		return err
	}

	// Write to csv file issues.csv
	csvwriter := csvwriter.NewWriter()
	csvwriter.Write(opts.User, allMergedPrs, allIssues)

	return nil
}

// validDate checks if a date is in the format YYYY-MM-DD
func validDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
