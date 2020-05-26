package tester

import (
	"testing"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	watchDir = "watch-dir"
)

func TestTesterReceive(t *testing.T)  {
	logger := log.NewLogger()
	watcher := NewTester(
		 WithTesterLogger(logger),
		 WithTesterExec(pkg.NewNullExec()))
	watcher.receive("./")
}

func TestTesterBindInput(t *testing.T)  {
	watcher := NewTester()

	v := viper.New()
	v.Set("watch_dir", watchDir)
	watcher.bindInput(v)

	assert.Equal(t, watcher.withWatchDir, watchDir)

	v.Set("watch_dir", "")
	watcher.bindInput(v)

	assert.Equal(t, watcher.withWatchDir, ".")

}