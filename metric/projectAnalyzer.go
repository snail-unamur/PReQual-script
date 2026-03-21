package metric

type ProjectAnalyser interface {
	AnalyzeProjectBranch(
		branchType string,
		prID string,
		repoName string,
		basePath string,
		metrics []string,
	) (map[string]interface{}, error)
}
