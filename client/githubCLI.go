package client

import (
	"PReQual/helper"
	"PReQual/model"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GhClient struct {
	Tokens  []string
	Current int
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatCursor(cursor string) string {
	if cursor == "" {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, cursor)
}

func NewGhClient() *GhClient {
	tokens := strings.Split(os.Getenv("GH_TOKENS"), ",")
	os.Setenv("GH_TOKEN", tokens[0])
	return &GhClient{
		Tokens:  tokens,
		Current: 0,
	}
}

func (c *GhClient) switchToken() {
	fmt.Printf("Switching token: %s -> ", c.Tokens[c.Current])
	c.Current = (c.Current + 1) % len(c.Tokens)
	os.Setenv("GH_TOKEN", c.Tokens[c.Current])
}
func (c *GhClient) GetPullRequests(repo string) ([]model.PullRequest, error) {
	owner, name, err := helper.SplitRepo(repo)
	if err != nil {
		return nil, err
	}

	var prs []model.PullRequest
	cursor := ""

	for {
		resp, err := c.fetchPullRequestPage(owner, name, cursor)
		if err != nil {
			if c.waitIfRateLimited() {
				continue
			}
			c.switchToken()
			continue
		}

		prs = append(prs, mapPRNodes(resp)...)

		if !resp.Data.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}

		cursor = resp.Data.Repository.PullRequests.PageInfo.EndCursor
	}

	return prs, nil
}

func (c *GhClient) fetchPullRequestPage(owner, name, cursor string) (*model.PullRequestResponse, error) {
	query := buildPRQuery(owner, name, cursor)

	output, err := c.runGh([]string{
		"api", "graphql",
		"-f", "query=" + query,
	})
	if err != nil {
		return nil, err
	}

	var resp model.PullRequestResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func buildPRQuery(owner, name, cursor string) string {
	return fmt.Sprintf(`
	{
	  repository(owner: "%s", name: "%s") {
	    pullRequests(first: 100, after: %s, states: [OPEN, CLOSED, MERGED]) {
	      nodes {
	        id
	        number
	        title
	        body
	        state
	        createdAt
	        closedAt
	        mergedAt
	        additions
	        deletions
	        changedFiles
	        baseRefOid
	        headRefOid
	        author { login }
	        comments(first: 100) {
	          nodes { body createdAt author { login } }
	        }
	        reviews(first: 100) {
	          nodes { body state author { login } }
	        }
	      }
	      pageInfo { hasNextPage endCursor }
	    }
	  }
	}`, owner, name, formatCursor(cursor))
}

func mapPRNodes(resp *model.PullRequestResponse) []model.PullRequest {
	var prs []model.PullRequest

	for _, n := range resp.Data.Repository.PullRequests.Nodes {
		prs = append(prs, model.PullRequest{
			Id:           n.ID,
			Number:       n.Number,
			Title:        n.Title,
			Body:         n.Body,
			State:        n.State,
			CreatedAt:    n.CreatedAt,
			ClosedAt:     deref(n.ClosedAt),
			MergedAt:     deref(n.MergedAt),
			Additions:    n.Additions,
			Deletions:    n.Deletions,
			ChangedFiles: n.ChangedFiles,
			BaseRefOid:   n.BaseRefOid,
			HeadRefOid:   n.HeadRefOid,
			Author: model.Author{
				Login: n.Author.Login,
			},
			Comments: n.Comments.Nodes,
			Reviews:  n.Reviews.Nodes,
		})
	}

	return prs
}

func (c *GhClient) GetRateLimit() (*model.RateLimitResponse, error) {
	output, err := c.runGh([]string{"api", "rate_limit"})
	if err != nil {
		return nil, fmt.Errorf("get rate limit: %w", err)
	}

	var rateLimit model.RateLimitResponse
	if err := json.Unmarshal(output, &rateLimit); err != nil {
		return nil, fmt.Errorf("decode rate limit JSON: %w", err)
	}

	return &rateLimit, nil
}

func (c *GhClient) RetrieveBranchZip(repo, sha, outputPath, outputName string) error {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/zipball/%s", repo, sha),
		"--header", "Accept: application/vnd.github+json",
	}

	maxAttempts := len(c.Tokens)
	for attempts := 0; attempts < maxAttempts; attempts++ {
		output, err := c.runGh(args)
		if err == nil {
			if err := helper.SaveToFile(outputPath, outputName, output); err != nil {
				return fmt.Errorf("save zip to %s: %w", outputPath, err)
			}
			return nil
		}
		fmt.Printf("error gh with token %s: %s\n", c.Tokens[c.Current], err)
		c.switchToken()
	}
	return fmt.Errorf("fail with all tokens %s@%s", repo, sha)
}

func (c *GhClient) waitIfRateLimited() bool {
	rl, err := c.GetRateLimit()
	if err != nil {
		return false
	}

	if rl.Rate.Remaining > 0 {
		return false
	}

	resetTime := time.Unix(rl.Rate.Reset, 0)
	waitDuration := time.Until(resetTime) + 2*time.Second

	if waitDuration <= 0 {
		return false
	}

	fmt.Printf("Rate limit exceeded. Waiting until %v...\n", resetTime)
	time.Sleep(waitDuration)

	return true
}

func (c *GhClient) decodePullRequests(output []byte) ([]model.PullRequest, error) {
	var prs []model.PullRequest
	decoder := json.NewDecoder(bytes.NewReader(output))

	for decoder.More() {
		var pr model.PullRequest
		if err := decoder.Decode(&pr); err != nil {
			return nil, fmt.Errorf("decode pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (c *GhClient) runGh(args []string) ([]byte, error) {
	cmd := exec.Command("gh", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh %v failed: %s", args, stderr.String())
	}

	return output, nil
}
