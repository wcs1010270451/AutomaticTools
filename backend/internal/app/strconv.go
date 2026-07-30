package app

import "fmt"

func fmtSprint(value any) string {
	return fmt.Sprint(value)
}

func fmtSscan(text string, target any) (int, error) {
	return fmt.Sscan(text, target)
}
