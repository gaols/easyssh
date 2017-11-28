package easyssh

import (
	"path/filepath"
	"fmt"
	"time"
	"errors"
)

// SCopy copy localDirPath to the remote dir specified by remoteDirPath,
// Be aware that localDirPath and remoteDirPath should exists before SCopy.
// At last, you should know, timeout is not reliable.
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

	isTimeout, err := ssh_conf.RtRun(fmt.Sprintf("cd %s;tar xf %s", remoteDirPath, tgzName), func(i string) {
	}, func(errLine string) {
		if verbose {
			fmt.Println(errLine)
		}
	}, timeout)

	if verbose {
		fmt.Printf("upload %s done\n", copyM)
	}

	if err != nil {
		return errors.New("extract tgz error: " + err.Error())
	}

	if isTimeout {
		return errors.New(fmt.Sprintf("SCopy timeout error: %s", copyM))
	}

	return nil
}

// SCopyM copy multiple local dir to their corresponding remote dir specified by para pathMappings.
// timeout is not reliable.
func (ssh_conf *SSHConfig) SCopyM(pathMappings map[string]string, timeout int, verbose bool) error {
	errCh := make(chan error, len(pathMappings))
	doneCh := make(chan bool, len(pathMappings))
	var err error
	for localDir, remoteDir := range pathMappings {
		go func(local, remote string) {
			if err == nil {
				if err = ssh_conf.SCopy(local, remote, timeout, verbose); err != nil {
					errCh <- err
				} else {
					doneCh <- true
				}
			}
		}(localDir, remoteDir)
	}

	timeoutChan := time.After(time.Duration(timeout) * time.Second)
L:
	for i := 0; i < len(pathMappings); i++ {
		select {
		case <-doneCh:
		case err = <-errCh:
			break L
		case <-timeoutChan:
			err = errors.New("SCopyM timeout error")
			break L
		}
	}
	return err
}
