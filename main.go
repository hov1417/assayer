package main

import (
	"assayer/assayer"
	"assayer/command_line"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := command_line.App(func(c *cli.Context) error {
		workingDirectory, err := command_line.RootDirectory(c)
		if err != nil {
			return err
		}

		arguments, err := command_line.ParseFlags(c)
		if err != nil {
			return err
		}

		err = assayer.TraverseDirectories(workingDirectory, arguments)
		if err != nil {
			return fmt.Errorf("error while traversing\n%s", err)
		}
		return nil
	})
	err := app.Run(os.Args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
