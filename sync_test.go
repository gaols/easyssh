package easyssh

import (
	"testing"
)

func TestSCopy(t *testing.T) {
	config := &SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Port:     "22",
		Password: "******",
	}

	config.SCopy("/home/gaols/Codes/go/src/github.com/gaols/easyssh/", "/tmp", true)
}
