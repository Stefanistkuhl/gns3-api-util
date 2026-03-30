package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

type Dirs struct {
	RootDir   string
	TmpDir    string
	ObjectDir string
}

func CreateDirStructure(rootDir string) (Dirs, error) {
	var dirs Dirs
	tmpDir := filepath.Join(rootDir, "tmp")
	objDir := filepath.Join(rootDir, "objects")

	if err := os.MkdirAll(rootDir, 0o750); err != nil {
		return dirs, fmt.Errorf("failed to create root dir: %w", err)
	}

	if err := os.MkdirAll(tmpDir, 0o750); err != nil {
		return dirs, fmt.Errorf("failed to create tmp dir: %w", err)
	}

	if err := os.MkdirAll(objDir, 0o750); err != nil {
		return dirs, fmt.Errorf("failed to create object dir: %w", err)
	}

	dirs.RootDir = rootDir
	dirs.ObjectDir = objDir
	dirs.TmpDir = tmpDir
	return dirs, nil
}

func (d *Dirs) CreateShardDirsIfNeed(hash string) (string, error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("hash too short for sharding: %s", hash)
	}

	targetDir := filepath.Clean(filepath.Join(d.ObjectDir, hash[:2], hash[2:4]))

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create shard structure: %w", err)
	}

	return targetDir, nil
}
