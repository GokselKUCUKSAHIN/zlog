package internal

import (
	"os"
	"sync"
	"time"
)

type FileWatcher struct {
	path      string
	onChange  func()
	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

func NewFileWatcher(
	path string,
	onChange func(),
) *FileWatcher {
	return &FileWatcher{
		path:     path,
		onChange: onChange,
		done:     make(chan struct{}),
	}
}

func (w *FileWatcher) Start() error {
	if w.onChange == nil {
		return nil
	}

	var startErr error
	w.startOnce.Do(func() {
		info, err := os.Stat(w.path)
		if err != nil {
			startErr = err
			return
		}

		lastModTime := info.ModTime()

		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-w.done:
					return
				case <-ticker.C:
					fileInfo, err := os.Stat(w.path)
					if err != nil {
						continue
					}

					if fileInfo.ModTime().After(lastModTime) {
						lastModTime = fileInfo.ModTime()
						w.onChange()
					}
				}
			}
		}()
	})

	return startErr
}

func (w *FileWatcher) Stop() {
	w.stopOnce.Do(func() {
		close(w.done)
	})
}
