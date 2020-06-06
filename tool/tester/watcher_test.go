package tester

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jukylin/esim/log"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/stretchr/testify/assert"
)

const (
	watchFolder = "./example"
)

func TestRwWatch(t *testing.T) {
	logger := log.NewLogger()

	fw := NewFsnotifyWatcher(WithFwLogger(logger))

	paths := make([]string, 0)
	paths = append(paths, "example")

	go func() {
		fw.watch(paths, func(s string) bool {
			assert.Equal(t, "example/example.go", s)
			return true
		})
	}()

	// modify example.go
	go func() {
		time.Sleep(500 * time.Millisecond)
		i := 0

		for i < 1 {
			content := `
package example

// example.
func example() bool {
	return true
}
`
			content += "// " + strconv.FormatInt(time.Now().UnixNano(), 10)

			err := filedir.EsimWrite(fmt.Sprintf("%s/%s", watchFolder, "example.go"), content)
			assert.Nil(t, err)

			i++
			time.Sleep(10 * time.Millisecond)
		}
	}()

	time.Sleep(3 * time.Second)
	fw.close()
}
