package assayer

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type HandleResponse struct {
	verdict Verdict
	err     error
}

func (r *HandleResponse) isEmpty() bool {
	return r.verdict == nil && r.err == nil
}

func checkRepository(directory, repository string, verdicts chan<- HandleResponse, args *Arguments) {
	fullPath := filepath.Join(directory, repository)
	if args.Exclude.Match(fullPath) {
		return
	}
	repo, err := git.PlainOpen(fullPath)
	if err != nil {
		verdicts <- HandleResponse{nil, fmt.Errorf("error opening git repository %s\n%s", repository, err)}
		return
	}

	if checkedWorktree := checkWorktree(directory, repository, repo, args); !checkedWorktree.isEmpty() {
		verdicts <- checkedWorktree
		return
	}

	if checkedStash := checkStash(repository, repo, args); !checkedStash.isEmpty() {
		verdicts <- checkedStash
		return
	}

	if checkedBranches := checkBranches(repository, repo, args); !checkedBranches.isEmpty() {
		verdicts <- checkedBranches
		return
	}

	if args.Unmodified {
		verdicts <- HandleResponse{NewUnmodified(repository), nil}
		return
	}
}

func checkRemoteBranches(repository string,
	repo *git.Repository,
	references storer.ReferenceIter,
	branchHashes map[string]plumbing.Hash,
	args *Arguments,
) HandleResponse {
	defer references.Close()
	for {
		ref, err := references.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return HandleResponse{nil, err}
		}

		if !ref.Name().IsRemote() {
			continue
		}

		onlyBranchName, err := extractBranchName(repository, ref)
		if err != nil {
			return HandleResponse{nil, err}
		}

		localHash, hasLocalClone := branchHashes[onlyBranchName]
		delete(branchHashes, onlyBranchName)

		remoteHash := ref.Hash()
		if hasLocalClone && remoteHash != localHash {
			remoteBranchCommit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return HandleResponse{nil, err}
			}
			localBranchCommit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return HandleResponse{nil, err}
			}

			isRemoteAncestor, err := remoteBranchCommit.IsAncestor(localBranchCommit)
			if err != nil {
				return HandleResponse{nil, err}
			}

			if isRemoteAncestor {
				if args.RemoteBehind {
					return HandleResponse{RemoteBehind{
						repository:    repository,
						localBranch:   onlyBranchName,
						remoteRefName: ref.Name().Short(),
					}, nil}
				}
			} else {
				if args.RemoteAhead {
					return HandleResponse{RemoteAhead{
						repository:    repository,
						localBranch:   onlyBranchName,
						remoteRefName: ref.Name().Short(),
					}, nil}
				}
			}
		}
	}
	return HandleResponse{}
}

func checkBranches(repository string, repo *git.Repository, args *Arguments) HandleResponse {
	if !args.LocalOnlyBranch && !args.RemoteAhead && !args.RemoteBehind {
		return HandleResponse{}
	}

	branches, err := repo.Branches()
	if err != nil {
		return HandleResponse{nil, fmt.Errorf("cannot get branches for %s\n%s", repository, err)}
	}

	var branchHashes = make(map[string]plumbing.Hash)
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		branchHashes[branch.Name().Short()] = branch.Hash()
		return nil
	})
	if err != nil {
		return HandleResponse{nil, fmt.Errorf("cannot get branches for %s\n%s", repository, err)}
	}

	references, err := repo.References()
	if err != nil {
		return HandleResponse{nil, fmt.Errorf("cannot get references for %s\n%s", repository, err)}
	}

	response := checkRemoteBranches(repository, repo, references, branchHashes, args)
	if !response.isEmpty() {
		return response
	}

	if args.LocalOnlyBranch {
		for branch := range branchHashes {
			return HandleResponse{LocalOnlyBranch{
				repository: repository,
				branchName: branch,
			}, nil}
		}
	}

	return HandleResponse{}
}

func checkModified(repository string, status git.Status) HandleResponse {
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

	if len(modifiedItem) == 0 {
		return HandleResponse{}
	}
	return HandleResponse{Modified{
		repository:       repository,
		modifiedItem:     modifiedItem,
		modificationType: modificationType,
	}, nil}
}

func checkStash(repository string, repo *git.Repository, args *Arguments) HandleResponse {
	if !args.StashedChanges {
		return HandleResponse{}
	}
	references, err := repo.References()
	if err != nil {
		return HandleResponse{nil, err}
	}
	defer references.Close()
	for {
		ref, err := references.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return HandleResponse{nil, err}
		}

		if ref.Name() == "refs/stash" {
			commit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return HandleResponse{nil, err}
			}
			firstParent, err := commit.Parent(0)
			if err != nil {
				return HandleResponse{nil, err}
			}
			return HandleResponse{StashedChanges{
				repository:       repository,
				commitUnderStash: firstParent,
			}, nil}
		}
	}
	return HandleResponse{}
}

func checkUntracked(directory string, repository string, status git.Status) HandleResponse {
	var untrackedItem string
	for path, s := range status {
		if s.Worktree == git.Untracked {
			untrackedItem = path
			break
		}
	}
	if len(untrackedItem) == 0 {
		return HandleResponse{}
	}

	untrackedPath := splitPath(untrackedItem)

	fullRepository, err := filepath.Abs(filepath.Join(directory, repository))
	if err != nil {
		return HandleResponse{nil, err}
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
		return HandleResponse{nil, err}
	}

	untrackedItem = filepath.Join(untrackedPath[0:(maxMatch + 1)]...)

	return HandleResponse{Untracked{
		repository:    repository,
		untrackedItem: untrackedItem,
	}, nil}
}

func checkWorktree(directory, repository string, repo *git.Repository, args *Arguments) HandleResponse {
	if !args.Modified && !args.Untracked {
		return HandleResponse{}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return HandleResponse{nil, fmt.Errorf("error checking repository worktree %s\n%s", repository, err)}
	}
	status, err := tree.Status()
	if err != nil {
		return HandleResponse{nil, fmt.Errorf("error checking repository status %s\n%s", repository, err)}
	}

	if args.Modified {
		if res := checkModified(repository, status); !res.isEmpty() {
			return res
		}
		return HandleResponse{}
	}
	return checkUntracked(directory, repository, status)
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, string(os.PathSeparator))
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
