package nodes

import "strings"

func split(path string) []string {
	return strings.Split(path, "/")
}
