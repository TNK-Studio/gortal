package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/core/jump"
	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/elfgzp/ssh"
)

var (
	// Port port
	Port *int

	hostKeyFile *string
)

func init() {
	Port = flag.Int("p", 2222, "Port")
	hostKeyFile = flag.String("hk", "~/.ssh/id_rsa", "Host key file")
}

func passwordAuth(ctx ssh.Context, pass string) bool {
	config.Conf.ReadFrom(*config.ConfPath)
	var success bool
	if (len(*config.Conf.Users)) < 1 {
		success = (pass == "newuser")
	} else {
		success = jump.VarifyUser(ctx, pass)
	}
	if !success {
		time.Sleep(time.Second * 3)
	}
	return success
}

func sessionHandler(sess *ssh.Session) {
	defer func() {
		(*sess).Close()
	}()

	rawCmd := (*sess).RawCommand()
	cmd, args, err := sshd.ParseRawCommand(rawCmd)
	if err != nil {
		sshd.ErrorInfo(err, sess)
		return
	}

	switch cmd {
	case "scp":
		sshd.ExecuteSCP(args, sess)
	default:
		sshHandler(sess)
	}
}

func sshHandler(sess *ssh.Session) {
	jps := jump.Service{}
	jps.Run(sess)
}

func scpHandler(args []string, sess *ssh.Session) {
	sshd.ExecuteSCP(args, sess)
}

func main() {
	flag.Parse()

	if !utils.FileExited(*hostKeyFile) {
		sshd.GenKey(*hostKeyFile)
	}

	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			if e, ok := recover().(error); ok {
				logger.Logger.Panic(e)
			}
		}()
		sessionHandler(&sess)
	})

	log.Printf("starting ssh server on port %d...\n", *Port)
	log.Fatal(ssh.ListenAndServe(
		fmt.Sprintf(":%d", *Port),
		nil,
		ssh.PasswordAuth(passwordAuth),
		ssh.HostKeyFile(utils.FilePath(*hostKeyFile)),
	),
	)
}
