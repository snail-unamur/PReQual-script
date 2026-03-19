package helper

func GenerateLimits(max int) []int {
	steps := []int{1000, 100, 10, 1}
	var limits []int

	current := max

	for i, step := range steps {
		for v := current; v >= step; v -= step {
			limits = append(limits, v)
		}

		if i < len(steps)-1 {
			nextStep := steps[i+1]
			current = step - nextStep
		}
	}

	return limits
}
