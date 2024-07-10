package src

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Directory struct {
	readDir []fs.DirEntry
	path    string
}

type HandleResponse struct {
	verdict Verdict
	err     error
}

func TraverseDirectories(directory string) error {
	dirFs := os.DirFS(directory)

	readDir, err := fs.ReadDir(dirFs, ".")
	if err != nil {
		return err
	}

	var activeTasks = int32(1)

	var repositories = make(chan string, 10)
	entry := Directory{
		readDir,
		".",
	}
	go handleDirEntry(dirFs, entry, &activeTasks, repositories)

	verdicts := make(chan HandleResponse, 10)

	var wg sync.WaitGroup
	for repository := range repositories {
		go func(repository string) {
			wg.Add(1)
			handleRepository(directory, repository, verdicts)
			wg.Done()
		}(repository)
	}
	go func() {
		wg.Wait()
		close(verdicts)
	}()

	for verdict := range verdicts {
		if verdict.err != nil {
			return verdict.err
		}
		switch verdict := verdict.verdict.(type) {
		case Unmodified:
			fmt.Printf("%s: Unmodified\n", verdict.Repository())
		case Untracked:
			fmt.Printf("%s: Path \"%s\" is untracked\n", verdict.Repository(), verdict.untrackedItem)
		case Modified:
			fmt.Printf("%s: File \"%s\" is %s\n", verdict.Repository(), verdict.modifiedItem, Stringify(verdict.modificationType))
		case RemoteMismatch:
			if verdict.remoteBehind {
				fmt.Printf("%s: Remote Mismatch, remote branch \"%s\" is behind\n ", verdict.Repository(), verdict.localBranch)
			} else {
				fmt.Printf("%s: Remote Mismatch, remote branch \"%s\" is ahead\n", verdict.Repository(), verdict.localBranch)
			}
		case LocalOnlyBranch:
			fmt.Printf("%s: Local Only Branch \"%s\"\n", verdict.Repository(), verdict.branchName)
		case StashedChanges:
			fmt.Printf("%s: Stashed Changes on commit \"%s\"\n", verdict.Repository(), FirstLine(verdict.commitUnderStash.Message))
		}
	}

	return nil
}

func FirstLine(message string) string {
	newline := strings.IndexFunc(message, func(char rune) bool {
		return char == '\n' || char == '\r'
	})
	if newline == -1 {
		return message
	}
	return message[:newline]
}

func handleRepository(directory, repository string, verdicts chan<- HandleResponse) {
	status, err := checkStatus(directory, repository)
	verdicts <- HandleResponse{status, err}
}

func checkStatus(directory, repository string) (Verdict, error) {
	repo, err := git.PlainOpen(filepath.Join(directory, repository))
	if err != nil {
		return nil, fmt.Errorf("error opening git repository %s\n%s", repository, err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error checking repository worktree %s\n%s", repository, err)
	}
	status, err := tree.Status()
	if err != nil {
		return nil, fmt.Errorf("error checking repository status %s\n%s", repository, err)
	}

	verdict, err := checkStash(repo, repository)
	if err != nil {
		return nil, err
	}
	if verdict != nil {
		return verdict, nil
	}

	untracked := checkUntracked(repository, status)
	if untracked != nil {
		return untracked, nil
	}
	modified := checkModified(repository, status)
	if modified != nil {
		return modified, nil
	}

	remoteMismatch, err := checkBranches(repository, repo)
	if remoteMismatch != nil {
		return remoteMismatch, nil
	}
	if err != nil {
		return nil, err
	}

	return NewUnmodified(repository), nil
}

func checkStash(repo *git.Repository, repository string) (Verdict, error) {
	references, err := repo.References()
	if err != nil {
		return nil, err
	}
	defer references.Close()
	for {
		ref, err := references.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if ref.Name() == "refs/stash" {
			commit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return nil, err
			}
			firstParent, err := commit.Parent(0)
			if err != nil {
				return nil, err
			}
			return StashedChanges{
				repository:       repository,
				commitUnderStash: firstParent,
			}, nil
		}
	}
	return nil, nil
}

func checkUntracked(repository string, status git.Status) Verdict {
	var untrackedItem string
	for path, s := range status {
		if s.Worktree == git.Untracked {
			untrackedItem = path
			break
		}
	}
	if len(untrackedItem) == 0 {
		return nil
	}

	untrackedPath := splitPath(untrackedItem)

	maxMatch := 0
	for path, s := range status {
		if s.Worktree != git.Untracked || s.Staging != git.Untracked {
			trackedPath := splitPath(path)
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
				}
			}
			if len(untrackedPath) == maxMatch {
				break
			}
		}
	}

	untrackedItem = filepath.Join(untrackedPath[0:(maxMatch + 1)]...)

	return Untracked{
		repository:    repository,
		untrackedItem: untrackedItem,
	}
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, string(os.PathSeparator))
}

func checkModified(repository string, status git.Status) Verdict {
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
		return nil
	}
	return Modified{
		repository:       repository,
		modifiedItem:     modifiedItem,
		modificationType: modificationType,
	}
}

func checkBranches(repository string, repo *git.Repository) (Verdict, error) {
	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("cannot get branches for %s\n%s", repository, err)
	}

	var branchHashes = make(map[string]plumbing.Hash)
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		branchHashes[branch.Name().Short()] = branch.Hash()
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get branches for %s\n%s", repository, err)
	}

	references, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("cannot get references for %s\n%s", repository, err)
	}

	defer references.Close()
	for {
		ref, err := references.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if ref.Name().IsRemote() {
			parts := strings.Split(ref.Name().Short(), "/")
			if len(parts) < 2 {
				return nil, fmt.Errorf("unknown remote ref format \"%s\" in repository %s",
					ref.Name().Short(),
					repository,
				)
			}
			onlyBranchName := strings.Join(parts[1:], "/")

			localHash, hasLocalClone := branchHashes[onlyBranchName]
			delete(branchHashes, onlyBranchName)

			remoteHash := ref.Hash()
			if hasLocalClone && remoteHash != localHash {
				remoteBranchCommit, err := repo.CommitObject(ref.Hash())
				if err != nil {
					return nil, err
				}
				localBranchCommit, err := repo.CommitObject(ref.Hash())
				if err != nil {
					return nil, err
				}

				isRemoteAncestor, err := remoteBranchCommit.IsAncestor(localBranchCommit)
				if err != nil {
					return nil, err
				}

				return RemoteMismatch{
					repository:    repository,
					localBranch:   onlyBranchName,
					remoteRefName: ref.Name().Short(),
					remoteBehind:  isRemoteAncestor,
				}, nil
			}
		}
	}

	for branch := range branchHashes {
		return LocalOnlyBranch{
			repository: repository,
			branchName: branch,
		}, nil
	}

	return nil, nil
}

func handleDirEntry(dirFs fs.FS, directory Directory, activeTasks *int32, repositories chan string) {
	for _, entry := range directory.readDir {
		if entry.IsDir() {
			path := filepath.Join(directory.path, entry.Name())

			handleDirectory(path, repositories)

			readDir, err := fs.ReadDir(dirFs, path)
			if err != nil {
				println(err)
			}

			dirEntry := Directory{
				readDir: readDir,
				path:    path,
			}
			atomic.AddInt32(activeTasks, 1)
			go handleDirEntry(dirFs, dirEntry, activeTasks, repositories)
		}
	}

	if atomic.AddInt32(activeTasks, -1) == 0 {
		close(repositories)
	}

}

func handleDirectory(path string, repositories chan string) {
	if !strings.HasSuffix(path, ".git") {
		return
	}
	repositories <- filepath.Dir(path)
}
