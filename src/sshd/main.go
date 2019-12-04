package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

// Serve 运行 sshd
func Serve() {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("zsh")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			ptmx, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			go func() {
				for win := range winCh {
					winSize := &pty.Winsize{
						Rows: uint16(win.Height),
						Cols: uint16(win.Width),
						X:    0,
						Y:    0,
					}
					if err := pty.Setsize(ptmx, winSize); err != nil {
						log.Printf("error resizing pty: %s", err)
					}
				}
			}()
			go func() {
				io.Copy(ptmx, s) // stdin
			}()
			io.Copy(s, ptmx) // stdout
			cmd.Wait()
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil))
}

func main() {
	Serve()
}
