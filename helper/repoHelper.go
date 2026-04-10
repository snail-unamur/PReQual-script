package helper

import (
	"fmt"
	"strings"
)

func SplitRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format")
	}
	return parts[0], parts[1], nil
}
