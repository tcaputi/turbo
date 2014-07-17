package turbo

import (
	"strings"
)

const (
	SLASH = "/"
	DOT   = "."
)

func joinPaths(basePath string, extension string) string {
	if strings.HasSuffix(basePath, SLASH) {
		if strings.HasPrefix(extension, SLASH) {
			return basePath + extension[1:]
		} else {
			return basePath + extension
		}
	} else {
		if strings.HasPrefix(extension, SLASH) {
			return basePath + extension
		} else {
			return basePath + SLASH + extension
		}
	}
}

func mongoizePath(path string) string {
	return strings.Replace(strings.Trim(path, SLASH), SLASH, DOT, -1)
}

func standardizePath(path string) string {
	if path[0] != SLASH[0] {
		if path[len(path)-1] == SLASH[0] {
			return SLASH + path[:(len(path)-1)]
		} else {
			return SLASH + path
		}
	} else {
		if path[len(path)-1] == SLASH[0] {
			return path[:(len(path) - 1)]
		} else {
			return path
		}
	}
}
