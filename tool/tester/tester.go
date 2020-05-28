package tester

import (
	"path/filepath"

	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)

type Tester struct {

	// watch the dir, default "."
	withWatchDir string

	logger log.Logger

	watcher EsimWatcher

	execer pkg.Exec

	receiveEvent map[string]int64
}

type Option func(*Tester)

func NewTester(options ...Option) *Tester {
	tester := &Tester{}
	for _, option := range options {
		option(tester)
	}

	tester.receiveEvent = make(map[string]int64)

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

	absDir, err := filepath.Abs(tester.withWatchDir)
	if err != nil {
		tester.logger.Fatalf(err.Error())
	}

	tester.logger.Infof("Watching : %s", absDir)

	tester.watcher.watch(paths, tester.receiver)
}

// receiver receive go file path of be changed and run go test
func (tester *Tester) receiver(path string) bool {
	if filepath.Ext(path) != ".go" {
		return false
	}

	dir := filepath.Dir(path)

	if eventTime := tester.receiveEvent[dir]; eventTime == time.Now().Unix() {
		return false
	}

	tester.receiveEvent[dir] = time.Now().Unix()

	tester.logger.Infof("Go file modified %s", path)

	err := tester.execer.ExecTest(dir)
	if err != nil {
		tester.logger.Errorf(err.Error())
		return false
	}

	return true
}
