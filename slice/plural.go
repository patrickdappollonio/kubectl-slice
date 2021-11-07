package slice

func pluralize(s string, n int) string {
	if n == 1 {
		return s
	}
	return s + "s"
}
