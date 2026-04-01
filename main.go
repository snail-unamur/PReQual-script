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
	rangeArg := flag.String("range", "", "PR range to analyze, e.g., 0,100")

	flag.Parse()

	if *reposArg == "" {
		flag.Usage()
		os.Exit(1)
	}

	database.InitMongoDB(os.Getenv("MONGODB_URL"))

	repos := strings.Split(*reposArg, ",")
	metrics := strings.Split(*metricsArg, ",")
	prRange, err := helper.ParseRange(*rangeArg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	prClient := client.PullRequestClient(client.NewGhClient())
	analyzer := metric.ProjectAnalyser(&metric.SonarQubeAnalyzer{})

	for _, repoKey := range repos {
		if err := processRepo(repoKey, *workspace, metrics, prClient, analyzer, prRange); err != nil {
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
	prRange [2]int,
) error {

	fmt.Printf("\n===== Repo: %s =====\n", repoKey)

	parts := strings.Split(repoKey, "/")
	org, repo := parts[0], parts[1]

	prs, err := prClient.GetPullRequests(repoKey)
	if err != nil {
		return err
	}

	for _, pr := range prs {
		if !helper.IsInRange(pr.Number, prRange) {
			fmt.Printf("PR #%d skipped. This PR is outside the specified range.\n", pr.Number)
			continue
		}
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

	if helper.IsPRDirExist(path) && pr.State != "OPEN" {
		fmt.Printf("PR #%d skipped. This PR was already analyzed.\n", pr.Number)
		return nil
	}

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
