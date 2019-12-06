package sshd

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/TNK-Studio/gortal/utils"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/helloyi/go-sshclient"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

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

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}
