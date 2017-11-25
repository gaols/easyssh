package easyssh

import (
	"fmt"
	"os/exec"
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
