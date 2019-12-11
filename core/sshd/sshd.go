package sshd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TNK-Studio/gortal/utils"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/helloyi/go-sshclient"
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

// Connect connect server
func Connect(host string, port int, username string, privKeyFile string, sess *ssh.Session) error {
	client, err := sshclient.DialWithKey(
		fmt.Sprintf("%s:%d", host, port),
		username,
		utils.FilePath(privKeyFile),
	)

	if err != nil {
		return err
	}

	// default terminal

	terminal := client.Terminal(nil)
	terminal = terminal.SetStdio(*sess, *sess, *sess)
	if terminal.Start(); err != nil {
		return err
	}

	return nil
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
