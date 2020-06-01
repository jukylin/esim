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
	loggerOptions = log.LoggerOptions{}
	logger        = log.NewLogger(loggerOptions.WithDebug(true))
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

	tester.isWireFile = true
	tester.runGoTest("./")
	assert.Equal(t, int32(0), tester.runningTest)

	tester.isWireFile = false
	tester.runGoTest("./")
	assert.Equal(t, int32(1), tester.runningTest)
}
