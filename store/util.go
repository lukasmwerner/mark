package store

import "os"

func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0775)
	}
	return nil
}
