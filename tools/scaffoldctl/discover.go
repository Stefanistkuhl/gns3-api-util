package main

import (
	"io/fs"
	"path/filepath"
)

func DiscoverYamlFiles(path string) ([]string, error) {
	paths := []string{}

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if ext == ".yaml" || ext == ".yml" {
			paths = append(paths, p)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}
