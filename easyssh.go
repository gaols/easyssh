// Package easyssh provides a simple implementation of some SSH protocol
// features in Go. You can simply run a command on a remote server or get a file
// even simpler than native console SSH client. You don't need to think about
// Dials, sessions, defers, or public keys... Let easyssh think about it!
package easyssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"net"
	"os"
	"time"

	"github.com/gaols/goutils"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	// TypeStdout is type of stdout
	TypeStdout = 0
	// TypeStderr is type of stderr
	TypeStderr = 1
)

// SSHConfig contains main authority information.
// User field should be a name of user on remote server (ex. john in ssh john@example.com).
// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
// Key is a path to private key on your local machine.
// Port is SSH server port on remote machine.
type SSHConfig struct {
	User     string
	Server   string
	Key      string
	Port     string
	Password string
	Timeout  int
}

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	pubKey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// connects to remote server using SSHConfig struct and returns *ssh.Session
func (ssh_conf *SSHConfig) connect() (*ssh.Session, error) {
	client, err := ssh_conf.Cli()
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Cli create ssh client
func (ssh_conf *SSHConfig) Cli() (*ssh.Client, error) {
	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if ssh_conf.Password != "" {
		auths = append(auths, ssh.Password(ssh_conf.Password))
	}

	if goutils.IsNotBlank(ssh_conf.Key) {
		if pubKey, err := getKeyFile(ssh_conf.Key); err == nil {
			auths = append(auths, ssh.PublicKeys(pubKey))
		}
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer sshAgent.Close()
	}
	// Default port 22
	ssh_conf.Port = goutils.DefaultIfBlank(ssh_conf.Port, "22")

	// Default current user
	ssh_conf.User = goutils.DefaultIfBlank(ssh_conf.User, os.Getenv("USER"))

	config := &ssh.ClientConfig{
		User:            ssh_conf.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// default maximum amount of time for the TCP connection to establish is 10s
	config.Timeout = time.Duration(time.Second * 10)
	if ssh_conf.Timeout > 0 {
		config.Timeout = time.Duration(ssh_conf.Timeout) * time.Second
	}

	return ssh.Dial("tcp", ssh_conf.Server+":"+ssh_conf.Port, config)
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (ssh_conf *SSHConfig) Stream(command string, timeout int) (stdout, stderr chan string, done chan bool, err error) {
	// connect to remote host
	session, err := ssh_conf.connect()
	if err != nil {
		return stdout, stderr, done, err
	}

	// connect to both outputs (they are of type io.Reader)
	stdOutReader, err := session.StdoutPipe()
	if err != nil {
		return stdout, stderr, done, err
	}
	stderrReader, err := session.StderrPipe()
	if err != nil {
		return stdout, stderr, done, err
	}
	err = session.Start(command)

	// continuously send the command's output over the channel
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	done = make(chan bool)

	go func() {
		defer close(stdoutChan)
		defer close(stderrChan)
		defer close(done)

		go func() {
			stdoutDone := make(chan byte)
			stderrDone := make(chan byte)

			// loop stdout
			go func() {
				stdoutScanner := bufio.NewScanner(stdOutReader)
				for stdoutScanner.Scan() {
					stdoutChan <- stdoutScanner.Text()
				}
				stdoutDone <- 1
			}()

			// loop stderr
			go func() {
				stderrScanner := bufio.NewScanner(stderrReader)
				for stderrScanner.Scan() {
					stderrChan <- stderrScanner.Text()
				}
				stderrDone <- 1
			}()

			<-stdoutDone
			<-stderrDone
			done <- true
		}()

		if timeout <= 0 {
			// a long timeout simulate wait forever
			timeout = 24 * 3600
		}
		timeoutChan := time.After(time.Duration(timeout) * time.Second)
		select {
		case r := <-done:
			done <- r
		case <-timeoutChan:
			stderrChan <- fmt.Sprintf("Run command timeout: %s", command)
			done <- false
		}
	}()

	return stdoutChan, stderrChan, done, err
}

// Run command on remote machine and returns its stdout as a string
func (ssh_conf *SSHConfig) Run(command string, timeout int) (outStr, errStr string, isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, err := ssh_conf.Stream(command, timeout)
	if err != nil {
		return outStr, errStr, isTimeout, err
	}
	// read from the output channel until the done signal is passed
	var stdoutBuf, stderrBuf bytes.Buffer

OuterL:
	for {
		select {
		case done := <-doneChan:
			isTimeout = !done
			break OuterL
		case outLine := <-stdoutChan:
			stdoutBuf.WriteString(outLine + "\n")
		case errLine := <-stderrChan:
			stderrBuf.WriteString(errLine + "\n")
		}
	}
	// return the concatenation of all signals from the output channel
	return stdoutBuf.String(), stderrBuf.String(), isTimeout, err
}

// RtRun run command on remote machine and get command output as soon as possible.
func (ssh_conf *SSHConfig) RtRun(command string, lineHandler func(string string, lineType int), timeout int) (isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, err := ssh_conf.Stream(command, timeout)
	if err != nil {
		return isTimeout, err
	}
	// read from the output channel until the done signal is passed
OuterL:
	for {
		select {
		case done := <-doneChan:
			isTimeout = !done
			break OuterL
		case outLine := <-stdoutChan:
			lineHandler(outLine, TypeStdout)
		case errLine := <-stderrChan:
			lineHandler(errLine, TypeStderr)
		}
	}
	// return the concatenation of all signals from the output channel
	return isTimeout, err
}

// Scp uploads localPath to remotePath like native scp console app.
// Warning: remotePath should contain the file name if the localPath is a regular file,
// however, if the localPath to copy is dir, the remotePath must be the dir into which the localPath will be copied.
func (ssh_conf *SSHConfig) Scp(localPath, remotePath string) error {
	if IsDir(localPath) {
		return ssh_conf.SCopyDir(localPath, remotePath, -1, true)
	}

	if IsRegular(localPath) {
		return ssh_conf.SCopyFile(localPath, remotePath)
	}

	panic("invalid local path: " + localPath)
}

// ScpM copy multiple local file or dir to their corresponding remote path specified by para pathMappings.
func (ssh_conf *SSHConfig) ScpM(dirPathMappings map[string]string) error {
	return ssh_conf.SCopyM(dirPathMappings, -1, true)
}

// RunScript run a serial of commands on remote
func (ssh_conf *SSHConfig) RunScript(script string) error {
	return ssh_conf.Work(func(s *ssh.Session) error {
		s.Stdin = bufio.NewReader(strings.NewReader(script))
		stdout, err := s.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := s.StderrPipe()
		if err != nil {
			return err
		}

		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)

		err = s.Shell()
		if err != nil {
			return err
		}
		return s.Wait()
	})
}

// RunScriptFile run a script file on remote
func (ssh_conf *SSHConfig) RunScriptFile(script string) error {
	bytes, err := ioutil.ReadFile(script)
	if err != nil {
		return err
	}

	return ssh_conf.RunScript(string(bytes))
}
