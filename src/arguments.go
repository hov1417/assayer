package src

type Arguments struct {
	Unmodified      bool
	Untracked       bool
	Modified        bool
	RemoteBehind    bool
	RemoteAhead     bool
	LocalOnlyBranch bool
	StashedChanges  bool
}

func DefaultArguments() Arguments {
	return Arguments{
		Unmodified:      false,
		Untracked:       true,
		Modified:        true,
		RemoteBehind:    true,
		RemoteAhead:     true,
		LocalOnlyBranch: true,
		StashedChanges:  true,
	}
}
