package easyssh

import (
	"testing"
	"fmt"
)

func TestSSHConfig_SCopy(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.155",
		Port:     "22",
		Password: "gaolsz",
	}

	config.Scp("/home/gaols/Codes/go/src/github.com/gaols/easyssh/", "/tmp")
}

func TestSSHConfig_SCopyM(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.155",
		Port:     "22",
		Password: "gaolsz",
	}

	pathMappings := map[string]string{
		"/home/gaols/Codes/go/src/github.com/gaols/easyssh/": "/tmp",
		"/home/gaols/Codes/go/src/github.com/gaols/easydeploy/": "/tmp",
	}
	err := config.ScpM(pathMappings)
	if err != nil {
		fmt.Println(err.Error())
	}
}
