package main

import (
	"github.com/gaols/easyssh"
	"fmt"
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
	ssh.RtRun("ps ax", func(stdoutLine string) {
		fmt.Println(stdoutLine)
	}, func(errLine string) {
		fmt.Println(errLine)
	}, 60)

}
