package utils

import "github.com/elfgzp/ssh"

// If If
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// SessIO SessIO
func SessIO(sess *ssh.Session) ssh.Session {
	var stdio ssh.Session
	if sess != nil {
		stdio = *sess
	} else {
		stdio = nil
	}
	return stdio
}
