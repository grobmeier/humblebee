package paths

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	envHomeOverride = "HUMBLEBEE_HOME"
	dbFileName      = "humblebee.db"
)

func DataDir() (string, error) {
	if override := os.Getenv(envHomeOverride); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("cannot determine home directory")
	}
	return filepath.Join(home, ".humblebee"), nil
}

func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbFileName), nil
}

