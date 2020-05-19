package pkg

import (
	"path/filepath"
	"strings"
)

// ./a/b/c/ => /a/b/c
func DirPathToImportPath(dirPath string) string {
	newPath := strings.TrimLeft(dirPath, ".")
	newPath = strings.Trim(dirPath, "/")
	newPath = string(filepath.Separator) + newPath
	return newPath
}
