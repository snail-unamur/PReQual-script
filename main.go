package main

import (
	"PReQual/client"
	"PReQual/database"
	"PReQual/helper"
	"PReQual/metric"
	"PReQual/model"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultWorkspace = "tmp"
	defaultMetrics   = "complexity,cognitive_complexity"
)

func main() {
	reposArg := flag.String("repos", "", "owner/repo,owner/repo (required)")
	workspace := flag.String("workspace", defaultWorkspace, "Workspace directory")
	metricsArg := flag.String("metrics", defaultMetrics, "Comma-separated metrics")

	flag.Parse()

	if *reposArg == "" {
		flag.Usage()
		os.Exit(1)
	}

	database.InitMongoDB(os.Getenv("MONGODB_URL"))

	repos := strings.Split(*reposArg, ",")
	metrics := strings.Split(*metricsArg, ",")

	prClient := client.PullRequestClient(client.NewGhClient())
	analyzer := metric.ProjectAnalyser(&metric.SonarQubeAnalyzer{})

	for _, repoKey := range repos {
		if err := processRepo(repoKey, *workspace, metrics, prClient, analyzer); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}
}

func processRepo(
	repoKey string,
	workspace string,
	metrics []string,
	prClient client.PullRequestClient,
	analyzer metric.ProjectAnalyser,
) error {

	fmt.Printf("\n===== Repo: %s =====\n", repoKey)

	parts := strings.Split(repoKey, "/")
	org, repo := parts[0], parts[1]

	prs, err := prClient.GetPullRequests(repoKey)
	if err != nil {
		return err
	}

	for _, pr := range prs {
		if err := processPR(repoKey, org, repo, pr, workspace, metrics, prClient, analyzer); err != nil {
			return err
		}
	}

	return nil
}

func processPR(
	repoKey, org, repo string,
	pr model.PullRequest,
	workspace string,
	metrics []string,
	prClient client.PullRequestClient,
	analyzer metric.ProjectAnalyser,
) error {

	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)

	path := fmt.Sprintf("%s/%s/pr_%d", workspace, repoKey, pr.Number)
	start := time.Now()

	if err := prClient.RetrieveBranchZip(repoKey, pr.HeadRefOid, path, "head.zip"); err != nil {
		return err
	}
	if err := prClient.RetrieveBranchZip(repoKey, pr.BaseRefOid, path, "base.zip"); err != nil {
		return err
	}

	baseMetrics, err := analyzer.AnalyzeProjectBranch("base", pr.Id, repoKey, path, metrics)
	if err != nil {
		return err
	}

	headMetrics, err := analyzer.AnalyzeProjectBranch("head", pr.Id, repoKey, path, metrics)
	if err != nil {
		return err
	}

	stats := model.AnalysisStat{
		TotalTime: int(time.Since(start).Seconds()),
		BaseSize:  helper.FormatSizeRounded([]string{path + "/base.zip"}),
		HeadSize:  helper.FormatSizeRounded([]string{path + "/head.zip"}),
	}

	database.InsertPR(org, repo, pr, headMetrics, baseMetrics, stats)
	return nil
}
