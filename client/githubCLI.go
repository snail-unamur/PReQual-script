package client

import (
	"PReQual/helper"
	"PReQual/model"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type GhClient struct{}

const data = "number,title,baseRefOid,headRefOid,state,createdAt,closedAt,comments,body,closingIssuesReferences,reviews"

func (c *GhClient) GetPullRequests(repo string) ([]model.PullRequest, error) {
	limits := helper.GenerateLimits(10000)

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

	return nil, fmt.Errorf("impossible de récupérer les pull requests après plusieurs tentatives")
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

	output, err := c.runGh(args)
	if err != nil {
		return fmt.Errorf("download zip for %s@%s: %w", repo, sha, err)
	}

	if err := helper.SaveToFile(outputPath, outputName, output); err != nil {
		return fmt.Errorf("save zip to %s: %w", outputPath, err)
	}

	return nil
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
