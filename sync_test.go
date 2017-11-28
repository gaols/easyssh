package easyssh

import (
	"testing"
	"fmt"
)

func TestSSHConfig_SCopy(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	config.SCopy("/home/gaols/Codes/go/src/github.com/gaols/easyssh/", "/tmp", 1, true)
}

func TestSSHConfig_SCopyM(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	pathMappings := map[string]string{
		"/home/gaols/Codes/go/src/github.com/gaols/easyssh/": "/tmp",
		"/home/gaols/Codes/go/src/github.com/gaols/easydeploy/": "/tmp",
	}
	err := config.SCopyM(pathMappings, 5, true)
	if err != nil {
		fmt.Println(err.Error())
	}
}
