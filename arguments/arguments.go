package arguments

import (
	"text/template"

	"github.com/gobwas/glob"
)

type FetchType int

const (
	FetchNone FetchType = iota
	FetchSome
	FetchAll
)

type Arguments struct {
	Unmodified      bool
	Modified        bool
	Untracked       bool
	StashedChanges  bool
	RemoteBehind    bool
	RemoteAhead     bool
	LocalOnlyBranch bool

	Count   bool
	Nested  bool
	Exclude *glob.Glob
	Deep    bool
	Verbose bool

	FetchType  FetchType
	FetchGroup *glob.Glob

	Reporter *template.Template
}

func DefaultArguments() Arguments {
	return Arguments{
		Unmodified:      false,
		Untracked:       true,
		Modified:        true,
		StashedChanges:  true,
		RemoteBehind:    true,
		RemoteAhead:     true,
		LocalOnlyBranch: true,

		FetchType:  FetchNone,
		FetchGroup: nil,

		Nested: false,
	}
}
