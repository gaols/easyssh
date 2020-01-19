// Package easyssh provides a simple implementation of some SSH protocol
// features in Go. You can simply run a command on a remote server or get a file
// even simpler than native console SSH client. You don't need to think about
// Dials, sessions, defers, or public keys... Let easyssh think about it!
package easyssh

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"os"
	"time"

	"github.com/gaols/goutils"

	"golang.org/x/crypto/ssh"
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
func getKeyFile(keyPath string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keyPath)
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
func (sshConf *SSHConfig) connect() (*ssh.Session, error) {
	client, err := sshConf.Cli()
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
func (sshConf *SSHConfig) Cli() (*ssh.Client, error) {
	// auths holds the detected ssh auth methods
	authMethods := make([]ssh.AuthMethod, 0)

	// figure out what auths are requested, what is supported
	if sshConf.Password != "" {
		authMethods = append(authMethods, ssh.Password(sshConf.Password))
	}

	if goutils.IsNotBlank(sshConf.Key) {
		if pubKey, err := getKeyFile(sshConf.Key); err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(pubKey))
		}
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer func() {
			err := sshAgent.Close()
			log.Println(err)
		}()
	}
	// Default port 22
	sshConf.Port = goutils.DefaultIfBlank(sshConf.Port, "22")

	// Default current user
	sshConf.User = goutils.DefaultIfBlank(sshConf.User, os.Getenv("USER"))

	config := &ssh.ClientConfig{
		User:            sshConf.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// default maximum amount of time for the TCP connection to establish is 10s
	config.Timeout = time.Second * 10
	if sshConf.Timeout > 0 {
		config.Timeout = time.Duration(sshConf.Timeout) * time.Second
	}

	return ssh.Dial("tcp", sshConf.Server+":"+sshConf.Port, config)
}

func loopReader(reader io.Reader, outCh chan string, doneCh chan byte) {
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			outCh <- scanner.Text()
		}
		doneCh <- 1
	}()
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (sshConf *SSHConfig) Stream(command string, timeout int) (stdout, stderr chan string, done chan bool, err error) {
	stdout = make(chan string)
	stderr = make(chan string)
	done = make(chan bool)
	// connect to remote host
	session, err := sshConf.connect()
	if err != nil {
		return
	}
	// connect to both outputs (they are of type io.Reader)
	stdOutReader, err := session.StdoutPipe()
	if err != nil {
		return
	}
	stderrReader, err := session.StderrPipe()
	if err != nil {
		return
	}
	err = session.Start(command)
	if err != nil {
		return
	}

	// continuously send the command's output over the channel
	go func() {
		defer close(stdout)
		defer close(stderr)
		defer close(done)

		go func() {
			stdoutDone := make(chan byte)
			stderrDone := make(chan byte)
			// loop stdout
			loopReader(stdOutReader, stdout, stdoutDone)
			// loop stderr
			loopReader(stderrReader, stderr, stderrDone)
			<-stdoutDone
			<-stderrDone
			done <- true
		}()

		if timeout <= 0 {
			timeout = 24 * 3600 // a long timeout simulate wait forever
		}
		timeoutChan := time.After(time.Duration(timeout) * time.Second)
		select {
		case r := <-done:
			done <- r
		case <-timeoutChan:
			stderr <- fmt.Sprintf("Run command timeout: %s", command)
			done <- false
		}
	}()

	return
}

// Run command on remote machine and returns its stdout as a string
func (sshConf *SSHConfig) Run(command string, timeout int) (outStr, errStr string, isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, err := sshConf.Stream(command, timeout)
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
func (sshConf *SSHConfig) RtRun(command string, lineHandler func(string string, lineType int), timeout int) (isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, err := sshConf.Stream(command, timeout)
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
func (sshConf *SSHConfig) Scp(localPath, remotePath string) error {
	if goutils.IsDir(localPath) {
		return sshConf.SCopyDir(localPath, remotePath, -1, true)
	}

	if goutils.IsRegular(localPath) {
		return sshConf.SCopyFile(localPath, remotePath)
	}

	panic("invalid local path: " + localPath)
}

// ScpM copy multiple local file or dir to their corresponding remote path specified by para pathMappings.
func (sshConf *SSHConfig) ScpM(dirPathMappings map[string]string) error {
	return sshConf.SCopyM(dirPathMappings, -1, true)
}

// RunScript run a serial of commands on remote
func (sshConf *SSHConfig) RunScript(script string) error {
	return sshConf.Work(func(s *ssh.Session) error {
		s.Stdin = bufio.NewReader(strings.NewReader(script))
		stdout, err := s.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := s.StderrPipe()
		if err != nil {
			return err
		}

		go func() {
			_, err := io.Copy(os.Stdout, stdout)
			if err != nil {
				log.Println(err)
			}
		}()
		go func() {
			_, err := io.Copy(os.Stderr, stderr)
			if err != nil {
				log.Println(err)
			}
		}()

		err = s.Shell()
		if err != nil {
			return err
		}
		return s.Wait()
	})
}

// RunScriptFile run a script file on remote
func (sshConf *SSHConfig) RunScriptFile(script string) error {
	content, err := ioutil.ReadFile(script)
	if err != nil {
		return err
	}

	return sshConf.RunScript(string(content))
}
