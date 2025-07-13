package assayer

import (
	"fmt"
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/check"
	"github.com/hov1417/assayer/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Directory struct {
	readDir []fs.DirEntry
	path    string
}

func TraverseDirectories(directory string, args arguments.Arguments) error {
	repositories, err := findRepositories(directory, args.Nested)
	if err != nil {
		return fmt.Errorf("error finding repositories\n%s", err)
	}

	verdicts, err := checkRepositories(directory, repositories, args)
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
	directory string,
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
		go func(repository string) {
			assayer.CheckRepository(directory, repository, verdicts, &args)
			wg.Done()
		}(*repositoryRecord.repository)
	}
	go func() {
		wg.Wait()
		close(verdicts)
	}()
	return verdicts, nil
}

type RepositoryRecord struct {
	repository *string
	err        error
}

func findRepositories(directory string, nestedRepos bool) (chan RepositoryRecord, error) {
	dirFs := os.DirFS(directory)

	readDir, err := fs.ReadDir(dirFs, ".")
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	var repositories = make(chan RepositoryRecord, 100)
	entry := Directory{
		readDir,
		".",
	}
	wg.Add(1)
	go handleDirEntry(dirFs, entry, &wg, repositories, nestedRepos)

	go func() {
		wg.Wait()
		close(repositories)
	}()

	return repositories, nil
}

func handleDirEntry(
	dirFs fs.FS,
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
				repositories <- RepositoryRecord{&repository, nil}
			}

			if !stop {
				readDir, err := fs.ReadDir(dirFs, path)
				if err != nil {
					repositories <- RepositoryRecord{nil, err}
				}

				dirEntry := Directory{
					readDir: readDir,
					path:    path,
				}
				wg.Add(1)
				go handleDirEntry(dirFs, dirEntry, wg, repositories, nestedRepos)
			}
		}
	}
	wg.Done()
}
