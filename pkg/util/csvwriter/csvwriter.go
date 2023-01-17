package csvwriter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/ffalor/credit/pkg/util/types"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(user string, allMergedPrs map[string]types.MergedPr, allIssues map[string]types.Issue) error {
	file, err := os.Create("issues.csv")
	defer file.Close()

	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"title", "description", "assignee", "repo", "type"})

	for _, pr := range allMergedPrs {
		body := fmt.Sprintf("%s\nURL: %s", pr.Body, pr.Url)
		writer.Write([]string{pr.Title, body, user, pr.RepoName, "pr"})
	}

	for _, issue := range allIssues {
		body := fmt.Sprintf("%s\nURL: %s", issue.Body, issue.Url)
		if len(issue.Labels) > 0 {
			body = fmt.Sprintf("%s\nLabels: %s", body, strings.Join(issue.Labels, ", "))
		}

		writer.Write([]string{issue.Title, body, user, issue.RepoName, "issue"})
	}

	return nil
}
