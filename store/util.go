package store

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
)

func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0775)
	}
	return nil
}

func DoesFileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func downloadCrSqlite(installPath string, filename string) error {
	err := EnsureDirExists(installPath)
	if err != nil {
		return err
	}
	platformPair := ""
	cpu := ""
	switch runtime.GOARCH {
	case "arm64":
		cpu = "aarch64"
	case "amd64":
		cpu = "x86_64"
	case "386":
		cpu = "i686"
	default:
		return fmt.Errorf("unsupported cpu architecture")
	}
	platformPair = runtime.GOOS + "-" + cpu
	if runtime.GOOS == "android" {
		platformPair = "aarch64-linux-android"
	}

	resp, err := http.Get(fmt.Sprintf("https://github.com/vlcn-io/cr-sqlite/releases/download/%s/crsqlite-%s.zip", CRSQLITE_VERSION, platformPair))
	if err != nil {
		return errors.Join(errors.New("unable to download crsqlite version: "+CRSQLITE_VERSION+" on platform: "+platformPair), err)
	}
	defer resp.Body.Close()

	f, err := os.Create(path.Join(installPath, filename+".zip"))
	if err != nil {
		return errors.Join(errors.New("unable to make dynamic lib file"), err)
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	f.Close()

	zr, err := zip.OpenReader(path.Join(installPath, filename+".zip"))
	if err != nil {
		return errors.Join(errors.New("unable to open zip"), err)
	}
	zipFile := zr.File[0]
	fileContents, err := zipFile.Open()
	if err != nil {
		return errors.Join(errors.New("unable to access zip file contents"), err)
	}
	defer fileContents.Close()
	osf, err := os.Create(path.Join(installPath, filename))

	_, err = io.Copy(osf, fileContents)
	if err != nil {
		return err
	}
	osf.Close()

	err = os.Chmod(path.Join(installPath, filename), 0755)
	if err != nil {
		return err
	}

	os.Remove(path.Join(installPath, filename+".zip"))

	return err
}

func justPath(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	return u.Path
}
