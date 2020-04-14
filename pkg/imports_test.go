package pkg


import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestImports_EmptyString(t *testing.T) {
	imports := Imports{}
	str := imports.String()
	assert.Empty(t, str)
}


func TestImports_String(t *testing.T) {
	var result = `import (

	//time
	//is a test
	time "time"

	//sync
	//is a test
	sync "sync"
)`

	imports := Imports{}

	docs1 := []string{"//time", "//is a test"}
	docs2 := []string{"//sync", "//is a test"}

	imports = append(imports, Import{Name: "time", Path: "time", Doc: docs1})
	imports = append(imports, Import{Name: "sync", Path: "sync", Doc: docs2})

	src := imports.String()
	assert.Equal(t, result, string(src))
}