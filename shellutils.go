package easyssh

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gaols/goutils"
	"os/exec"
	"path/filepath"
)

// Local run a cmd on local pc.
// https://studygolang.com/articles/4004   <- run shell command and read output line by line
// https://studygolang.com/articles/7767   <- run command without known args
func Local(localCmd string, paras ...interface{}) (out string, err error) {
	localCmd = fmt.Sprintf(localCmd, paras...)
	cmd := exec.Command("/bin/bash", "-c", localCmd)
	ret, err := cmd.CombinedOutput()
	out = string(ret)
	return
}

// RtLocal run a cmd on local machine and show command output in real time.
func RtLocal(localCmd string, lineHandler func(line string, lineType int8), paras ...interface{}) error {
	localCmd = fmt.Sprintf(localCmd, paras...)
	cmd := exec.Command("/bin/bash", "-c", localCmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	ch := make(chan int8)
	go func() {
		defer stdout.Close()
		defer stderr.Close()
		stdoutScanner := bufio.NewScanner(stdout)
		stderrScanner := bufio.NewScanner(stderr)
		for stdoutScanner.Scan() {
			lineHandler(stdoutScanner.Text(), TypeStdout)
		}
		for stderrScanner.Scan() {
			lineHandler(stderrScanner.Text(), TypeStderr)
		}
		ch <- 1
	}()

	<-ch

	return nil
}

// Tar pack the targetPath and put tarball to tgzPath, targetPath and tgzPath should both the absolute path.
func Tar(tgzPath, targetPath string) error {
	if !goutils.IsDir(targetPath) && !goutils.IsRegular(targetPath) {
		return errors.New("invalid pack path: " + targetPath)
	}

	targetPathDir := filepath.Dir(RemoveTrailingSlash(targetPath))
	target := filepath.Base(RemoveTrailingSlash(targetPath))
	_, err := Local("tar czf %s -C %s %s", tgzPath, targetPathDir, target)
	return err
}

// UnTar unpack the tarball specified by tgzPath and extract it to the path specified by targetPath
func UnTar(tgzPath, targetPath string) error {
	if !goutils.IsDir(targetPath) {
		return errors.New("tar extract path invalid: " + targetPath)
	}

	if !goutils.IsRegular(tgzPath) {
		return errors.New("tar path invalid: " + tgzPath)
	}

	_, err := Local("tar xf %s -C %s", tgzPath, targetPath)
	return err
}
