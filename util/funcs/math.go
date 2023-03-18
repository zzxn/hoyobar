package funcs

func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func Clip(val, min, max int) int {
	if val >= max {
		return max
	}
	if val <= min {
		return min
	}
	return val
}
