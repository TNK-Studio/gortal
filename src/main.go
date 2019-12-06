package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/TNK-Studio/gortal/src/core/jump"
	"github.com/TNK-Studio/gortal/src/utils"
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
	err := jump.Configurate()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	ssh.Handle(func(s ssh.Session) {
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
