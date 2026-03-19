package model

type Comment struct {
	Author    Author `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}
