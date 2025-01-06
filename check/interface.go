package check

import (
	"github.com/go-git/go-git/v5"
	"github.com/hov1417/assayer/types"
	"iter"
)

type Checker interface {
	Check(directory, repository string, repo *git.Repository) iter.Seq[types.Response]
	ToString() string
}
