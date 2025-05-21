package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Must pass os.O_CREATE or O_RDONLY
func GetCleanPath(loc string, flag int) (*os.File, error) {
	// Remove project dir if it already has it
	var cleanPath string
	if strings.HasPrefix(loc, E_PROJECT_DIR+"/") && len(loc) > (len(E_PROJECT_DIR)+1) {
		cleanPath = loc[len(E_PROJECT_DIR)+1:]
	} else {
		cleanPath = loc
	}

	// Clean the path
	cleanPath = filepath.Clean(cleanPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: path traversal attempt detected, %s", loc)
	}

	// Verify project dir
	cleanProjectDir := filepath.Clean(E_PROJECT_DIR)
	if strings.Contains(cleanProjectDir, "..") {
		return nil, fmt.Errorf("invalid file path: project path traversal attempt detected, %s", loc)
	}

	// Join and final check
	cleanPath = filepath.Join(cleanProjectDir, cleanPath)
	if !strings.HasPrefix(filepath.Clean(cleanPath), filepath.Clean(cleanProjectDir)) {
		return nil, fmt.Errorf("invalid file path: path is outside of project directory")
	}

	if flag == os.O_CREATE {
		file, err := os.Create(cleanPath)
		if err != nil {
			return nil, ErrCheck(err)
		}
		return file, nil
	} else if flag == os.O_RDONLY {
		file, err := os.Open(cleanPath)
		if err != nil {
			return nil, ErrCheck(err)
		}
		return file, nil
	}
	return nil, nil
}
