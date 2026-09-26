package scanner

import (
	"io/fs"
	"path/filepath"
)

func ScanLogDirectory(dirPath string) ([]string, error) {
	paths := make([]string, 0)
	if err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".log" {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return paths, nil
}
