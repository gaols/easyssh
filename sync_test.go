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
		Password: "******",
	}

	config.Scp("/home/gaols/Codes/go/src/github.com/gaols/easyssh/", "/tmp")
}

func TestSSHConfig_SCopyM(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.155",
		Port:     "22",
		Password: "******",
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

func TestSSHConfig_SafeScp(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.155",
		Port:     "22",
		Password: "******",
	}

	config.Scp("/home/gaols/Codes/go/src/github.com/gaols/easyssh/easyssh.go", "/home/gaols/Downloads/easyssh.go")
}
