package model

type PullRequestResponse struct {
	Data struct {
		Repository struct {
			PullRequests struct {
				Nodes []struct {
					ID           string
					Number       int
					Title        string
					Body         string
					State        string
					CreatedAt    string
					ClosedAt     *string
					MergedAt     *string
					Additions    int
					Deletions    int
					ChangedFiles int
					BaseRefOid   string
					HeadRefOid   string

					Author struct {
						Login string
					}

					Comments struct {
						Nodes []Comment
					}

					Reviews struct {
						Nodes []Review
					}
				}

				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			}
		}
	}
}
