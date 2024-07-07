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
	"sync/atomic"
)

type Directory struct {
	readDir []fs.DirEntry
	path    string
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

	errorChan := make(chan error, 10)

	for repository := range repositories {
		go handleRepository(filepath.Join(directory, repository), errorChan)
	}

	for err := range errorChan {
		return err
	}

	return nil
}

func handleRepository(repository string, errorChan chan error) {
	status, err := checkStatus(repository)
	switch verdict := status.(type) {
	case Unmodified:
		// Do Nothing
		//fmt.Println("unmodified", verdict.Repository())
	case Untracked:
		fmt.Printf("%s untracked file %s\n", verdict.Repository(), verdict.untrackedItem)
	case Modified:
		fmt.Printf("%s modified file %s is %s\n", verdict.Repository(), verdict.modifiedItem, Stringify(verdict.modificationType))
	case RemoteMismatch:
		fmt.Printf("%s Remote Mismatch\n", verdict.Repository())
	}

	errorChan <- err
}

func checkStatus(repository string) (Verdict, error) {
	repo, err := git.PlainOpen(repository)
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

	if untracked, ok := NewUntracked(repository, status); ok {
		return untracked, nil
	}
	if modified, ok := NewModified(repository, status); ok {
		return modified, nil
	}

	remoteMismatch, err := checkRemotes(repository, repo)
	if remoteMismatch != nil {
		return remoteMismatch, nil
	}
	if err != nil {
		return nil, err
	}

	return NewUnmodified(repository), nil
}

func checkRemotes(repository string, repo *git.Repository) (Verdict, error) {
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
			if len(parts) != 2 {
				return nil, fmt.Errorf("unknown remote ref format")
			}
			remoteHash, hasLocalClone := branchHashes[parts[1]]
			localHash := ref.Hash()
			if hasLocalClone && localHash != remoteHash {
				// TODO: check parentage
				return NewRemoteMismatch(repository, parts[1], ref.Name().Short()), nil
			}
		}
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
