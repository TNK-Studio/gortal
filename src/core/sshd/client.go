package sshd

import (
	"fmt"

	"github.com/TNK-Studio/gortal/src/utils"
	"github.com/gliderlabs/ssh"
	sshclient "github.com/helloyi/go-sshclient"
)

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
