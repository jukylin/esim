package tester

import (
	"github.com/fsnotify/fsnotify"
	"github.com/jukylin/esim/log"
)

type EsimWatcher interface {
	// Watch folder for changes, receiver receive folder of changed files
	watch(folder []string, receiver func(string) bool)

	// close watcher
	close() error
}

type fsnotifyWatcher struct {
	*fsnotify.Watcher

	logger log.Logger
}

type FwOption func(*fsnotifyWatcher)

func NewFsnotifyWatcher(options ...FwOption) EsimWatcher {
	fw := &fsnotifyWatcher{}

	for _, option := range options {
		option(fw)
	}

	return fw
}

func WithFwLogger(logger log.Logger) FwOption {
	return func(fw *fsnotifyWatcher) {
		fw.logger = logger
	}
}

func (fw *fsnotifyWatcher) watch(folders []string, receiver func(string) bool) {
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

				receiver(event.Name)
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
