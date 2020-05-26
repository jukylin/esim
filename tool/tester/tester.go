package tester

import (
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg"
)

type Tester struct {

	// watch the dir, default "."
	withWatchDir string

	logger log.Logger

	watcher EsimWatcher

	execer pkg.Exec
}

type Option func(*Tester)

func NewTester(options ...Option) *Tester {
	tester := &Tester{}
	for _, option := range options {
		option(tester)
	}

	return tester
}

func WithTesterLogger(logger log.Logger) Option {
	return func(t *Tester) {
		t.logger = logger
	}
}

func WithTesterWatcher(watcher EsimWatcher) Option {
	return func(t *Tester) {
		t.watcher = watcher
	}
}

func WithTesterExec(execer pkg.Exec) Option {
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


func (tester *Tester) Run(v *viper.Viper) {
	tester.bindInput(v)

	paths, err := filedir.ReadDir(tester.withWatchDir)
	if err != nil {
		tester.logger.Fatalf(err.Error())
	}
	paths = append(paths, tester.withWatchDir)

	tester.watcher.watch(paths, tester.receive)
}

// receive file path of be changed and run go test
func (tester *Tester) receive(path string)  {
	dir := filepath.Dir(path)

	err := tester.execer.ExecTest(dir)
	if err != nil {
		tester.logger.Errorf(err.Error())
	}
}