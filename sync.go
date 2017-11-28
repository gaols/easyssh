package easyssh

import (
	"path/filepath"
	"fmt"
	"time"
	"errors"
)

// SCopy copy localDirPath to the remote dir specified by remoteDirPath,
// Be aware that localDirPath and remoteDirPath should exists before SCopy.
func (ssh_conf *SSHConfig) SCopy(localDirPath, remoteDirPath string, timeout int, verbose bool) error {
	localDirPath = RemoveTrailingSlash(localDirPath)
	remoteDirPath = RemoveTrailingSlash(remoteDirPath)

	if !IsFileExists(localDirPath) {
		return errors.New("no such dir: " + localDirPath)
	}

	localDirParentPath := filepath.Dir(localDirPath)
	localDirname := filepath.Base(localDirPath)
	tgzName := fmt.Sprintf("%s_%s.tar.gz", Sha1(fmt.Sprintf("%s_%d", localDirPath, time.Now().UnixNano())), localDirname)
	defer Local("cd %s;rm -f %s", localDirParentPath, tgzName)
	defer ssh_conf.Run(fmt.Sprintf("cd %s;rm -f %s", remoteDirPath, tgzName), timeout)

	_, err := Local("cd %s;tar czf %s %s", localDirParentPath, tgzName, localDirname)
	if err != nil {
		return errors.New(fmt.Sprintf("create tgz pack for (%s) error: %s", localDirPath, err.Error()))
	}

	copyM := fmt.Sprintf("%s -> %s", localDirPath, remoteDirPath)
	tgzPath := filepath.Join(localDirParentPath, tgzName)
	if err = ssh_conf.Scp(tgzPath, filepath.Join(remoteDirPath, tgzName)); err != nil {
		if verbose {
			fmt.Printf("upload %s error\n", copyM)
		}
		return err
	}

	if verbose {
		fmt.Printf("upload %s done\n", copyM)
	}

	isTimeout, err := ssh_conf.RtRun(fmt.Sprintf("cd %s;tar xf %s", remoteDirPath, tgzName), func(i string) {
	}, func(errLine string) {
		if verbose {
			fmt.Println(errLine)
		}
	}, timeout)

	if err != nil {
		return errors.New("extract tgz error: " + err.Error())
	}

	if isTimeout {
		return errors.New(fmt.Sprintf("SCopy timeout error: %s", copyM))
	}

	return nil
}
