package main

import (
	"fmt"
	"github.com/gaols/easyssh"
)

func main() {
	// Create SSHConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.SSHConfig{
		User:     "gaols",
		Server:   "192.168.2.100",
		Password: "******",
		Port:     "22",
	}

	// Call Scp method with file you want to upload to remote server.
	err := ssh.Scp("/dirpath/to/copy", "/tmp/")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("success")

	}
}
