package utils

import (
	"os"
	"path/filepath"
)

// Removes the oldest subdirectories in dir, keeping only `keep` of them.
// Directories are sorted by name assuming they are named with timestamps; the oldest are deleted first.
func PruneOldestSubdirs(dir string, keep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Collect subdirectories only
	// os.ReadDir already returns entries sorted by name, so no sort.Slice needed
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(dir, e.Name()))
		}
	}

	if len(dirs) <= keep {
		return nil
	}

	// Delete the oldest (lexicographically smallest) entries
	for _, d := range dirs[:len(dirs)-keep] {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}

	return nil
}
