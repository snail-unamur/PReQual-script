package helper

import (
	"PReQual/model"
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteMetaDataFile(path string, pr model.PullRequest) {
	writePullRequestMetaFile(path, pr)
	writeCommentsMetaFile(path, pr)
	writeReviewsMetaFile(path, pr)
}

func writeReviewsMetaFile(path string, pr model.PullRequest) {
	type ReviewJSON struct {
		Author      string `json:"author"`
		SubmittedAt string `json:"submitted_at"`
		State       string `json:"state"`
		Body        string `json:"body"`
	}

	reviews := make([]ReviewJSON, 0, len(pr.Reviews))
	for _, c := range pr.Reviews {
		reviews = append(reviews, ReviewJSON{
			Author:      c.Author.Login,
			SubmittedAt: c.SubmittedAt,
			State:       c.State,
			Body:        c.Body,
		})
	}

	reviewsData, err := json.MarshalIndent(reviews, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(path, "reviews.json"), reviewsData, 0644)
	if err != nil {
		panic(err)
	}
}

func writeCommentsMetaFile(path string, pr model.PullRequest) {
	type CommentJSON struct {
		Author    string `json:"author"`
		CreatedAt string `json:"created_at"`
		Body      string `json:"body"`
	}

	comments := make([]CommentJSON, 0, len(pr.Comments))
	for _, c := range pr.Comments {
		comments = append(comments, CommentJSON{
			Author:    c.Author.Login,
			CreatedAt: c.CreatedAt,
			Body:      c.Body,
		})
	}

	commentsData, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(path, "comments.json"), commentsData, 0644)
	if err != nil {
		panic(err)
	}
}

func writePullRequestMetaFile(path string, pr model.PullRequest) {
	metadata := map[string]interface{}{
		"title":      pr.Title,
		"body":       pr.Body,
		"created_at": pr.CreateAt,
		"closed_at":  pr.ClosedAt,
		"state":      pr.State,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}

	filePath := filepath.Join(path, "metadata.json")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		panic(err)
	}
}
