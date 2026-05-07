package check

import (
	"testing"

	"github.com/gobwas/glob"
	"github.com/hov1417/assayer/arguments"
)

func TestFetcherAll(t *testing.T) {
	fetcher := FetcherChecker{
		FetchType:  arguments.FetchAll,
		FetchGroup: nil,
	}

	if !fetcher.NeedsFetch("git@github.com:go-git/go-git.git") {
		t.Errorf(`Should return true when type is all`)
	}
	if !fetcher.NeedsFetch("https://github.com/go-git/go-git.git") {
		t.Errorf(`Should return true when type is all`)
	}
	if !fetcher.NeedsFetch("https://codeberg.org/ziglang/zig.git") {
		t.Errorf(`Should return true when type is all`)
	}
	if !fetcher.NeedsFetch("git@github.com:hov1417/assayer.git") {
		t.Errorf(`Should return true when type is all`)
	}
	if !fetcher.NeedsFetch("git clone ssh://login@server.com:12345/group/repository.git") {
		t.Errorf(`Should return true when type is all`)
	}
}

func TestFetcherNone(t *testing.T) {
	fetcher := FetcherChecker{
		FetchType:  arguments.FetchNone,
		FetchGroup: nil,
	}

	if fetcher.NeedsFetch("git@github.com:go-git/go-git.git") {
		t.Errorf(`Should return false when type is none`)
	}
	if fetcher.NeedsFetch("https://github.com/go-git/go-git.git") {
		t.Errorf(`Should return false when type is none`)
	}
	if fetcher.NeedsFetch("https://codeberg.org/ziglang/zig.git") {
		t.Errorf(`Should return false when type is none`)
	}
	if fetcher.NeedsFetch("git@github.com:hov1417/assayer.git") {
		t.Errorf(`Should return false when type is none`)
	}
	if fetcher.NeedsFetch("git clone ssh://login@server.com:12345/group/repository.git") {
		t.Errorf(`Should return false when type is none`)
	}
}

func TestFetcherGroup(t *testing.T) {
	group, err := glob.Compile("*")
	if err != nil {
		t.Fatal(err)
		return
	}
	fetcher := FetcherChecker{
		FetchType:  arguments.FetchSome,
		FetchGroup: &group,
	}

	if !fetcher.NeedsFetch("git@github.com:go-git/go-git.git") {
		t.Errorf(`Should return true when type is some but glob is all`)
	}
	if !fetcher.NeedsFetch("https://github.com/go-git/go-git.git") {
		t.Errorf(`Should return true when type is some but glob is all`)
	}
	if !fetcher.NeedsFetch("https://codeberg.org/ziglang/zig.git") {
		t.Errorf(`Should return true when type is some but glob is all`)
	}
	if !fetcher.NeedsFetch("git@github.com:hov1417/assayer.git") {
		t.Errorf(`Should return true when type is some but glob is all`)
	}
	if !fetcher.NeedsFetch("git clone ssh://login@server.com:12345/group/repository.git") {
		t.Errorf(`Should return true when type is some but glob is all`)
	}
}

func TestFetcherGroupSingleGroup(t *testing.T) {
	group, err := glob.Compile("go-git")
	if err != nil {
		t.Fatal(err)
		return
	}
	fetcher := FetcherChecker{
		FetchType:  arguments.FetchSome,
		FetchGroup: &group,
	}

	if !fetcher.NeedsFetch("git@github.com:go-git/go-git.git") {
		t.Errorf(`Should return true when glob is go-git`)
	}
	if !fetcher.NeedsFetch("https://github.com/go-git/go-git.git") {
		t.Errorf(`Should return true when glob is go-git`)
	}
	if fetcher.NeedsFetch("https://codeberg.org/ziglang/zig.git") {
		t.Errorf(`Should return false when glob is go-git`)
	}
	if fetcher.NeedsFetch("git@github.com:hov1417/assayer.git") {
		t.Errorf(`Should return false when glob is go-git`)
	}
	if fetcher.NeedsFetch("git clone ssh://login@server.com:12345/group/repository.git") {
		t.Errorf(`Should return false when glob is go-git`)
	}
}

func TestFetcherGroupMultipleGroups(t *testing.T) {
	group, err := glob.Compile("{go-git,ziglang,group}")
	if err != nil {
		t.Fatal(err)
		return
	}
	fetcher := FetcherChecker{
		FetchType:  arguments.FetchSome,
		FetchGroup: &group,
	}

	if !fetcher.NeedsFetch("git@github.com:go-git/go-git.git") {
		t.Errorf(`Should return true`)
	}
	if !fetcher.NeedsFetch("https://github.com/go-git/go-git.git") {
		t.Errorf(`Should return true`)
	}
	if !fetcher.NeedsFetch("https://codeberg.org/ziglang/zig.git") {
		t.Errorf(`Should return true`)
	}
	if fetcher.NeedsFetch("git@github.com:hov1417/assayer.git") {
		t.Errorf(`Should return false`)
	}
	if !fetcher.NeedsFetch("git clone ssh://login@server.com:12345/group/repository.git") {
		t.Errorf(`Should return true`)
	}
}
