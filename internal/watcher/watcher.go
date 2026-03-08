package watcher

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	root    string
	watched map[string]bool
}

func New(root string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ww := &Watcher{
		watcher: w,
		root:    root,
		watched: make(map[string]bool),
	}
	err = ww.walk()
	if err != nil {
		return nil, err
	}
	return ww, nil
}

func (w *Watcher) walk() error {
	return filepath.WalkDir(w.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		return w.add(path)
	})
}

func (w *Watcher) add(path string) error {
	if w.watched[path] {
		return nil
	}
	err := w.watcher.Add(path)
	if err != nil {
		return err
	}
	w.watched[path] = true
	return nil
}

func (w *Watcher) Run(events chan string) {
	for {
		select {
		case event := <-w.watcher.Events:
			events <- event.Name
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					w.add(event.Name)
				}
			}
		case err := <-w.watcher.Errors:
			log.Println("watch error:", err)
		}
	}
}
