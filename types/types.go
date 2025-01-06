package types

import (
	"github.com/go-git/go-git/v5"
)

type Verdict interface {
	Repository() string
}

type Unmodified struct {
	repository string
}

func NewUnmodified(repository string) Unmodified {
	return Unmodified{repository: repository}
}

func (u Unmodified) Repository() string {
	return u.repository
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
