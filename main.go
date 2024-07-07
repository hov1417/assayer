package main

import (
	"assayer/src"
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "Assayer" // Argonaut or Prospector
	app.Usage = "List repositories with uncompleted work"
	app.HideHelp = true
	app.Commands = []cli.Command{}
	app.CommandNotFound = func(c *cli.Context, command string) {
		println("Command " + command + " not found")
		cli.ShowAppHelpAndExit(c, 2)
	}
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelpAndExit(c, 1)
		return err
	}
	app.Action = func(c *cli.Context) error {
		if c.NArg() > 1 {
			return fmt.Errorf("remove unknown flags or command")
		}
		var workingDirectory string
		if c.NArg() == 1 {
			workingDirectory = c.Args().First()
		} else {
			var err error
			workingDirectory, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot get current working directory")
			}
		}
		err := src.TraverseDirectories(workingDirectory)
		if err != nil {
			return fmt.Errorf("error while traversing\n%s", err)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
