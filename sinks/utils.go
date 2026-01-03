package sinks

import "fmt"

func GetLogLevelIndex[T comparable](a T, list []T) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return 1
}

func PadRight(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}
