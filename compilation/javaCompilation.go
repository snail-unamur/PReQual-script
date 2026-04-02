package compilation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type JavaCompiler struct{}

func (j *JavaCompiler) CompileProject(path string) error {
	// Vérifie si pom.xml existe
	pom := filepath.Join(path, "pom.xml")
	if _, err := exec.LookPath("mvn"); err != nil {
		return fmt.Errorf("maven not found in PATH")
	}
	if _, err := os.Stat(pom); os.IsNotExist(err) {
		// Pas de pom.xml => rien à compiler
		return nil
	}

	cmd := exec.Command("mvn", "clean", "package", "-DskipTests")
	cmd.Dir = path
	cmd.Stdout = nil
	cmd.Stderr = nil

	fmt.Printf("Compiling Maven project at %s\n", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("maven build failed: %v", err)
	}

	return nil
}

func (j *JavaCompiler) SetSonarProperties(path string, projectName string) error {
	// Vérifie si target/classes existe
	binaries := ""
	targetClasses := filepath.Join(path, "target", "classes")
	if _, err := os.Stat(targetClasses); err == nil {
		binaries = "sonar.java.binaries=target/classes\n"
	}

	content := fmt.Sprintf(`
sonar.projectKey=%s
sonar.projectName=%s
sonar.sources=.
sonar.sourceEncoding=UTF-8
%s
sonar.host.url=%s
`, projectName, projectName, binaries, os.Getenv("SONAR_DOCKER_URL"))

	propPath := filepath.Join(path, "sonar-project.properties")
	return os.WriteFile(propPath, []byte(content), 0644)
}
