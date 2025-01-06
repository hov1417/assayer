package check

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/types"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

type WorkTreeChecker struct {
	modified  bool
	untracked bool
}

func NewWorkTreeChecker(
	arguments arguments.Arguments,
) *WorkTreeChecker {
	if !arguments.Modified && !arguments.Untracked {
		return nil
	}
	return &WorkTreeChecker{
		modified:  arguments.Modified,
		untracked: arguments.Untracked,
	}
}

type StatusHolder struct {
	status *git.Status
}

func (s *StatusHolder) getStatus(repository string, repo *git.Repository) (*git.Status, error) {
	if s.status == nil {
		tree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("error checking repository worktree %s\n%s", repository, err)
		}
		status, err := tree.Status()
		if err != nil {
			return nil, fmt.Errorf("error checking repository status %s\n%s", repository, err)
		}
		s.status = &status
	}

	return s.status, nil
}

func (w *WorkTreeChecker) Check(directory, repository string, repo *git.Repository) iter.Seq[types.Response] {
	return func(yield func(types.Response) bool) {
		statusHolder := StatusHolder{status: nil}
		if w.modified {
			status, err := statusHolder.getStatus(repository, repo)
			if err != nil {
				yield(types.Response{Err: err})
				return
			}
			if !checkModified(repository, *status, yield) {
				return
			}
		}

		if w.untracked {
			status, err := statusHolder.getStatus(repository, repo)
			if err != nil {
				yield(types.Response{Err: err})
				return
			}
			if !checkUntracked(directory, repository, *status, yield) {
				return
			}
		}
	}
}

func (w *WorkTreeChecker) ToString() string {
	return "WorkTreeCheck"
}

type Modified struct {
	repository       string
	modifiedItem     string
	modificationType git.StatusCode
}

func (u Modified) Repository() string {
	return u.repository
}

func (u Modified) ModifiedItem() string {
	return u.modifiedItem
}

func (u Modified) ModificationType() git.StatusCode {
	return u.modificationType
}

// returns value indicating to "continue" or not
func checkModified(repository string, status git.Status, yield func(types.Response) bool) bool {
	var modifiedItem string
	var modificationType git.StatusCode
	for path, s := range status {
		if s.Worktree != git.Untracked && s.Worktree != git.Unmodified {
			modifiedItem = path
			modificationType = s.Worktree
			break
		}
		if s.Staging != git.Untracked && s.Staging != git.Unmodified {
			modifiedItem = path
			modificationType = s.Staging
			break
		}
	}

	if len(modifiedItem) != 0 {
		return yield(types.Response{Verdict: Modified{
			repository:       repository,
			modifiedItem:     modifiedItem,
			modificationType: modificationType,
		}})
	}
	return true
}

type Untracked struct {
	repository    string
	untrackedItem string
}

func (u Untracked) Repository() string {
	return u.repository
}

func (u Untracked) UntrackedItem() string {
	return u.untrackedItem
}

// returns value indicating to "continue" or not
func checkUntracked(directory string, repository string, status git.Status, yield func(types.Response) bool) bool {
	var untrackedItem string
	for path, s := range status {
		if s.Worktree == git.Untracked {
			untrackedItem = path
			break
		}
	}
	if len(untrackedItem) == 0 {
		return true
	}

	untrackedPath := splitPath(untrackedItem)

	fullRepository, err := filepath.Abs(filepath.Join(directory, repository))
	if err != nil {
		yield(types.Response{Err: err})
		return false
	}

	root := filepath.Join(fullRepository, untrackedPath[0])

	maxMatch := 0
	err = filepath.WalkDir(
		root,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(fullRepository, path)
			if err != nil {
				return err
			}

			if _, ok := status[relPath]; !ok {
				trackedPath := splitPath(relPath)
				var minLen int
				if len(trackedPath) < len(untrackedPath) {
					minLen = len(trackedPath)
				} else {
					minLen = len(untrackedPath)
				}
				for matchIndex := 0; matchIndex < minLen; matchIndex++ {
					if trackedPath[matchIndex] != untrackedPath[matchIndex] {
						if maxMatch < matchIndex {
							maxMatch = matchIndex
						}
						break
					}
				}
				if len(untrackedPath) == maxMatch {
					return filepath.SkipAll
				}
			}
			return nil
		})

	if err != nil {
		yield(types.Response{Err: err})
		return false
	}

	untrackedItem = filepath.Join(untrackedPath[0:(maxMatch + 1)]...)

	return yield(types.Response{Verdict: Untracked{
		repository:    repository,
		untrackedItem: untrackedItem,
	}})
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, string(os.PathSeparator))
}
