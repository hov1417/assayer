package assayer

import "github.com/gobwas/glob"

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
	Exclude glob.Glob
	Deep    bool
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

		Nested: false,
	}
}
