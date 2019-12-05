package pui

import (
	"github.com/gliderlabs/ssh"
)

var (
	// Sess Session
	Sess ssh.Session
)

// SetSession SetSession
func SetSession(s *ssh.Session) {
	Sess = *s
}
