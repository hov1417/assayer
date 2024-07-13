package command_line

import (
	"fmt"
	"github.com/hov1417/assayer/assayer"
	"github.com/urfave/cli/v2"
	"os"
)

func App(action func(c *cli.Context) error) *cli.App {
	return &cli.App{
		Name: "Assayer", // Argonaut or Prospector
		Usage: "List repositories with uncompleted work\n\n" +
			"If none of the Check Type are provided and also `--all` flag is not provided, " +
			"everything would be checked and reported except unmodified repositories.\n" +
			"If some of the Check Type are provided then everything else would not be checked.",
		HideHelp:               false,
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		UsageText:              "assayer [options] [path-to-check]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Usage: "Check all in repositories", Aliases: []string{"a"}},

			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "unmodified",
				Usage:    "Show repositories where nothing is changed",
				Aliases:  []string{"u"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "modified",
				Usage:    "Check if worktree is changed",
				Aliases:  []string{"m"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "untracked",
				Usage:    "Check if there are untracked files",
				Aliases:  []string{"t"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "stashed",
				Usage:    "Check if there are stashed changes",
				Aliases:  []string{"s"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "behind-branches",
				Usage:    "Check if there are branches that are behind remote",
				Aliases:  []string{"b"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "ahead-branches",
				Usage:    "Check if there are branches that are ahead remote",
				Aliases:  []string{"A"},
			},
			&cli.BoolFlag{
				Category: "Check Type",
				Name:     "local-only-branches",
				Usage:    "Check if there are local only branches",
				Aliases:  []string{"l"},
			},

			&cli.BoolFlag{
				Name:    "nested",
				Usage:   "Check repositories in repositories",
				Aliases: []string{"n"},
			},
		},
		CommandNotFound: func(c *cli.Context, command string) {
			println("Command " + command + " not found")
			cli.ShowAppHelpAndExit(c, 2)
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			cli.ShowAppHelpAndExit(c, 1)
			return err
		},
		Action: action,
	}

}

func ParseFlags(c *cli.Context) (assayer.Arguments, error) {
	if noTypeFlagsAreSet(c) {
		return assayer.DefaultArguments(), nil
	}
	if c.Bool("all") {
		if anyTypeFlagsIsSet(c) {
			return assayer.Arguments{},
				fmt.Errorf("flag `--all` and Check Type flags should not be given simultaneously")
		}
		return assayer.Arguments{
			Unmodified:      true,
			Modified:        true,
			Untracked:       true,
			StashedChanges:  true,
			RemoteBehind:    true,
			RemoteAhead:     true,
			LocalOnlyBranch: true,

			Nested: c.Bool("nested"),
		}, nil
	}

	return assayer.Arguments{
		Unmodified:      c.Bool("unmodified"),
		Modified:        c.Bool("modified"),
		Untracked:       c.Bool("untracked"),
		StashedChanges:  c.Bool("stashed"),
		RemoteBehind:    c.Bool("behind-branches"),
		RemoteAhead:     c.Bool("ahead-branches"),
		LocalOnlyBranch: c.Bool("local-only-branches"),

		Nested: c.Bool("nested"),
	}, nil
}

func noTypeFlagsAreSet(c *cli.Context) bool {
	return !c.IsSet("unmodified") &&
		!c.IsSet("modified") &&
		!c.IsSet("untracked") &&
		!c.IsSet("stashed") &&
		!c.IsSet("behind-branches") &&
		!c.IsSet("ahead-branches") &&
		!c.IsSet("local-only-branches")
}

func anyTypeFlagsIsSet(c *cli.Context) bool {
	return !noTypeFlagsAreSet(c)
}

func RootDirectory(c *cli.Context) (string, error) {
	if c.NArg() > 1 {
		return "", fmt.Errorf("remove unknown argument(s) %s", c.Args().Slice()[1:])
	}
	var workingDirectory string
	if c.NArg() == 1 {
		workingDirectory = c.Args().First()
	} else {
		var err error
		workingDirectory, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot get current working directory")
		}
	}
	return workingDirectory, nil
}
