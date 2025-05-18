package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Must pass os.O_CREATE or O_RDONLY
func GetCleanPath(loc string, flag int) (*os.File, error) {
	cleanLoc := filepath.Clean(loc)
	if strings.Contains(cleanLoc, "..") {
		return nil, fmt.Errorf("invalid file path: path traversal attempt detected, %s", loc)
	}

	projectDir := filepath.Clean(os.Getenv("PROJECT_DIR"))
	if strings.Contains(projectDir, "..") {
		return nil, fmt.Errorf("invalid file path: project path traversal attempt detected, %s", loc)
	}

	cleanPath := filepath.Join(projectDir, cleanLoc)

	if !strings.HasPrefix(filepath.Clean(cleanPath), filepath.Clean(projectDir)) {
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

func GetEnvFile(envFilePath string, byteSize uint16) (string, error) {
	file, err := GetCleanPath(os.Getenv(envFilePath), os.O_RDONLY)
	if err != nil {
		return "", ErrCheck(err)
	}

	fileBytes := make([]byte, byteSize)
	_, err = file.Read(fileBytes)
	if err != nil {
		return "", ErrCheck(err)
	}

	fileStr := string(fileBytes)

	return strings.Trim(fileStr, "\n"), nil
}
