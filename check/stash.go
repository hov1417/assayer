package check

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/types"
	"io"
	"iter"
)

type StashChecker struct {
}

func NewStashChecker(arguments arguments.Arguments) *StashChecker {
	if !arguments.StashedChanges {
		return nil
	}

	return &StashChecker{}
}

func (s *StashChecker) Check(
	directory, repository string,
	repo *git.Repository,
) iter.Seq[types.Response] {
	return func(yield func(types.Response) bool) {
		references, err := repo.References()
		if err != nil {
			yield(types.Response{Err: err})
			return
		}
		defer references.Close()
		for {
			ref, err := references.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				yield(types.Response{Err: err})
				return
			}

			if ref.Name() == "refs/stash" {
				commit, err := repo.CommitObject(ref.Hash())
				if err != nil {
					yield(types.Response{Err: err})
					return
				}
				firstParent, err := commit.Parent(0)
				if err != nil {
					yield(types.Response{Err: err})
					return
				}
				yield(types.Response{Verdict: StashedChanges{
					repository:       repository,
					commitUnderStash: firstParent,
				}})
			}
		}
	}
}

func (s *StashChecker) ToString() string {
	return "StashChecker"
}

type StashedChanges struct {
	repository       string
	commitUnderStash *object.Commit
}

func (u StashedChanges) Repository() string {
	return u.repository
}

func (u StashedChanges) CommitUnderStash() *object.Commit {
	return u.commitUnderStash
}
