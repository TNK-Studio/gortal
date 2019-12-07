package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/core/jump"
	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/gliderlabs/ssh"
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

func main() {
	flag.Parse()

	if !utils.FileExited(*hostKeyFile) {
		sshd.GenKey(*hostKeyFile)
	}

	ssh.Handle(func(s ssh.Session) {
		defer func() {
			s.Close()
		}()
		jps := jump.JumpService{}
		jps.Run(&s)
	})

	log.Printf("starting ssh server on port %d...\n", *Port)
	log.Fatal(ssh.ListenAndServe(
		fmt.Sprintf(":%d", *Port),
		nil,
		ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
			config.Conf.ReadFrom(*config.ConfPath)
			if (len(*config.Conf.Users)) < 1 {
				return pass == "newuser"
			}
			return jump.VarifyUser(ctx, pass)
		}),
		ssh.HostKeyFile(utils.FilePath(*hostKeyFile)),
	),
	)
}
