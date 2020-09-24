package tester

import (
	"path/filepath"

	"time"

	"os/exec"
	"sync/atomic"

	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)

var (
	ignoreFiles = []string{"wire.go", "wire_gen.go"}

	wireProvidersFiles = []string{"infra.go", "controllers.go"}

	runMockDir = []string{"repo", "gateway"}
)

type Tester struct {
	// 在被修改的目录下运行wire.
	withWire bool

	// 在被修改的目录下运行 mockery.
	withMockery bool

	// 在被修改的目录下运行 golangci-lint.
	withLint bool

	runningTest int32

	runningWire int32

	runningMock int32

	runningLint int32

	notRunTest bool

	// 运行命令前等待一下.
	waitTime time.Duration

	receiveEvent map[string]int64

	execer pkg.Exec

	// 监听的目录，默认当前.
	withWatchDir string

	logger log.Logger

	watcher EsimWatcher
}

type Option func(*Tester)

func NewTester(options ...Option) *Tester {
	tester := &Tester{
		receiveEvent: make(map[string]int64),
	}

	for _, option := range options {
		option(tester)
	}

	if tester.logger == nil {
		tester.logger = log.NewLogger()
	}

	if tester.execer == nil {
		tester.execer = pkg.NewCmdExec()
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

	mockeryBool := v.GetBool(pkg.MockeryCmd)
	if mockeryBool {
		tester.withMockery = true
		_, err := exec.LookPath(pkg.MockeryCmd)
		if err != nil {
			tester.logger.Warnf(err.Error())
			// no found, set to false.
			tester.withMockery = false
		}
	}

	golangciLintBool := v.GetBool(pkg.LintCmd)
	if golangciLintBool {
		tester.withLint = true
		_, err := exec.LookPath(pkg.LintCmd)
		if err != nil {
			tester.logger.Warnf(err.Error())
			// no found, set to false.
			tester.withLint = false
		}
	}
}

// 递归读取目录，并监听它们.
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

// 接收被修改的go文件路径并在对应路径运行命令
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

	// wire and mock are not running at the same time
	tester.checkAndRunWire(fileName, dir)
	tester.checkAndRunMock(dir)

	tester.runGoTest(dir)

	tester.runLint()

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

// runGoTest 运行 go test.
func (tester *Tester) runGoTest(dir string) {
	if atomic.CompareAndSwapInt32(&tester.runningTest, 0, 1) {
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

// 对指定文件，运行wire命令.
func (tester *Tester) checkAndRunWire(fileName, dir string) {
	if len(wireProvidersFiles) != 0 && tester.withWire &&
		atomic.CompareAndSwapInt32(&tester.runningWire, 0, 1) {
		for _, provideFile := range wireProvidersFiles {
			if provideFile == fileName {
				go func() {
					// Avoid redundant execution
					time.Sleep(tester.waitTime)

					tester.logger.Debugf("Running wire......")
					err := tester.execer.ExecWire(dir)
					if err != nil {
						tester.logger.Errorf(err.Error())
					}
				}()
			}
		}
		atomic.StoreInt32(&tester.runningWire, 0)
	}
}

func (tester *Tester) checkAndRunMock(dir string) {
	absDir, _ := filepath.Abs(dir)
	if tester.withMockery && tester.checkNeedRunMockDir(dir) &&
		atomic.CompareAndSwapInt32(&tester.runningMock, 0, 1) {
		go func() {
			// Avoid redundant execution
			time.Sleep(tester.waitTime)

			tester.logger.Debugf("Running mockery......")
			err := tester.execer.ExecMock(dir, "-all", "-dir", absDir)
			if err != nil {
				tester.logger.Errorf(err.Error())
			}

			atomic.StoreInt32(&tester.runningMock, 0)
		}()
	}
}

func (tester *Tester) checkNeedRunMockDir(dir string) bool {
	if len(runMockDir) != 0 {
		dirSplit := strings.Split(dir, string(filepath.Separator))
		for _, mockDir := range runMockDir {
			for _, partPath := range dirSplit {
				if mockDir == partPath {
					return true
				}
			}
		}
	}

	return false
}

func (tester *Tester) runLint() {
	if tester.withLint && atomic.CompareAndSwapInt32(&tester.runningLint, 0, 1) {
		go func() {
			// Avoid redundant execution
			time.Sleep(tester.waitTime)
			tester.logger.Debugf("Running golangci-lint ......")

			err := tester.execer.ExecLint(tester.withWatchDir)
			if err != nil {
				tester.logger.Errorf(err.Error())
			}
			atomic.StoreInt32(&tester.runningLint, 0)
		}()
	}
}
