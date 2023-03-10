package types

type Issue struct {
	Id       string
	RepoName string
	Body     string
	Title    string
	Url      string
	Epic     string
	Labels   []string
}

type MergedPr struct {
	Id        string
	RepoName  string
	Title     string
	Body      string
	Url       string
	CreatedAt string
	Epic      string
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
					Id                      string
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
					BaseRepository struct {
						Name string
					} `graphql:"baseRepository"`
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
				Repository struct {
					Name string
				} `graphql:"repository"`
			} `graphql:"... on Issue"`
		}
	} `graphql:"search(query: $query, type: ISSUE, first: 100, after: $searchCursor)"`
}
