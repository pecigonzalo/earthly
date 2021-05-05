package cliutil

import (
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/earthly/earthly/util/fileutil"
	"github.com/pkg/errors"
)

var earthlyDir string
var earthlyDirErr error
var earthlyDirOnce sync.Once

// GetEarthlyDir returns the .earthly dir. (Usually ~/.earthly).
func GetEarthlyDir() (string, error) {
	earthlyDirOnce.Do(func() {
		earthlyDir, earthlyDirErr = makeEarthlyDir()
	})

	return earthlyDir, earthlyDirErr
}

func EnsurePermissions() error {
	if earthlyDir == "" {
		_, err := GetEarthlyDir()
		return err
	}

	u, err := currentNonSudoUser()
	if err != nil {
		return errors.Wrap(err, "chown generated certificates")
	}

	fileutil.EnsureUserOwned(earthlyDir, u)
	return nil
}

func makeEarthlyDir() (string, error) {
	homeDir, sudoUser, err := detectHomeDir()
	if err != nil {
		return "", err
	}
	earthlyDir := filepath.Join(homeDir, ".earthly")
	if !fileutil.DirExists(earthlyDir) {
		err := os.MkdirAll(earthlyDir, 0755)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create dir %s", earthlyDir)
		}
		fileutil.EnsureUserOwned(earthlyDir, sudoUser)

		os.UserHomeDir()
	}
	return earthlyDir, nil
}

func detectHomeDir() (homeDir string, sudoUser *user.User, err error) {
	u, err := currentNonSudoUser()
	if err != nil {
		return "", nil, errors.Wrap(err, "lookup user for homedir")
	}

	if u.HomeDir == "" {
		return "/etc", u, nil
	}

	return u.HomeDir, u, nil
}

func currentNonSudoUser() (*user.User, error) {
	if sudoUserName, ok := os.LookupEnv("SUDO_USER"); ok {
		sudoUser, err := user.Lookup(sudoUserName)
		if err == nil {
			return sudoUser, nil
		}
	}

	return user.Current()
}
