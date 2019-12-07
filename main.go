package main

import (
	"flag"
	"fmt"
	"github.com/TNK-Studio/gortal/core/jump"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/gliderlabs/ssh"
	"log"
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
	err := jump.Configurate()
	if err != nil {
		logger.Logger.Infof("%s\n", err)
		return
	}

	ssh.Handle(func(s ssh.Session) {
		defer func() {
			if e, ok := recover().(error); ok {
				logger.Logger.Error(e)
			}
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
			return jump.VarifyUser(ctx, pass)
		}),
		ssh.HostKeyFile(utils.FilePath(*hostKeyFile)),
	),
	)
}
