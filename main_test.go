package main

import (
	"github.com/hov1417/assayer/arguments"
	"github.com/hov1417/assayer/assayer"
	"os"
	"testing"
)

func TestProject(t *testing.T) {
	home, _ := os.UserHomeDir()
	path := home + "/Projects/"
	err := assayer.TraverseDirectories(path, arguments.DefaultArguments())
	if err != nil {
		t.Error(err)
		return
	}
}
