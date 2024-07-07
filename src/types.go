package src

import "github.com/go-git/go-git/v5"

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

func NewUntracked(repository string, status git.Status) (Untracked, bool) {
	var untrackedItem string
	for path, s := range status {
		if s.Worktree == git.Untracked || s.Staging == git.Untracked {
			untrackedItem = path
			break
		}
	}
	if len(untrackedItem) == 0 {
		return Untracked{}, false
	}
	return Untracked{
		repository:    repository,
		untrackedItem: untrackedItem,
	}, true
}

type Modified struct {
	repository       string
	modifiedItem     string
	modificationType git.StatusCode
}

func NewModified(repository string, status git.Status) (Modified, bool) {
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
		return Modified{}, false
	}
	return Modified{
		repository:       repository,
		modifiedItem:     modifiedItem,
		modificationType: modificationType,
	}, true
}

func (u Modified) Repository() string {
	return u.repository
}

type RemoteMismatch struct {
	repository    string
	localBranch   string
	remoteRefName string
}

func (u RemoteMismatch) Repository() string {
	return u.repository
}

func (u RemoteMismatch) LocalBranch() string {
	return u.localBranch
}

func (u RemoteMismatch) RemoteRefName() string {
	return u.remoteRefName
}

func NewRemoteMismatch(repository, localBranch, remoteRefName string) RemoteMismatch {
	return RemoteMismatch{
		repository:    repository,
		localBranch:   localBranch,
		remoteRefName: remoteRefName,
	}
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
