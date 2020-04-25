package pkg

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestDirPathToImportPath(t *testing.T) {
	assert.Equal(t, "/a/b/c", DirPathToImportPath("/a/b/c"))
}