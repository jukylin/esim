package tester

import (
	"testing"
	// "github.com/jukylin/esim/log"
)

func TestTesterReceive(t *testing.T)  {
	// logger := log.NewLogger()
	watcher := NewTester(
		// WithTesterLogger(logger),
		// WithTesterExec(pkg.NewCmdExec()),
		// WithTesterWatcher(newFsnotifyWatcher(
		//	withLogger(logger),
		// )),
		)

	watcher.receive("./test/tester.go")
}