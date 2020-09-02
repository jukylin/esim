package tester

import (
	"testing"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	watchDir    = "watch-dir"
	receivePath = "./test.go"
	wirePath    = "./wire.go"
)

var (
	logger = log.NewLogger(log.WithDebug(true))
)

func TestTesterReceive(t *testing.T) {
	tester := NewTester(
		WithTesterLogger(logger),
		WithTesterExec(pkg.NewNullExec()))

	result := tester.receiver(receivePath)
	assert.True(t, result)

	result = tester.receiver(receivePath)
	assert.False(t, result)
}

func TestTesterIgnoreFiles(t *testing.T) {
	tester := NewTester(
		WithTesterLogger(logger),
		WithTesterExec(pkg.NewNullExec()))

	result := tester.receiver(receivePath)
	assert.True(t, result)

	result = tester.receiver(wirePath)
	assert.False(t, result)
}

func TestTesterBindInput(t *testing.T) {
	tester := NewTester()

	v := viper.New()
	v.Set("watch_dir", watchDir)
	tester.bindInput(v)

	assert.Equal(t, tester.withWatchDir, watchDir)

	v.Set("watch_dir", "")
	tester.bindInput(v)

	assert.Equal(t, tester.withWatchDir, ".")
}

func TestTesterCheckAndRunWire(t *testing.T) {
	tester := NewTester(
		WithTesterLogger(logger),
		WithTesterExec(pkg.NewNullExec()))

	tester.withWire = true
	tester.waitTime = 20 * time.Millisecond

	tester.checkAndRunWire("infra.go", "./")
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, tester.runningWire, int32(1))
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, tester.runningWire, int32(0))
}

func TestTesterRunGoTest(t *testing.T) {
	tester := NewTester(
		WithTesterLogger(logger),
		WithTesterExec(pkg.NewNullExec()))

	tester.notRunTest = true
	tester.runGoTest("./")
	assert.Equal(t, int32(0), tester.runningTest)

	tester.notRunTest = false
	tester.runGoTest("./")
	assert.Equal(t, int32(1), tester.runningTest)
}

func TestTester_checkNeedRunMockDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"need run mock", "/test/internal/infra/repo/test.go", true},
		{"not need", "/test/internal/infra/test.go", false},
		{"need run mock", "/test/internal/infra/gateway/test.go", true},
	}

	tester := NewTester()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tester.checkNeedRunMockDir(tt.dir); got != tt.want {
				t.Errorf("Tester.checkNeedRunMockDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTester_checkAndRunMock(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		withMockery bool
	}{
		{"run mockery", "./repo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTester()
			tester.withMockery = tt.withMockery
			tester.checkAndRunMock(tt.dir)
		})
	}
}

func TestTester_runLint(t *testing.T) {
	tests := []struct {
		name        string
		watchDir    string
		withMockery bool
	}{
		{"run lint", "./repo", true},
	}

	execer := pkg.NewNullExec()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTester(WithTesterLogger(log.NewLogger(log.WithDebug(true))))
			tester.execer = execer
			tester.withWatchDir = tt.watchDir
			tester.runLint()
			tester.runLint()
		})
	}
}
