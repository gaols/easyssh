# easyssh

## Description

Package easyssh provides a simple implementation of some SSH protocol features in Go.
You can simply run command on remote server or upload a file even simple than native console SSH client.
Do not need to think about Dials, sessions, defers and public keys...Let easyssh will be think about it!

## Scp

Scp support single file or a directory.

```
sshconfig := &easyssh.SSHConfig{...}
sshconfig.Scp(localpath, remotepath)
```

ScpM support copy multiple dir to multiple destination on remote simultaneously.

```
sshconfig := &easyssh.SSHConfig{...}
sshconfig.ScpM(pathmapping)
```

## Install

```
go get github.com/gaols/easyssh
```

## So easy to use

[Run a command on remote server and get STDOUT output](https://github.com/gaols/easyssh/blob/master/example/run.go)

[Run a command on remote server and get STDOUT output line by line](https://github.com/gaols/easyssh/blob/master/example/rtrun.go)

[Upload a file to remote server](https://github.com/gaols/easyssh/blob/master/example/scp.go)

[Upload a directory to remote server](https://github.com/gaols/easyssh/blob/master/example/scopy.go)
