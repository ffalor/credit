/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "credit [from]",
	Short:   "Export all github issues into a csv file for Jira import",
	Long:    "Export all github issues from a start date into a csv file for Jira import.",
	Example: "$ credit 2020-01-01",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		githubToken, ok := os.LookupEnv("GITHUB_TOKEN")

		if !ok {
			prompt := &survey.Password{
				Message: "Please enter your github token",
			}

			survey.AskOne(prompt, &githubToken)
		}

		// get all issues closed after the from date
		//src := oauth2.StaticTokenSource(
		//	&oauth2.Token{AccessToken: githubToken},
		//)
		//httpClient := oauth2.NewClient(context.Background(), src)

		//client := githubv4.NewClient(httpClient)

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.credit.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
