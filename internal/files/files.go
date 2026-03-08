package files

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

type Filter struct {
	root  string
	git   *gitignore.GitIgnore
	extra *gitignore.GitIgnore
}

func New(root string, extra []string) (*Filter, error) {
	var git *gitignore.GitIgnore
	path := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(path); err == nil {
		ig, err := gitignore.CompileIgnoreFile(path)
		if err != nil {
			return nil, err
		}
		git = ig
	}
	var extraMatcher *gitignore.GitIgnore
	if len(extra) > 0 {
		extraMatcher = gitignore.CompileIgnoreLines(extra...)
	}
	return &Filter{
		root:  root,
		git:   git,
		extra: extraMatcher,
	}, nil
}

func (f *Filter) Ignore(path string) bool {
	rel, err := filepath.Rel(f.root, path)
	if err != nil {
		return false
	}
	if f.git != nil && f.git.MatchesPath(rel) {
		return true
	}
	if f.extra != nil && f.extra.MatchesPath(rel) {
		return true
	}
	return false
}
