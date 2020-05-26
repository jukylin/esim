package tester

import (
	"github.com/jukylin/esim/log"
	"github.com/fsnotify/fsnotify"
)

type esimWatcher interface {
	// Watch folder for changes, receiver receive folder of changed files
	// Only watch file write operations
	watch(folder []string, receiver func(string))

	// close watcher
	close() error
}

type fsnotifyWatcher struct {
	*fsnotify.Watcher

	logger log.Logger
}

type fwOption func(*fsnotifyWatcher)

func newFsnotifyWatcher(options ...fwOption) esimWatcher {
	fw := &fsnotifyWatcher{}

	for _, option := range options {
		option(fw)
	}

	return fw
}

func withLogger(logger log.Logger) fwOption {
	return func(fw *fsnotifyWatcher) {
		fw.logger = logger
	}
}

func (fw *fsnotifyWatcher)  watch(folders []string, receiver func(string)) {
	if len(folders) == 0 {
		fw.logger.Errorf("There is no folder to be watch")
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fw.logger.Errorf(err.Error())
		return
	}
	fw.Watcher = watcher

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					receiver(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fw.logger.Errorf(err.Error())
			}
		}
	}()

	for _, folder := range folders {
		err = watcher.Add(folder)
		if err != nil {
			fw.logger.Errorf(err.Error())
		}
	}

	<-done
}

func (fw *fsnotifyWatcher) close() error {
	return fw.Watcher.Close()
}