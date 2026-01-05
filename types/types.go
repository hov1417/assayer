package types

import (
	"path"

	"github.com/go-git/go-git/v5"
)

type Verdict interface {
	Repository() string
	RepositoryPath() string
}

func RepoName(v Verdict, detailed bool) string {
	if detailed {
		return v.RepositoryPath()
	} else {
		return v.Repository()
	}
}

type Unmodified struct {
	base       string
	repository string
}

func NewUnmodified(directory, repository string) Unmodified {
	base := path.Base(directory)
	return Unmodified{base: base, repository: repository}
}

func (u Unmodified) Repository() string {
	return u.repository
}

func (u Unmodified) RepositoryPath() string {
	return path.Join(u.base, u.repository)
}

func Stringify(status git.StatusCode) string {
	switch status {
	case git.Unmodified:
		return "Unmodified"
	case git.Untracked:
		return "Untracked"
	case git.Modified:
		return "Modified"
	case git.Added:
		return "Added"
	case git.Deleted:
		return "Deleted"
	case git.Renamed:
		return "Renamed"
	case git.Copied:
		return "Copied"
	case git.UpdatedButUnmerged:
		return "Updated But Unmerged"
	default:
		return "unknown"
	}

}

type Response struct {
	Verdict Verdict
	Err     error
}
