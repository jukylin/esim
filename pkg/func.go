package pkg

import (
	"path/filepath"
	"strings"
)

// ./a/b/c/ => /a/b/c
func DirPathToImportPath(dirPath string) string {
	path := strings.TrimLeft(dirPath, ".")
	path = strings.Trim(dirPath, "/")
	path = string(filepath.Separator) + path
	return path
}
