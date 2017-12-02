package easyssh

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"errors"
)

// https://studygolang.com/articles/4004   <- run shell command and read output line by line
// https://studygolang.com/articles/7767   <- run command without known args
func Local(localCmd string, paras ...interface{}) (out string, err error) {
	localCmd = fmt.Sprintf(localCmd, paras...)
	cmd := exec.Command("/bin/bash", "-c", localCmd)
	ret, err := cmd.Output()
	out = string(ret)
	return
}

// Tar pack the targetPath and put tarball to tgzPath, targetPath and tgzPath should both the absolute path.
func Tar(tgzPath, targetPath string) error {
	if !IsDir(targetPath) && !IsRegular(targetPath) {
		return errors.New("invalid pack path: " + targetPath)
	}

	targetPathDir := filepath.Dir(RemoveTrailingSlash(targetPath))
	target := filepath.Base(RemoveTrailingSlash(targetPath))
	_, err := Local("tar czf %s -C %s %s", tgzPath, targetPathDir, target)
	return err
}

// UnTar unpack the tarball specified by tgzPath and extract it to the path specified by targetPath
func UnTar(tgzPath, targetPath string) error {
	if !IsDir(targetPath) {
		return errors.New("tar extract path invalid: " + targetPath)
	}

	if !IsRegular(tgzPath) {
		return errors.New("tar path invalid: " + tgzPath)
	}

	_, err := Local("tar xf %s -C %s", tgzPath, targetPath)
	return err
}
