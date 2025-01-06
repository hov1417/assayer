package check

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/types"
	"io"
	"iter"
	"strings"
)

type BranchChecker struct {
	localOnlyBranch bool
	remoteAhead     bool
	remoteBehind    bool
}

func NewBranchChecker(arguments arguments.Arguments) *BranchChecker {
	if !arguments.LocalOnlyBranch && !arguments.RemoteAhead && !arguments.RemoteBehind {
		return nil
	}
	return &BranchChecker{
		localOnlyBranch: arguments.LocalOnlyBranch,
		remoteAhead:     arguments.RemoteAhead,
		remoteBehind:    arguments.RemoteBehind,
	}
}

func (b *BranchChecker) Check(directory, repository string, repo *git.Repository) iter.Seq[types.Response] {
	return func(yield func(types.Response) bool) {
		branches, err := repo.Branches()
		if err != nil {
			yield(types.Response{Err: fmt.Errorf("cannot get branches for %s\n%s", repository, err)})
			return
		}

		var branchHashes = make(map[string]plumbing.Hash)
		err = branches.ForEach(func(branch *plumbing.Reference) error {
			branchHashes[branch.Name().Short()] = branch.Hash()
			return nil
		})
		if err != nil {
			yield(types.Response{Err: fmt.Errorf("cannot get branches for %s\n%s", repository, err)})
			return
		}

		references, err := repo.References()
		if err != nil {
			yield(types.Response{Err: fmt.Errorf("cannot get references for %s\n%s", repository, err)})
			return
		}
		b.checkRemoteBranches(references, branchHashes, yield, repository, repo)

		if b.localOnlyBranch {
			for branch := range branchHashes {
				if !yield(types.Response{Verdict: LocalOnlyBranch{
					repository: repository,
					branchName: branch,
				}}) {
					return
				}
			}
		}
	}
}

func (b *BranchChecker) ToString() string {
	return "BranchChecker"
}

func (b *BranchChecker) checkRemoteBranches(
	references storer.ReferenceIter,
	branchHashes map[string]plumbing.Hash,
	yield func(types.Response) bool,
	repository string,
	repo *git.Repository,
) {
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

		if !ref.Name().IsRemote() {
			continue
		}

		onlyBranchName, err := extractBranchName(repository, ref)
		if err != nil {
			yield(types.Response{Err: err})
			return
		}

		localHash, hasLocalClone := branchHashes[onlyBranchName]
		delete(branchHashes, onlyBranchName)

		remoteHash := ref.Hash()
		if hasLocalClone && remoteHash != localHash {
			remoteBranchCommit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				yield(types.Response{Err: err})
				return
			}
			localBranchCommit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				yield(types.Response{Err: err})
				return
			}

			isRemoteAncestor, err := remoteBranchCommit.IsAncestor(localBranchCommit)
			if err != nil {
				yield(types.Response{Err: err})
				return
			}

			if isRemoteAncestor {
				if b.remoteBehind {
					if !yield(types.Response{Verdict: RemoteBehind{
						repository:    repository,
						localBranch:   onlyBranchName,
						remoteRefName: ref.Name().Short(),
					}}) {
						return
					}
				}
			} else {
				if b.remoteAhead {
					if !yield(types.Response{Verdict: RemoteAhead{
						repository:    repository,
						localBranch:   onlyBranchName,
						remoteRefName: ref.Name().Short(),
					}}) {
						return
					}
				}
			}
		}
	}
}

type RemoteBehind struct {
	repository    string
	localBranch   string
	remoteRefName string
}

func (u RemoteBehind) Repository() string {
	return u.repository
}

func (u RemoteBehind) LocalBranch() string {
	return u.localBranch
}

func (u RemoteBehind) RemoteRefName() string {
	return u.remoteRefName
}

type RemoteAhead struct {
	repository    string
	localBranch   string
	remoteRefName string
}

func (u RemoteAhead) Repository() string {
	return u.repository
}

func (u RemoteAhead) LocalBranch() string {
	return u.localBranch
}

func (u RemoteAhead) RemoteRefName() string {
	return u.remoteRefName
}

type LocalOnlyBranch struct {
	repository string
	branchName string
}

func (u LocalOnlyBranch) Repository() string {
	return u.repository
}

func (u LocalOnlyBranch) BranchName() string {
	return u.branchName
}

func extractBranchName(repository string, ref *plumbing.Reference) (string, error) {
	parts := strings.Split(ref.Name().Short(), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("unknown remote ref format \"%s\" in repository %s",
			ref.Name().Short(),
			repository,
		)
	}
	onlyBranchName := strings.Join(parts[1:], "/")
	return onlyBranchName, nil
}
