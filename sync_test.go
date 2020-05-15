package easyssh

import (
	"fmt"
	"testing"
)

func TestSSHConfig_SCopy(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	err := config.Scp("/home/gaols/Codes/go/src/github.com/gaols/easyssh/", "/tmp")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
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
	err := config.ScpM(pathMappings)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
}

func TestSSHConfig_SafeScp(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	err := config.Scp("/home/gaols/Codes/go/src/github.com/gaols/easyssh/easyssh.go", "/home/gaols/Downloads/easyssh.go")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestSSHConfig_DownloadF(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	err := config.DownloadF("/home/gaols/Codes/go/src/github.com/gaols/easyssh/sync_test.go", "/tmp/sync_test.go")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
