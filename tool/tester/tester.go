package tester

import (

	"github.com/spf13/viper"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg"
	"os"
	"github.com/davecgh/go-spew/spew"
)

type Tester struct {

	// watch the dir
	withWatchDir string

	logger log.Logger

	watcher esimWatcher

	execer pkg.Exec
}

type TesterOption func(*Tester)

func NewTester(options ...TesterOption) *Tester {
	tester := &Tester{}
	for _, option := range options {
		option(tester)
	}

	return tester
}

func WithTesterLogger(logger log.Logger) TesterOption {
	return func(t *Tester) {
		t.logger = logger
	}
}

func WithTesterWatcher(watcher esimWatcher) TesterOption {
	return func(t *Tester) {
		t.watcher = watcher
	}
}

func WithTesterExec(execer pkg.Exec) TesterOption {
	return func(t *Tester) {
		t.execer = execer
	}
}

func (tester *Tester) bindInput(v *viper.Viper) {
	watchDir := v.GetString("watch_dir")
	if watchDir == "" {
		watchDir = "."
	}
	tester.withWatchDir = watchDir
}

func (tester *Tester) run() {
	paths, err := filedir.ReadDir(tester.withWatchDir)
	if err != nil {
		tester.logger.Fatalf(err.Error())
	}
	paths = append(paths, tester.withWatchDir)

	tester.watcher.watch(paths, tester.receive)

}

// receive file path of be changed
func (tester *Tester) receive(path string)  {

	fileInfo, _ := os.Stat(path)
	//fileInfo.
spew.Dump(fileInfo)
	//tester.exec.ExecTest()
}