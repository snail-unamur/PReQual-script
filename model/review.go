package model

type Review struct {
	Author      Author `json:"author"`
	State       string `json:"state"`
	Body        string `json:"body"`
	SubmittedAt string `json:"submittedAt"`
}
