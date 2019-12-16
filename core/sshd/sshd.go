package sshd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/fatih/color"
	"github.com/elfgzp/ssh"
	"github.com/helloyi/go-sshclient"
	gossh "golang.org/x/crypto/ssh"
)

// GetClientByPasswd GetClientByPasswd
func GetClientByPasswd(username, host string, port int, passwd string) (*sshclient.Client, error) {
	client, err := sshclient.DialWithPasswd(
		fmt.Sprintf("%s:%d", host, port),
		username,
		passwd,
	)

	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewTerminal NewTerminal
func NewTerminal(server *config.Server, sshUser *config.SSHUser, sess *ssh.Session) error {
	upstreamClient, err := NewSSHClient(server, sshUser)
	if err != nil {
		return nil
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return nil
	}
	defer upstreamSess.Close()

	upstreamSess.Stdout = *sess
	upstreamSess.Stdin = *sess
	upstreamSess.Stderr = *sess

	pty, winCh, _ := (*sess).Pty()

	if err := upstreamSess.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, pty.TerminalModes); err != nil {
		return err
	}

	if err := upstreamSess.Shell(); err != nil {
		return err
	}
	
	go func () {
		for win := range winCh {
			upstreamSess.WindowChange(win.Height, win.Width)
		}
	}()

	if err := upstreamSess.Wait(); err != nil {
		return err
	}

	return nil
}

// NewSSHClient NewSSHClient
func NewSSHClient(server *config.Server, sshUser *config.SSHUser) (*gossh.Client, error) {
	if !utils.FileExited(sshUser.IdentityFile) {
		return nil, errors.New("Jumpserver can not find the identity file of the target server. ")
	}

	key, err := ioutil.ReadFile(utils.FilePath(sshUser.IdentityFile))
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	config := &gossh.ClientConfig{
		User: sshUser.SSHUsername,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }),
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return client, nil
}

// ParseRawCommand ParseRawCommand
func ParseRawCommand(command string) (string, []string, error) {
	parts := strings.Split(command, " ")

	if len(parts) < 1 {
		return "", nil, errors.New("No command in payload: " + command)
	}

	if len(parts) < 2 {
		return parts[0], []string{}, nil
	}

	return parts[0], parts[1:], nil
}

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}

// Info Info
func Info(msg string, sess *ssh.Session) {
	green := color.New(color.FgGreen)
	green.Fprint(*sess, msg)
}
