package main

import (
	"github.com/hov1417/assayer/assayer"
	"path/filepath"
	"testing"
)

func TestProject(t *testing.T) {
	path, err := filepath.Abs("~/Projects/")
	if err != nil {
		t.Error(err)
		return
	}
	err = assayer.TraverseDirectories(path, assayer.DefaultArguments())
	if err != nil {
		t.Error(err)
		return
	}
}
