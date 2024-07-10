package src

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Verdict interface {
	Repository() string
}

type Unmodified struct {
	repository string
}

func (u Unmodified) Repository() string {
	return u.repository
}

func NewUnmodified(repository string) Unmodified {
	return Unmodified{
		repository: repository,
	}
}

type Untracked struct {
	repository    string
	untrackedItem string
}

func (u Untracked) Repository() string {
	return u.repository
}

type Modified struct {
	repository       string
	modifiedItem     string
	modificationType git.StatusCode
}

func (u Modified) Repository() string {
	return u.repository
}

type RemoteMismatch struct {
	repository    string
	localBranch   string
	remoteRefName string
	remoteBehind  bool
}

func (u RemoteMismatch) Repository() string {
	return u.repository
}

type LocalOnlyBranch struct {
	repository string
	branchName string
}

func (u LocalOnlyBranch) Repository() string {
	return u.repository
}

type StashedChanges struct {
	repository       string
	commitUnderStash *object.Commit
}

func (u StashedChanges) Repository() string {
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
