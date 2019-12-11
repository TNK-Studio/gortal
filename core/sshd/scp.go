package sshd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/gliderlabs/ssh"
)

const (
	flagCopyFile       = "C"
	flagStartDirectory = "D"
	flagEndDirectory   = "E"
	flagTime           = "T"
)

const (
	responseOk        = 0
	responseError     = 1
	responseFailError = 2
)

type sourceProtocol struct {
	remIn     io.WriteCloser
	remOut    io.Reader
	remReader *bufio.Reader
}

// ExecuteSCP ExecuteSCP
func ExecuteSCP(args []string, sess *ssh.Session) error {
	err := replyOk(sess)
	if err != nil {
		return err
	}

	bufferedReader := bufio.NewReader(*sess)
	message, err := bufferedReader.ReadString('\n')
	if err != nil {
		replyErr(sess, err)
		return err
	}

	flag, perm, size, filename, err := parseMsg(message)
	if err != nil {
		replyErr(sess, err)
		return err
	}
	switch flag {
	case flagCopyFile:
		err = copyFileToServer(bufferedReader, size, filename, args[1], perm, sess)
		if err != nil {
			return err
		}
	case flagEndDirectory:
	case flagStartDirectory:
		replyErr(sess, errors.New("Folder transfer is not yet supported. You can try to compress the folder and upload it. "))
	default:
		return nil
	}

	return nil
}

func copyFileToServer(bfReader *bufio.Reader, size int64, filename, filePath string, perm os.FileMode, sess *ssh.Session) error {
	args := strings.SplitN(filePath, ":", 2)
	if len(args) < 2 {
		err := errors.New(
			"Please input your server key before your target path, like 'scp -P 2222 /tmp/tmp.file user@jumpserver:server1:/tmp/tmp.file'",
		)
		replyErr(sess, err)
		return err
	}

	serverKey, filePath := args[0], args[1]
	server := (*config.Conf.Servers)[serverKey]
	if server == nil {
		err := fmt.Errorf("Server key '%s' of server not found. ", serverKey)
		replyErr(sess, err)
		return err
	}

	err := replyOk(sess)
	if err != nil {
		return err
	}

	dir := path.Dir(filePath)
	logger.Logger.Debug(dir)

	if filePath[len(filePath)-1:] == "/" {
		filePath = filePath[:len(filePath)-1]
	}
	if dir == filePath {
		filePath = fmt.Sprintf("%s/%s", filePath, filename)
	} else {
		filePath = fmt.Sprintf("%s/", filePath)
	}

	// Create a temp file
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(perm))
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	var off int64
	buf := make([]byte, 2048)
	for {
		n, err := bfReader.Read(buf)
		buffSize := int64(n)

		if err != nil && err != io.EOF {
			return err
		}

		if off+buffSize >= size && buf[n-1] == '\x00' {
			_, err := f.WriteAt(buf[:n-1], off)
			if err != nil {
				return err
			}
			break
		} else if off+buffSize >= size && buf[n-1] != '\x00' {
			return errors.New("File size not match. ")
		}

		_, err = f.WriteAt(buf, off)
		if err != nil {
			return err
		}
		off = off + buffSize
	}
	err = replyOk(sess)
	if err != nil {
		return err
	}
	return nil
}

func parseMsg(msg string) (string, os.FileMode, int64, string, error) {
	strs := strings.Split(msg, " ")
	if len(strs) < 3 {
		return "", 0644, 0, "", errors.New("Message parsed error")
	}

	size, err := strconv.Atoi(strs[1])
	if err != nil {
		return "", 0644, 0, "", errors.New("Message parsed error")
	}

	permissions, filename := strs[0], strs[2]

	flag := permissions[0:1]
	permissions = permissions[1:]
	filename = filename[:len(filename)-1]

	perm, err := strconv.ParseInt(permissions, 8, 0)
	if err != nil {
		return "", 0644, 0, "", errors.New("Message parsed error")
	}

	return flag, os.FileMode(perm), int64(size), filename, nil
}

func replyOk(sess *ssh.Session) error {
	bufferedWriter := bufio.NewWriter(*sess)
	_, err := bufferedWriter.Write([]byte{responseOk})

	if err != nil {
		return err
	}

	err = bufferedWriter.Flush()
	if err != nil {
		return err
	}
	return nil
}

func replyErr(sess *ssh.Session, replyErr error) error {
	bufferedWriter := bufio.NewWriter(*sess)
	_, err := bufferedWriter.Write([]byte{responseError})
	_, err = bufferedWriter.Write([]byte(replyErr.Error()))
	_, err = bufferedWriter.Write([]byte{'\n'})

	if err != nil {
		return err
	}

	err = bufferedWriter.Flush()
	if err != nil {
		return err
	}
	return nil
}
