package utils

func TruncateString(s string, length int) string {
	isLenExceeded := len(s) > length
	if isLenExceeded {
		return s[:length] + "..."
	}

	return s
}
