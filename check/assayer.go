package check

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/types"
	"path/filepath"
	"reflect"
)

type Assayer struct {
	checkers []Checker
}

func NewAssayer(arguments arguments.Arguments) Assayer {
	checkers := make([]Checker, 0)

	checkers = append(checkers, NewWorkTreeChecker(arguments))
	checkers = append(checkers, NewStashChecker(arguments))
	checkers = append(checkers, NewBranchChecker(arguments))

	filteredSlice := make([]Checker, 0, len(checkers))
	for _, item := range checkers {
		if !(item == nil || reflect.ValueOf(item).IsNil()) {
			filteredSlice = append(filteredSlice, item)
		}
	}

	return Assayer{
		checkers: filteredSlice,
	}
}

func (a *Assayer) CheckRepository(directory, repository string, verdicts chan<- types.Response, args *arguments.Arguments) {
	fullPath := filepath.Join(directory, repository)
	if args.Exclude != nil && (*args.Exclude).Match(fullPath) {
		return
	}
	repo, err := git.PlainOpen(fullPath)
	if err != nil {
		verdicts <- types.Response{Err: fmt.Errorf("error opening git repository %s\n%s", repository, err)}
		return
	}

	foundVerdict := false
	for _, checker := range a.checkers {
		for v := range checker.Check(directory, repository, repo) {
			verdicts <- v
			if !args.Deep {
				return
			}
			foundVerdict = true
		}
	}

	if !foundVerdict {
		verdicts <- types.Response{Verdict: types.NewUnmodified(repository)}
	}

}
