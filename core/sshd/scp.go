package sshd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TNK-Studio/gortal/config"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

const (
	flagCopyFile       = "C"
	flagStartDirectory = "D"
	flagEndDirectory   = "E"
	flagTime           = "T"
)

const (
	responseOk        uint8 = 0
	responseError     uint8 = 1
	responseFailError uint8 = 2
)

type response struct {
	Type    uint8
	Message string
}

// ParseResponse Reads from the given reader (assuming it is the output of the remote) and parses it into a Response structure
func parseResponse(reader io.Reader) (response, error) {
	buffer := make([]uint8, 1)
	_, err := reader.Read(buffer)
	if err != nil {
		return response{}, err
	}

	responseType := buffer[0]
	message := ""
	if responseType > 0 {
		bufferedRader := bufio.NewReader(reader)
		message, err = bufferedRader.ReadString('\n')
		if err != nil {
			return response{}, err
		}
	}

	return response{responseType, message}, nil
}

func (r *response) IsOk() bool {
	return r.Type == responseOk
}

func (r *response) IsError() bool {
	return r.Type == responseError
}

// Returns true when the remote responded with an error
func (r *response) FailError() bool {
	return r.Type == responseFailError
}

// Returns true when the remote answered with a warning or an error
func (r *response) IsFailure() bool {
	return r.Type > 0
}

// Returns the message the remote sent back
func (r *response) GetMessage() string {
	return r.Message
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
			replyErr(sess, err)
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

func parseServerPath(fullPath, filename, currentUsername string) (*config.SSHUser, *config.Server, string, error) {
	args := strings.SplitN(fullPath, ":", 2)
	invaildPathErr := errors.New(
		"Please input your server key before your target path, like 'scp -P 2222 /tmp/tmp.file user@jumpserver:user@server1:/tmp/tmp.file'",
	)

	invaildSSHUserErr := errors.New("Please make sure you have ssh user to access this server. ")

	if len(args) < 2 {
		return nil, nil, "", invaildPathErr
	}

	inputServer, remotePath := args[0], args[1]
	serverArgs := strings.SplitN(inputServer, "@", 2)
	if len(serverArgs) < 2 {
		return nil, nil, "", invaildPathErr
	}

	sshUsername, serverKey := serverArgs[0], serverArgs[1]
	server := (*config.Conf.Servers)[serverKey]
	if server == nil {
		err := fmt.Errorf("Server key '%s' of server not found. ", serverKey)
		return nil, nil, "", err
	}

	if *server.SSHUsers == nil {
		return nil, nil, "", invaildSSHUserErr
	}

	var user *config.SSHUser

loop:
	for _, sshUser := range *server.SSHUsers {
		if (*sshUser).SSHUsername == sshUsername {

			if sshUser.AllowUsers == nil || len(*sshUser.AllowUsers) < 1 {
				user = sshUser
				break loop
			}

			for _, allowUser := range *sshUser.AllowUsers {
				if allowUser == currentUsername {
					user = sshUser
					break loop
				}
			}
		}
	}

	if user == nil {
		return nil, nil, "", invaildSSHUserErr
	}

	return user, server, remotePath, nil
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func checkResponse(r io.Reader) error {
	response, err := parseResponse(r)
	if err != nil {
		return err
	}

	if response.IsFailure() {
		return errors.New(response.GetMessage())
	}

	return nil

}

func copyFileToServer(bfReader *bufio.Reader, size int64, filename, filePath string, perm string, sess *ssh.Session) error {
	sshUser, server, filePath, err := parseServerPath(filePath, filename, (*sess).User())
	if err != nil {
		return err
	}
	err = replyOk(sess)
	if err != nil {
		return err
	}

	client, err := NewSSHClient(server, sshUser)
	if err != nil {
		return err
	}

	clientSess, err := client.NewSession()
	if err != nil {
		return err
	}

	err = copyToSession(bfReader, clientSess, perm, filePath, filename, size)
	if err != nil {
		return err
	}

	err = replyOk(sess)
	if err != nil {
		return err
	}

	return nil
}

func copyToSession(r *bufio.Reader, clientSess *gossh.Session, perm, filePath, filename string, size int64) error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)
	defer func() {
		select {
		case <-errCh:
			return
		default:
		}
		close(errCh)
	}()

	stdout, err := clientSess.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {
		defer wg.Done()
		w, err := clientSess.StdinPipe()
		if err != nil {
			errCh <- err
			return
		}

		defer w.Close()
		if err != nil {
			errCh <- err
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}

		_, err = fmt.Fprintln(w, flagCopyFile+perm, size, filename)
		if err != nil {
			errCh <- err
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}

		// Create a temp file
		tmp, err := createTmpFile(r, perm, size)
		if err != nil {
			errCh <- err
			return
		}

		tmpReader := bufio.NewReader(tmp)
		io.Copy(w, tmpReader)

		_, err = fmt.Fprint(w, "\x00")
		if err != nil {
			errCh <- err
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err := clientSess.Run(fmt.Sprintf("scp -t %s", filePath)); err == nil {
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}

	}()

	// if waitTimeout(&wg, time.Second*300) {
	// 	return errors.New("timeout when upload files")
	// }
	//

	// Todo Timeout Handling
	wg.Wait()

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func createTmpFile(r *bufio.Reader, perm string, size int64) (*os.File, error) {
	fileMode, err := strconv.ParseUint(perm, 8, 0)
	if err != nil {
		return nil, err
	}

	tmpFilePath := fmt.Sprintf("/tmp/gortal-tmp-file-%d", time.Now().UnixNano())
	f, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	var off int64
	buf := make([]byte, 2048)
	for {
		n, err := r.Read(buf)
		buffSize := int64(n)

		if err != nil && err != io.EOF {
			return nil, err
		}

		if off+buffSize > size && buf[n-1] == '\x00' {
			_, err := f.WriteAt(buf[:n-1], off)
			if err != nil {
				return nil, err
			}
			break
		} else if off+buffSize > size && buf[n-1] != '\x00' {
			return nil, errors.New("File size not match. ")
		}

		_, err = f.WriteAt(buf, off)
		if err != nil {
			return nil, err
		}
		off = off + buffSize
	}

	tmp, err := os.Open(tmpFilePath)
	if err != nil {
		return nil, err
	}

	return tmp, nil
}

func parseMsg(msg string) (string, string, int64, string, error) {
	strs := strings.Split(msg, " ")
	if len(strs) < 3 {
		return "", "0644", 0, "", errors.New("Message parsed error")
	}

	size, err := strconv.Atoi(strs[1])
	if err != nil {
		return "", "0644", 0, "", errors.New("Message parsed error")
	}

	permissions, filename := strs[0], strs[2]

	flag := permissions[0:1]
	permissions = permissions[1:]
	filename = filename[:len(filename)-1]

	if err != nil {
		return "", "0644", 0, "", errors.New("Message parsed error")
	}

	return flag, permissions, int64(size), filename, nil
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
	_, err = bufferedWriter.Write([]byte(strings.ReplaceAll(replyErr.Error(), "\n", " ")))
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
