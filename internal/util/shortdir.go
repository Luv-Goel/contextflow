// Package util provides shared helper functions for ContextFlow.
package util

import "strings"

// ShortDir shortens a directory path to the last two components.
// e.g. /home/user/projects/myapp -> ~/projects/myapp
func ShortDir(dir string) string {
	parts := strings.Split(dir, "/")
	if len(parts) > 2 {
		return "~/" + strings.Join(parts[len(parts)-2:], "/")
	}
	return dir
}
