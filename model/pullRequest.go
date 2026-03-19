package model

type PullRequest struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	BaseRefOid string    `json:"baseRefOid"`
	HeadRefOid string    `json:"headRefOid"`
	CreateAt   string    `json:"createdAt"`
	ClosedAt   string    `json:"closedAt"`
	State      string    `json:"state"`
	Comments   []Comment `json:"comments"`
	Reviews    []Review  `json:"reviews"`
}
