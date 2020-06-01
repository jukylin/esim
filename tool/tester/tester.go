package tester

import (
	"path/filepath"

	"time"

	"os/exec"
	"sync/atomic"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)

var (
	ignoreFiles = []string{"wire.go", "wire_gen.go"}

	wireProvidersFiles = []string{"infra.go", "controllers.go"}
)

type Tester struct {
	// Watchint the directory, default current directory.
	withWatchDir string

	// If true, run wire command in the changed directory.
	withWire bool

	logger log.Logger

	watcher EsimWatcher

	execer pkg.Exec

	receiveEvent map[string]int64

	isWireFile bool

	runningWire int32

	runningTest int32

	// Wait a few seconds before run command
	waitTime time.Duration
}

type Option func(*Tester)

func NewTester(options ...Option) *Tester {
	tester := &Tester{
		receiveEvent: make(map[string]int64),
	}

	for _, option := range options {
		option(tester)
	}

	tester.waitTime = 1 * time.Second

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

	wireBool := v.GetBool(pkg.WireCmd)
	if wireBool {
		tester.withWire = true
		_, err := exec.LookPath(pkg.WireCmd)
		if err != nil {
			tester.logger.Warnf(err.Error())
			// no found, set to false.
			tester.withWire = false
		}
	}
}

// Run read directory recursively by withWatchDir and watching them
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

// receiver receive go file path of be changed and run command in the path.
// Run :
// 	1. go test (under all paths)
//  2. wire (infra/controllers)
func (tester *Tester) receiver(path string) bool {
	if path == "" {
		return false
	}

	if filepath.Ext(path) != ".go" {
		return false
	}

	fileName := filepath.Base(path)
	if tester.checkIsIgnoreFile(fileName) {
		return false
	}

	dir := filepath.Dir(path)

	if eventTime := tester.receiveEvent[dir]; eventTime == time.Now().Unix() {
		return false
	}

	tester.receiveEvent[dir] = time.Now().Unix()

	tester.checkAndRunWire(fileName, dir)

	tester.runGoTest(dir)

	return true
}

func (tester *Tester) checkIsIgnoreFile(fileName string) bool {
	if len(ignoreFiles) != 0 {
		for _, ignoreFile := range ignoreFiles {
			if ignoreFile == fileName {
				return true
			}
		}
	}

	return false
}

func (tester *Tester) runGoTest(dir string) {
	if !tester.isWireFile && atomic.CompareAndSwapInt32(&tester.runningTest, 0, 1) {
		go func() {
			tester.logger.Infof("Go file modified %s", dir)

			// Avoid redundant execution
			time.Sleep(tester.waitTime)
			tester.logger.Debugf("Running go test......")

			err := tester.execer.ExecTest(dir)
			if err != nil {
				tester.logger.Errorf(err.Error())
			}
			atomic.StoreInt32(&tester.runningTest, 0)
		}()
	}
}

// checkAndRunWire If fileName is provider file then run wire command in directory.
func (tester *Tester) checkAndRunWire(fileName, dir string) {
	if len(wireProvidersFiles) != 0 && tester.withWire &&
		atomic.CompareAndSwapInt32(&tester.runningWire, 0, 1) {
		for _, provideFile := range wireProvidersFiles {
			if provideFile == fileName {
				tester.isWireFile = true
				go func() {
					// Avoid redundant execution
					time.Sleep(tester.waitTime)

					tester.logger.Debugf("Running wire......")
					err := tester.execer.ExecWire(dir)
					if err != nil {
						tester.logger.Errorf(err.Error())
					}

					atomic.StoreInt32(&tester.runningWire, 0)
				}()
			}
		}
	}
}
