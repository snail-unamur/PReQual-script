package metric

import (
	"PReQual/compilation"
	"PReQual/helper"
	"PReQual/model"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type SonarQubeAnalyzer struct{}

func (a *SonarQubeAnalyzer) AnalyzeProjectBranch(
	branchType string,
	prID string,
	repoName string,
	basePath string,
	metrics []string,
) (map[string]interface{}, error) {

	archivePath := filepath.Join(basePath, branchType+".zip")
	defer cleanupExtractedDir(archivePath)

	compiler := compilation.Compiler(&compilation.JavaCompiler{})

	return analyzeArchive(
		branchType,
		archivePath,
		prID,
		repoName,
		metrics,
		compiler,
	)
}

func analyzeArchive(
	branchType string,
	archivePath string,
	prID string,
	repoName string,
	metrics []string,
	compiler compilation.Compiler,
) (map[string]interface{}, error) {

	if _, err := os.Stat(archivePath); err != nil {
		return nil, fmt.Errorf("archive not found: %s", archivePath)
	}

	extractDir := strings.TrimSuffix(archivePath, ".zip")
	if err := helper.Unzip(archivePath, extractDir); err != nil {
		return nil, err
	}

	projectRoot, err := helper.FindProjectRoot(extractDir)
	if err != nil {
		return nil, err
	}

	if err := compiler.CompileProject(projectRoot); err != nil {
		return nil, fmt.Errorf("compile failed: %w", err)
	}

	projectKey := buildProjectKey(repoName, prID, branchType)

	if err := compiler.SetSonarProperties(projectRoot, projectKey); err != nil {
		return nil, err
	}

	if err := runSonarScanner(projectRoot); err != nil {
		return nil, err
	}

	if err := waitForAnalysisCompletion(projectKey, metrics, 5*time.Minute); err != nil {
		return nil, err
	}

	measures, err := retrieveSonarMetrics(projectKey, metrics)
	if err != nil {
		return nil, err
	}

	return helper.ConvertMeasuresToMap(measures), nil
}

func buildProjectKey(repoName, prID, branchType string) string {
	repo := strings.ReplaceAll(repoName, "/", "-")
	return fmt.Sprintf("%s-%s-%s", prID, repo, branchType)
}

func cleanupExtractedDir(archivePath string) {
	dir := strings.TrimSuffix(archivePath, ".zip")
	_ = os.RemoveAll(dir)
}

func waitForAnalysisCompletion(
	projectKey string,
	metrics []string,
	timeout time.Duration,
) error {

	client := helper.NewHTTPClient(
		os.Getenv("SONAR_URL"),
		os.Getenv("SONAR_TOKEN"),
	)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if time.Now().After(deadline) {
			return fmt.Errorf("sonar analysis timeout after %v", timeout)
		}

		var resp model.SonarMeasures
		path := fmt.Sprintf(
			"/api/measures/component?metricKeys=%s&component=%s",
			strings.Join(metrics, ","),
			projectKey,
		)

		if err := client.DoRequest("GET", path, nil, &resp); err == nil {
			if len(resp.Component.Measures) == len(metrics) {
				return nil
			}
		}
	}

	return nil
}

func runSonarScanner(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"docker", "run", "--rm",
		"--network", os.Getenv("DOCKER_NET"),
		"-v", absPath+":/usr/src",
		"-e", "SONAR_TOKEN="+os.Getenv("SONAR_TOKEN"),
		"sonarsource/sonar-scanner-cli",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func retrieveSonarMetrics(
	projectKey string,
	metrics []string,
) (model.SonarMeasures, error) {

	client := helper.NewHTTPClient(
		os.Getenv("SONAR_URL"),
		os.Getenv("SONAR_TOKEN"),
	)

	path := fmt.Sprintf(
		"/api/measures/component?metricKeys=%s&component=%s",
		strings.Join(metrics, ","),
		projectKey,
	)

	var resp model.SonarMeasures
	if err := client.DoRequest("GET", path, nil, &resp); err != nil {
		return model.SonarMeasures{}, err
	}

	return resp, nil
}
