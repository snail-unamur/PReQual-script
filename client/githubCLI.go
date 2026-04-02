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

func NewGhClient() *GhClient {
	tokens := strings.Split(os.Getenv("GH_TOKENS"), ",")
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

const data = "id,number,title,author,baseRefOid,headRefOid,state,createdAt,closedAt,comments,body,closingIssuesReferences,reviews,mergedAt,reviewDecision,additions,deletions,changedFiles"

func (c *GhClient) GetPullRequests(repo string) ([]model.PullRequest, error) {
	limits := helper.GenerateLimits(10000)

	for tokenIndex := 0; tokenIndex < len(c.Tokens); tokenIndex++ {
		var previousLimit int
		for i, limit := range limits {

			if i == 0 || limit != previousLimit {
				fmt.Printf("Changement de limite détecté : nouvelle limite = %d\n", limit)
				previousLimit = limit
			}

			prs, err := c.fetchPullRequests(repo, limit)
			if err == nil {
				return prs, nil
			}

			if handled := c.waitIfRateLimited(); handled {
				continue
			}
		}
		c.switchToken()
	}
	return nil, fmt.Errorf("impossible de récupérer les pull requests après avoir essayé tous les tokens et toutes les limites")
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

func (c *GhClient) fetchPullRequests(repo string, limit int) ([]model.PullRequest, error) {
	args := c.buildPRArgs(repo, limit)

	output, err := c.runGh(args)
	if err != nil {
		return nil, err
	}

	return c.decodePullRequests(output)
}

func (c *GhClient) buildPRArgs(repo string, limit int) []string {
	return []string{
		"pr", "list",
		"-R", repo,
		"--state", "all",
		"--limit", fmt.Sprintf("%d", limit),
		"--json", data,
	}
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
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("decode pull requests JSON: %w", err)
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
