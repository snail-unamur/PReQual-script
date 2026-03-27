package compilation

import (
	"fmt"
	"os"
	"path/filepath"
)

type PythonCompiler struct{}

func (j *PythonCompiler) CompileProject(path string) error {
	return nil
}

func (j *PythonCompiler) SetSonarProperties(path string, projectName string) error {
	content := fmt.Sprintf(`
sonar.projectKey=%s
sonar.projectName=%s
sonar.projectVersion=1.0
sonar.sources=.
sonar.language=py
sonar.sourceEncoding=UTF-8
sonar.python.version=3.10
sonar.exclusions=**/venv/**,**/__pycache__/**
sonar.host.url=%s
sonar.login=${SONAR_TOKEN}


`, projectName, projectName, os.Getenv("SONAR_DOCKER_URL"))

	propPath := filepath.Join(path, "sonar-project.properties")
	return os.WriteFile(propPath, []byte(content), 0644)
}
