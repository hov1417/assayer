package main

import (
	"fmt"
	"github.com/hov1417/assayer/assayer"
	"github.com/hov1417/assayer/command_line"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := command_line.App(func(c *cli.Context) error {
		workingDirectories, err := command_line.RootDirectories(c)
		if err != nil {
			return err
		}

		arguments, err := command_line.ParseFlags(c)
		if err != nil {
			return err
		}

		err = assayer.TraverseDirectories(workingDirectories, arguments)
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
