package util

func COALESCE[T comparable](v ...T) T {
	var defaultValue T
	for _, val := range v {
		if val != defaultValue {
			return val
		}
	}
	return defaultValue
}
