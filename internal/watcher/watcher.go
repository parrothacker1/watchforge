package watcher

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/parrothacker1/watchforge/internal/files"
	"github.com/parrothacker1/watchforge/internal/logger"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	root    string
	watched map[string]bool
	filter  *files.Filter
}

func New(root string, filter *files.Filter) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ww := &Watcher{
		watcher: w,
		root:    root,
		watched: make(map[string]bool),
		filter:  filter,
	}
	err = ww.walk()
	if err != nil {
		return nil, err
	}
	logger.Log.Info(
		"watcher initialized",
		"root", root,
		"directories", len(ww.watched),
	)
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
		if w.filter != nil && w.filter.Ignore(path) {
			logger.Log.Debug("directory skipped", "path", path)
			return filepath.SkipDir
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
	logger.Log.Debug("directory added", "path", path, "total", len(w.watched))
	return nil
}

func (w *Watcher) Run(events chan string) {
	for {
		select {
		case event := <-w.watcher.Events:
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			if w.filter != nil && w.filter.Ignore(event.Name) {
				logger.Log.Debug("ignored event", "path", event.Name)
				continue
			}
			logger.Log.Debug(
				"fsnotify event",
				"path", event.Name,
				"op", event.Op.String(),
			)
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					err = w.add(event.Name)
					if err != nil {
						logger.Log.Error("failed to watch directory", "path", event.Name, "error", err)
					}
					continue
				}
			}
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				if _, ok := w.watched[event.Name]; ok {
					_ = w.watcher.Remove(event.Name)
					delete(w.watched, event.Name)
					logger.Log.Debug("directory removed", "path", event.Name, "total", len(w.watched))
					continue
				}
			}
			info, err := os.Stat(event.Name)
			if err == nil && info.IsDir() {
				continue
			}
			events <- event.Name
		case err := <-w.watcher.Errors:
			logger.Log.Error("watch error", "error", err)
		}
	}
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}
