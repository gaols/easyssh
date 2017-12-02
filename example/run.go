package main

import (
	"fmt"

	"github.com/gaols/easyssh"
)

func main() {
	// Create SSHConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.SSHConfig{
		User:   "john",
		Server: "example.com",
		// Optional key or Password without either we try to contact your agent SOCKET
		//Password: "password",
		Key:  "/.ssh/id_rsa",
		Port: "22",
	}

	// Call Run method with command you want to run on remote server.
	done, err := ssh.RtRun("ps aufx", func(line string, lineType int) {
		fmt.Println(line)
	}, 60)
	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("don is :", done)
	}

}
