package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/TNK-Studio/gortal/src/core/jump"
	"github.com/gliderlabs/ssh"
)

var (
	// Port port
	Port *int
)

func init() {
	Port = flag.Int("p", 2222, "Port")
}
func main() {
	flag.Parse()
	jump.Configurate()
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
	),
	)
}
