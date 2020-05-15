package easyssh

import (
	"errors"
	"fmt"
	"github.com/gaols/goutils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SCopyDir copy localDirPath to the remote dir specified by remoteDirPath,
// Be aware that localDirPath and remoteDirPath should exists before SCopy.
// At last, you should know, timeout is not reliable.
func (sshConf *SSHConfig) SCopyDir(localDirPath, remoteDirPath string, timeout int, verbose bool) error {
	localDirPath = RemoveTrailingSlash(localDirPath)
	remoteDirPath = RemoveTrailingSlash(remoteDirPath)

	if !IsFileExists(localDirPath) {
		return errors.New("no such dir: " + localDirPath)
	}

	localDirParentPath := filepath.Dir(localDirPath)
	localDirname := filepath.Base(localDirPath)
	tgzName := fmt.Sprintf("%s_%s.tar.gz", Sha1(fmt.Sprintf("%s_%d", localDirPath, time.Now().UnixNano())), localDirname)
	defer func() {
		_, _ = Local("cd %s;rm -f %s", localDirParentPath, tgzName)
	}() // safe
	defer func() {
		_, _, _, _ = sshConf.Run(fmt.Sprintf("cd %s;rm -f %s", remoteDirPath, tgzName), timeout)
	}() // safe

	_, err := Local("cd %s;tar czf %s %s", localDirParentPath, tgzName, localDirname)
	if err != nil {
		return fmt.Errorf("create tgz pack for (%s) error: %s", localDirPath, err.Error())
	}

	copyM := fmt.Sprintf("%s -> %s", localDirPath, remoteDirPath)
	tgzPath := filepath.Join(localDirParentPath, tgzName)
	if err = sshConf.SCopyFile(tgzPath, filepath.Join(remoteDirPath, tgzName)); err != nil {
		if verbose {
			fmt.Printf("upload %s error\n", copyM)
		}
		return err
	}

	isTimeout, err := sshConf.RtRun(fmt.Sprintf("cd %s;tar xf %s", remoteDirPath, tgzName), func(line string, lineType int) {
		if verbose && TypeStderr == lineType {
			fmt.Println(line)
		}
	}, timeout)

	if err != nil {
		return errors.New("extract tgz error: " + err.Error())
	}

	if isTimeout {
		return fmt.Errorf("SCopy timeout error: %s", copyM)
	}

	return nil
}

// SCopyFile uploads srcFilePath to remote machine like native scp console app.
// destFilePath should be an absolute file path including filename and cannot be a dir.
func (sshConf *SSHConfig) SCopyFile(srcFilePath, destFilePath string) error {
	return sshConf.Work(func(session *ssh.Session) error {
		src, err := os.Open(srcFilePath)
		if err != nil {
			return err
		}

		stat, err := src.Stat()
		if err != nil {
			return err
		}

		var copyErr error
		go func() {
			stdin, err := session.StdinPipe()
			if err != nil {
				copyErr = err
				return
			}
			defer Close(stdin)
			defer Close(src)
			if _, err = fmt.Fprintf(stdin, "C%#o %d %s\n", stat.Mode().Perm(), stat.Size(), filepath.Base(destFilePath)); err != nil {
				copyErr = fmt.Errorf("copy control char error: %s", err)
			}
			if stat.Size() > 0 {
				if _, err = io.Copy(stdin, src); err != nil {
					copyErr = fmt.Errorf("copy %s error: %s", srcFilePath, err)
					return
				}
			}
			_, copyErr = fmt.Fprint(stdin, "\x00")
		}()

		if copyErr != nil {
			return copyErr
		}

		return session.Run(fmt.Sprintf("scp -t %s", destFilePath))
	})
}

// SCopyM copy multiple local path to their corresponding remote path specified by para pathMappings.
// Warning: to copy a local file, the remote path should contains the filename, however, to copy
// a local dir, the remote path must be a dir into which the local path will be copied.
func (sshConf *SSHConfig) SCopyM(pathMappings map[string]string, timeout int, verbose bool) error {
	errCh := make(chan error, len(pathMappings))
	doneCh := make(chan bool, len(pathMappings))
	var err error
	for localPath, remotePath := range pathMappings {
		go func(local, remote string) {
			if err == nil {
				if err = sshConf.Scp(local, remote); err != nil {
					errCh <- err
				} else {
					doneCh <- true
				}
			}
		}(localPath, remotePath)
	}

	if -1 == timeout {
		// a long timeout simulate wait forever
		timeout = 24 * 3600
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

// Work a helper method to build a ssh connection.
func (sshConf *SSHConfig) Work(fn func(session *ssh.Session) error) error {
	session, err := sshConf.connect()
	if err != nil {
		return err
	}
	defer Close(session)
	return fn(session)
}

// SafeScp first copy localPath to remote /tmp path, then move tmp file to remotePath if upload successfully.
func (sshConf *SSHConfig) SafeScp(localPath, remotePath string) error {
	if goutils.IsDir(localPath) {
		return sshConf.SCopyDir(localPath, remotePath, -1, false)
	}

	remoteTmpName := Sha1(fmt.Sprintf("%s_%d", localPath, time.Now().UnixNano()))
	destTmpPath := filepath.Join("/tmp", remoteTmpName)
	err := sshConf.SCopyFile(localPath, destTmpPath)
	defer func() {
		_, _, _, _ = sshConf.Run(fmt.Sprintf("rm -f /tmp/%s", remoteTmpName), -1) // safe
	}()

	if err != nil {
		return err
	}
	_, _, _, err = sshConf.Run(fmt.Sprintf("mv %s %s", destTmpPath, remotePath), -1) // safe
	return err
}

// DownloadF is short for download file, both the remote path and local path should be the absolute path.
func (sshConf *SSHConfig) DownloadF(remotePath, localPath string) error {
	cli, err := sshConf.Cli()
	if err != nil {
		return err
	}
	defer Close(cli)

	client, err := sftp.NewClient(cli)
	if err != nil {
		return err
	}
	defer Close(client)

	if goutils.IsDir(localPath) {
		return fmt.Errorf("%s is a dir", localPath)
	}

	if goutils.IsRegular(localPath) {
		goutils.Confirm(fmt.Sprintf("%s is already exists, do you want to override it(yY/nN)", localPath), []string{"y", "Y"}, []string{"n", "N"}, func() {
			os.Exit(1)
		})
	}

	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0666); err != nil {
		return fmt.Errorf("mkdir for localpath: %s failed", localPath)
	}

	// create destination file
	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local file error: %s, localPath: %s", err.Error(), localPath)
	}
	defer Close(dstFile)

	// open source file
	srcFile, err := client.Open(remotePath)
	if err != nil {
		return err
	}

	// copy source file to destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		return err
	}
	return nil
}
