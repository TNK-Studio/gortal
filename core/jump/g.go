package jump

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/core/pui"
	"github.com/TNK-Studio/gortal/core/sshd"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/elfgzp/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func init() {
	config.ConfPath = flag.String("c", fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.gortal.yml"), "Config file")
}

// Service Service
type Service struct {
	sess      *ssh.Session
	persionUI *pui.PUI
}

func (jps *Service) setSession(sess *ssh.Session) {
	jps.sess = sess
}

// Run jump
func (jps *Service) Run(sess *ssh.Session) {
	defer func() {
		(*sess).Exit(0)
	}()
	relogin, err := Configurate(sess)
	if err != nil {
		logger.Logger.Error("%s\n", err)
		return
	}
	if relogin {
		sshd.Info("Please login again with your new acount. \n", sess)
		sshConn := (*sess).Context().Value(ssh.ContextKeyConn).(gossh.Conn)
		sshConn.Close()
		return
	}
	jps.setSession(sess)
	jps.persionUI = &pui.PUI{}
	jps.persionUI.SetSession(jps.sess)
	jps.persionUI.ShowMainMenu()
}

// VarifyUser VarifyUser
func VarifyUser(ctx ssh.Context, pass string) bool {
	username := ctx.User()
	logger.Logger.Debugf("VarifyUser username: %s\n", username)
	for _, user := range *config.Conf.Users {
		// Todo Password hash
		if user.Username == username && user.HashPasswd == pass {
			return true
		}
	}
	return false
}

// Configurate Configurate
func Configurate(sess *ssh.Session) (bool, error) {
	if *config.ConfPath == "" {
		return false, errors.New("Please specify a config file. ")
	}
	logger.Logger.Info("Read config file", *config.ConfPath)
	if !utils.FileExited(*config.ConfPath) {
		_, _, err := pui.CreateUser(false, true, sess)
		if err != nil {
			sshd.ErrorInfo(err, sess)
			return false, err
		}
		config.Conf.SaveTo(*config.ConfPath)
		return true, nil
	} else {
		config.Conf.ReadFrom(*config.ConfPath)
		if config.Conf.Users == nil || len(*config.Conf.Users) < 1 {
			_, _, err := pui.CreateUser(false, true, sess)
			if err != nil {
				sshd.ErrorInfo(err, sess)
				return false, err
			}
			config.Conf.SaveTo(*config.ConfPath)
			return true, nil
		}
	}
	return false, nil
}
