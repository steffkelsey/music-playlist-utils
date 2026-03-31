package common

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

type WalkResults struct {
	Count               int64
	MapSizeStringSlices map[int64][]string
	Files               []string
	RootPath            string
}

func WalkAllMusicFiles(folder string, processFunc func(path string, info fs.FileInfo, results *WalkResults) error) (WalkResults, error) {
	results := WalkResults{
		Count:               0,
		MapSizeStringSlices: make(map[int64][]string),
		Files:               make([]string, 0),
		RootPath:            folder,
	}

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() && path != folder {
			return nil
		}

		// Skip the root folder itself
		if path == folder {
			return nil
		}

		// If it's an music file, run the processFunc
		if IsMusicFile(path) {
			err = processFunc(path, info, &results)
			if err != nil {
				fmt.Printf("Error on file at in processFunc at path: %s\n, err: %v", path, err)
			}
			// TODO return the error? Does that stop walking?
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through folder: %v\n", err)
		return results, err
	}

	return results, nil
}
