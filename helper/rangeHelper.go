package helper

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRange(rangeStr string) ([2]int, error) {
	var result [2]int

	if rangeStr == "" {
		return result, nil
	}

	parts := strings.Split(rangeStr, ",")
	if len(parts) != 2 {
		return result, fmt.Errorf("range must be in format 'min,max'")
	}

	min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return result, fmt.Errorf("invalid min value: %w", err)
	}

	max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return result, fmt.Errorf("invalid max value: %w", err)
	}

	return [2]int{min, max}, nil
}

func IsInRange(value int, r [2]int) bool {
	if r == [2]int{0, 0} {
		return true
	}
	return value >= r[0] && value <= r[1]
}
