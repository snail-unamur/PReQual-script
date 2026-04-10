package model

type PullRequest struct {
	Id           string    `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Author       Author    `json:"author"`
	Body         string    `json:"body"`
	BaseRefOid   string    `json:"baseRefOid"`
	HeadRefOid   string    `json:"headRefOid"`
	CreatedAt    string    `json:"createdAt"`
	ClosedAt     string    `json:"closedAt"`
	MergedAt     string    `json:"mergedAt"`
	State        string    `json:"state"`
	Comments     []Comment `json:"comments"`
	Reviews      []Review  `json:"reviews"`
	ChangedFiles int       `json:"changedFiles"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
}
