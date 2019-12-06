package sshd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
)

// CopyID CopyID
func CopyID(username, host string, port int, passwd, pubKeyFile string) error {
	client, err := GetClientByPasswd(username, host, port, passwd)
	if err != nil {
		return err
	}

	file, err := os.Open(utils.FilePath(pubKeyFile))
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	pubKey := fmt.Sprintf("%s %s@%s", string(b), username, host)

	copyIDCmd := fmt.Sprintf("echo \"%s\" >> ~/.ssh/authorized_keys", pubKey)
	copyIDCmd = strings.ReplaceAll(copyIDCmd, "\n", "")
	logger.Logger.Debugf("CopyID run command:\n%s", copyIDCmd)

	out, err := client.Cmd(copyIDCmd).Output()
	if err != nil {
		return err
	}

	logger.Logger.Debugf("%s", string(out))

	return nil
}
