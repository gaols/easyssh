package easyssh

import (
	"fmt"
	"testing"
)

var sshConfig = &SSHConfig{
	User:   "root",
	Server: "192.168.2.24",
	Key:    "/home/gaols/.ssh/id_rsa",
	Port:   "22",
}

func TestStream(t *testing.T) {
	t.Parallel()
	// input command/output string pairs
	testCases := [][]string{
		{`for i in $(seq 1 5); do echo "$i"; done`, "12345"},
		{`echo "test"`, "test"},
	}
	for _, testCase := range testCases {
		outChannel, errChannel, done, err := sshConfig.Stream(testCase[0], 10)
		if err != nil {
			t.Errorf("Stream failed: %s", err)
		}
		stillGoing := true
		stdout := ""
		stderr := ""
		for stillGoing {
			select {
			case <-done:
				stillGoing = false
			case line := <-outChannel:
				stdout += line
			case line := <-errChannel:
				stderr += line
			}
		}
		if stdout != testCase[1] {
			t.Errorf("Output didn't match expected: %s,%s", stdout, stderr)
		}
	}
}

func TestRun(t *testing.T) {
	commands := []string{
		"cd /root/sfdeploy/projects/ERP3.0.2;mvn clean compile",
	}
	for _, cmd := range commands {
		_, err := sshConfig.RtRun(cmd, func(out string, lineType int) {
			fmt.Println(out)
		}, 50)
		if err != nil {
			t.FailNow()
		}
	}
}

func TestSSHConfig_Scp(t *testing.T) {
	// Call Scp method with file you want to upload to remote server.
	err := sshConfig.Scp("/tmp/hello", "/tmp/target.html")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("success")
	}
}

func TestSSHConfig_RunScript(t *testing.T) {
	script := `
	ls -l /tmp
	echo list tmp done
	`

	err := sshConfig.RunScript(script)
	if err != nil {
		t.Error(err)
	}
}
