package main

import (
	"assayer/assayer"
	"path/filepath"
	"testing"
)

func TestProject(t *testing.T) {
	path, err := filepath.Abs("~/Projects/")
	if err != nil {
		t.Error(err)
		return
	}
	err = assayer.TraverseDirectories(path, false)
	if err != nil {
		t.Error(err)
		return
	}
}
