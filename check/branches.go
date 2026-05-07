package check

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/types"
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

func (b *BranchChecker) Check(
	directory, repository string,
	repo *git.Repository,
) iter.Seq[types.Response] {
	return func(yield func(types.Response) bool) {
		branches, err := repo.Branches()
		if err != nil {
			yield(
				types.Response{Err: fmt.Errorf("cannot get branches for %s\n%s", repository, err)},
			)
			return
		}

		var branchHashes = make(map[string]plumbing.Hash)
		err = branches.ForEach(func(branch *plumbing.Reference) error {
			branchHashes[branch.Name().Short()] = branch.Hash()
			return nil
		})
		if err != nil {
			yield(
				types.Response{Err: fmt.Errorf("cannot get branches for %s\n%s", repository, err)},
			)
			return
		}

		references, err := repo.References()
		if err != nil {
			yield(
				types.Response{
					Err: fmt.Errorf("cannot get references for %s\n%s", repository, err),
				},
			)
			return
		}
		if !b.checkRemoteBranches(references, branchHashes, yield, directory, repository, repo) {
			return
		}

		if b.localOnlyBranch {
			for branch := range branchHashes {
				if !yield(
					types.Response{Verdict: newLocalOnlyBranch(directory, repository, branch)},
				) {
					return
				}
			}
		}
	}
}

func newLocalOnlyBranch(directory, repository string, branch string) LocalOnlyBranch {
	base := path.Base(directory)
	return LocalOnlyBranch{
		base:       base,
		repository: repository,
		branchName: branch,
	}
}

func (b *BranchChecker) ToString() string {
	return "BranchChecker"
}

func (b *BranchChecker) checkRemoteBranches(
	references storer.ReferenceIter,
	branchHashes map[string]plumbing.Hash,
	yield func(types.Response) bool,
	directory, repository string,
	repo *git.Repository,
) bool {
	defer references.Close()
	for {
		ref, err := references.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			yield(
				types.Response{
					Err: fmt.Errorf("%s, error while traversing references: %s", repository, err),
				},
			)
			return false
		}

		if !ref.Name().IsRemote() {
			continue
		}

		onlyBranchName, err := extractBranchName(repository, ref)
		if err != nil {
			yield(
				types.Response{
					Err: fmt.Errorf("%s, error while extracting branch name: %s", repository, err),
				},
			)
			return false
		}

		localHash, hasLocalClone := branchHashes[onlyBranchName]
		delete(branchHashes, onlyBranchName)

		remoteHash := ref.Hash()
		if hasLocalClone && remoteHash != localHash {
			remoteBranchCommit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				yield(types.Response{
					Err: fmt.Errorf(
						"%s, error while getting branch \"%s\" remote commit: %s",
						repository,
						onlyBranchName,
						err,
					),
				})
				return false
			}
			localBranchCommit, err := repo.CommitObject(localHash)
			if err != nil {
				yield(types.Response{Err: fmt.Errorf(
					"%s, error while getting branch \"%s\" local commit: %s",
					repository,
					onlyBranchName,
					err,
				)})
				return false
			}

			isRemoteAncestor, err := remoteBranchCommit.IsAncestor(localBranchCommit)
			if errors.Is(err, plumbing.ErrObjectNotFound) {
				// partial git history, assuming ancestry connection not found
				isRemoteAncestor = false
			} else if err != nil {
				yield(types.Response{Err: fmt.Errorf(
					"%s: error while checking %s and %s ancestory: %s",
					repository,
					remoteBranchCommit.Hash,
					localBranchCommit.Hash,
					err,
				)})
				return false
			}

			if isRemoteAncestor {
				if b.remoteBehind {
					if !yield(
						types.Response{
							Verdict: newRemoteBehind(directory, repository, onlyBranchName, ref),
						},
					) {
						return false
					}
				}
			} else {
				if b.remoteAhead {
					if !yield(
						types.Response{
							Verdict: newRemoteAhead(directory, repository, onlyBranchName, ref),
						},
					) {
						return false
					}
				}
			}
		}
	}
	return true
}

func newRemoteBehind(
	directory, repository string,
	onlyBranchName string,
	ref *plumbing.Reference,
) RemoteBehind {
	base := path.Base(directory)
	return RemoteBehind{
		base:          base,
		repository:    repository,
		localBranch:   onlyBranchName,
		remoteRefName: ref.Name().Short(),
	}
}

func newRemoteAhead(
	directory, repository string,
	onlyBranchName string,
	ref *plumbing.Reference,
) RemoteAhead {
	base := path.Base(directory)
	return RemoteAhead{
		base:          base,
		repository:    repository,
		localBranch:   onlyBranchName,
		remoteRefName: ref.Name().Short(),
	}
}

type RemoteBehind struct {
	base          string
	repository    string
	localBranch   string
	remoteRefName string
}

func (u RemoteBehind) Repository() string {
	return u.repository
}

func (u RemoteBehind) RepositoryPath() string {
	return path.Join(u.base, u.repository)
}

func (u RemoteBehind) LocalBranch() string {
	return u.localBranch
}

func (u RemoteBehind) RemoteRefName() string {
	return u.remoteRefName
}

type RemoteAhead struct {
	base          string
	repository    string
	localBranch   string
	remoteRefName string
}

func (u RemoteAhead) Repository() string {
	return u.repository
}

func (u RemoteAhead) RepositoryPath() string {
	return path.Join(u.base, u.repository)
}

func (u RemoteAhead) LocalBranch() string {
	return u.localBranch
}

func (u RemoteAhead) RemoteRefName() string {
	return u.remoteRefName
}

type LocalOnlyBranch struct {
	base       string
	repository string
	branchName string
}

func (u LocalOnlyBranch) Repository() string {
	return u.repository
}

func (u LocalOnlyBranch) RepositoryPath() string {
	return path.Join(u.base, u.repository)
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
