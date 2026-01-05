package assayer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/check"
	"github.com/hov1417/assayer/types"
)

type Directory struct {
	readDir []fs.DirEntry
	path    string
}

func TraverseDirectories(directories []string, args arguments.Arguments) error {

	var repositories = make(chan RepositoryRecord, 100)
	wg := sync.WaitGroup{}

	for _, dir := range directories {
		err := findRepositories(dir, repositories, &wg, args.Nested)
		if err != nil {
			return fmt.Errorf("error finding repositories\n%s", err)
		}
	}
	go func() {
		wg.Wait()
		close(repositories)
	}()

	verdicts, err := checkRepositories(repositories, args)
	if err != nil {
		return fmt.Errorf("error checking repositories\n%s", err)
	}

	if args.Count {
		err = ReportResultByCount(verdicts, args)
	} else if args.Reporter != nil {
		err = ReportResultWithReporter(verdicts, args)
	} else {
		err = ReportResults(verdicts, args)
	}

	if err != nil {
		return err
	}

	return nil
}

func checkRepositories(
	repositories chan RepositoryRecord,
	args arguments.Arguments,
) (chan types.Response, error) {
	verdicts := make(chan types.Response, 100)
	assayer := check.NewAssayer(args)

	var wg sync.WaitGroup
	for repositoryRecord := range repositories {
		if repositoryRecord.err != nil {
			return nil, repositoryRecord.err
		}
		wg.Add(1)
		go func(repository, rootDirectory string) {
			assayer.CheckRepository(rootDirectory, repository, verdicts, &args)
			wg.Done()
		}(*repositoryRecord.repository, *repositoryRecord.rootDirectory)
	}
	go func() {
		wg.Wait()
		close(verdicts)
	}()
	return verdicts, nil
}

type RepositoryRecord struct {
	repository    *string
	rootDirectory *string
	err           error
}

func findRepositories(directory string, repositories chan RepositoryRecord, wg *sync.WaitGroup, nestedRepos bool) error {
	dirFs := os.DirFS(directory)

	readDir, err := fs.ReadDir(dirFs, ".")
	if err != nil {
		return err
	}
	entry := Directory{
		readDir,
		".",
	}
	wg.Add(1)
	go handleDirEntry(dirFs, directory, entry, wg, repositories, nestedRepos)

	return nil
}

func handleDirEntry(
	dirFs fs.FS,
	rootDirectory string,
	directory Directory,
	wg *sync.WaitGroup,
	repositories chan RepositoryRecord,
	nestedRepos bool,
) {
	stop := false
	if !nestedRepos {
		for _, entry := range directory.readDir {
			if entry.IsDir() && entry.Name() == ".git" {
				stop = true
			}
		}
	}

	for _, entry := range directory.readDir {
		if entry.IsDir() {
			path := filepath.Join(directory.path, entry.Name())

			if strings.HasSuffix(path, ".git") {
				repository := filepath.Dir(path)
				repositories <- RepositoryRecord{&repository, &rootDirectory, nil}
			}

			if !stop {
				readDir, err := fs.ReadDir(dirFs, path)
				if err != nil {
					repositories <- RepositoryRecord{nil, nil, err}
				}

				dirEntry := Directory{
					readDir: readDir,
					path:    path,
				}
				wg.Add(1)
				go handleDirEntry(dirFs, rootDirectory, dirEntry, wg, repositories, nestedRepos)
			}
		}
	}
	wg.Done()
}
