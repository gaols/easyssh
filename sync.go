package easyssh

import (
	"path/filepath"
	"fmt"
	"time"
	"errors"
)

// SCopy copy localDirPath to the remote dir specified by remoteDirPath,
// Be aware that localDirPath and remoteDirPath should exists before SCopy.
func (config *SSHConfig) SCopy(localDirPath, remoteDirPath string, verbose bool) error {
	localDirPath = RemoveTrailingSlash(localDirPath)
	remoteDirPath = RemoveTrailingSlash(remoteDirPath)

	if !IsFileExists(localDirPath) {
		return errors.New("no such dir: " + localDirPath)
	}

	localDirParentPath := filepath.Dir(localDirPath)
	localDirname := filepath.Base(localDirPath)
	tgzName := fmt.Sprintf("%s_%s.tar.gz", Sha1(fmt.Sprintf("%s_%d", localDirPath, time.Now().UnixNano())), localDirname)
	defer Local("cd %s;rm -f %s", localDirParentPath, tgzName)
	defer config.Run(fmt.Sprintf("cd %s;rm -f %s", remoteDirPath, tgzName), 24*3600)

	_, err := Local("cd %s;tar czf %s %s", localDirParentPath, tgzName, localDirname)
	if err != nil {
		return errors.New(fmt.Sprintf("create tgz pack for (%s) error: %s", localDirPath, err.Error()))
	}

	tgzPath := filepath.Join(localDirParentPath, tgzName)
	if err = config.Scp(tgzPath, filepath.Join(remoteDirPath, tgzName)); err != nil {
		if verbose {
			fmt.Printf("upload %s -> %s error\n", localDirPath, remoteDirPath)
		}
		return err
	}

	if verbose {
		fmt.Printf("upload %s -> %s done\n", localDirPath, remoteDirPath)
	}

	timeout, err := config.RtRun(fmt.Sprintf("cd %s;tar xf %s", remoteDirPath, tgzName), func(i string) {
		if verbose {
			fmt.Println(i)
		}
	}, func(i string) {
		if verbose {
			fmt.Println(i)
		}
	}, 24*3600)

	if err != nil {
		return errors.New("extract tgz error: " + err.Error())
	}

	if timeout {
		return errors.New("SCopy timeout error")
	}

	return nil
}
