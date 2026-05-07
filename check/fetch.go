package check

import (
	"regexp"

	"github.com/gobwas/glob"
	"github.com/hov1417/assayer/arguments"
)

type FetcherChecker struct {
	FetchType  arguments.FetchType
	FetchGroup *glob.Glob
}

var re = regexp.MustCompile(`[/:]`)

func (f *FetcherChecker) NeedsFetch(remoteUrl string) bool {
	if f.FetchType == arguments.FetchAll {
		return true
	}
	if f.FetchType == arguments.FetchNone {
		return false
	}
	res := re.Split(remoteUrl, -1)
	if len(res) < 2 {
		return false
	}
	group := res[len(res)-2]
	return (*f.FetchGroup).Match(group)
}
