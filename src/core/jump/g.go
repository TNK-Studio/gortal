package jump

import (
	"flag"
	"fmt"
	"os"

	"github.com/TNK-Studio/gortal/src/config"
	"github.com/TNK-Studio/gortal/src/core/pui"
	"github.com/TNK-Studio/gortal/src/utils/logger"
	"github.com/gliderlabs/ssh"
)

func init() {
	config.ConfPath = flag.String("c", fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.gortal.yml"), "Config file")
}

// JumpService JumpService
type JumpService struct {
	sess      *ssh.Session
	persionUI *pui.PUI
}

func (jps *JumpService) setSession(sess *ssh.Session) {
	jps.sess = sess
}

// Run jump
func (jps *JumpService) Run(s *ssh.Session) {
	jps.setSession(s)
	jps.persionUI = &pui.PUI{}
	jps.persionUI.SetSession(jps.sess)
	jps.persionUI.ShowMainMenu()
}

func setupConfig() {
	fmt.Println("Config file not found. Setup config.", *config.ConfPath)
	_, _, err := pui.CreateUser(false, true, nil)
	if err != nil {
		return
	}
	serverKey, _, err := pui.AddServer(nil)
	if err != nil {
		return
	}
	_, _, err = pui.AddServerSSHUser(*serverKey, nil)
	if err != nil {
		return
	}
	config.Conf.SaveTo(*config.ConfPath)
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
func Configurate() {
	if *config.ConfPath == "" {
		fmt.Println("Please specify a config file.")
		return
	}
	fmt.Println("Read config file", *config.ConfPath)
	if !config.ConfigFileExisted(*config.ConfPath) {
		setupConfig()
	} else {
		config.Conf.ReadFrom(*config.ConfPath)
		if config.Conf.Users == nil || len(*config.Conf.Users) < 1 {
			_, _, err := pui.CreateUser(false, true, nil)
			if err != nil {
				return
			}
		}
	}
}
