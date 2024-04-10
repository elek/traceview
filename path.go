package main

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

// ProcessPath execute process function either on the file (specified by path), or on all json files (specified by path of a directory).
func ProcessPath(path string, process func(f string) error) error {
	stat, err := os.Stat(path)
	if err != nil {
		return errors.WithStack(err)
	}
	if stat.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "json") {
				err := process(filepath.Join(path, e.Name()))
				if err != nil {
					return err
				}
			}
		}
	} else {
		err := process(path)
		if err != nil {
			return err
		}
	}
	return nil
}
