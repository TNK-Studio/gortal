package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"gopkg.in/yaml.v2"
)

// Conf config obj
var Conf *Config

// ConfPath config file path
var ConfPath *string

func init() {
	users := make(map[string]*User)
	servers := make(map[string]*Server)
	Conf = &Config{
		Users:   &users,
		Servers: &servers,
	}
}

// Config config
type Config struct {
	Users   *map[string]*User   `yaml:"users"`
	Servers *map[string]*Server `yaml:"servers"`
}

// User gortal login user
type User struct {
	Username   string `yaml:"username"`
	HashPasswd string `yaml:"hashPasswd"`
	Admin      bool   `yaml:"admin"`
}

// Server server
type Server struct {
	Name     string               `yaml:"name"`
	Host     string               `yaml:"host"`
	Port     int                  `yaml:"port"`
	SSHUsers *map[string]*SSHUser `yaml:"sshUsers"`
}

// SSHUser ssh user
type SSHUser struct {
	SSHUsername  string    `yaml:"sshUsername"`
	IdentityFile string    `yaml:"identityFile"`
	AllowUsers   *[]string `yaml:"allowUsers,omitempty"`
}

// ReadFrom read config
func (c *Config) ReadFrom(path string) error {
	users := make(map[string]*User)
	servers := make(map[string]*Server)
	config := &Config{
		Users:   &users,
		Servers: &servers,
	}

	var readConfigErr *error
	defer func() {
		if readConfigErr != nil && *readConfigErr != nil {
			*c = *config
		}
	}()

	configFile, err := ioutil.ReadFile(utils.FilePath(path))
	if err != nil {
		logger.Logger.Warningf("Error reading YAML file: %s\n", err)
		readConfigErr = &err
		return err
	}

	err = yaml.Unmarshal([]byte(configFile), config)
	if err != nil {
		logger.Logger.Warningf("Error parsing YAML file: %s\n", err)
		readConfigErr = &err
		return err
	}

	*c = *config
	return nil
}

// SaveTo save config
func (c *Config) SaveTo(path string) error {
	logger.Logger.Infof("Save config to '%s'\n", path)
	bytes, err := yaml.Marshal(c)
	if err != nil {
		logger.Logger.Infof("Error parsing YAML obj: %s\n", err)
		return err
	}
	ioutil.WriteFile(path, bytes, 0644)
	return nil
}

// AddUser add user to config
func (c *Config) AddUser(username string, password string, IsAdmin bool) (string, *User) {
	// Todo Add sha256 password
	user := &User{
		Username:   username,
		HashPasswd: password,
		Admin:      IsAdmin,
	}
	userAmount := len(*c.Users) + 1
	log.Printf("Add user, user amount: %d", userAmount)
	key := fmt.Sprintf("%s%d", "users", userAmount)
	(*c.Users)[key] = user
	return key, user
}

// AddServer add server to config
func (c *Config) AddServer(name string, host string, port int) (string, *Server) {
	server := &Server{
		Name: name,
		Host: host,
		Port: port,
	}
	serverAmount := len(*c.Servers) + 1
	log.Printf("Add server, server amount: %d", serverAmount)
	key := fmt.Sprintf("%s%d", "server", serverAmount)
	(*c.Servers)[key] = server
	return key, server
}

// AddServerSSHUser add server ssh user to config
func (c *Config) AddServerSSHUser(serverKey string, username string, identityFile string, allowUsers *[]string) (string, *SSHUser) {
	sshUser := &SSHUser{
		SSHUsername:  username,
		IdentityFile: identityFile,
		AllowUsers:   allowUsers,
	}

	server := (*c.Servers)[serverKey]
	if server == nil {
		return "", nil
	}

	if server.SSHUsers == nil {
		sshUsers := make(map[string]*SSHUser)
		server.SSHUsers = &sshUsers
	}

	serverSSHUserAmount := len(*server.SSHUsers) + 1
	log.Printf("Add server ssh user, server ssh user amount: %d", serverSSHUserAmount)
	key := fmt.Sprintf("%s%d", "sshUser", serverSSHUserAmount)

	(*server.SSHUsers)[key] = sshUser
	return key, sshUser
}

// GetUserByUsername GetUserByUsername
func (c *Config) GetUserByUsername(username string) *User {
	for _, user := range *c.Users {
		if user.Username == username {
			return user
		}
	}
	return nil
}

// GetServerByName GetServerByName
func (c *Config) GetServerByName(name string) *Server {
	for _, server := range *c.Servers {
		if server.Name == name {
			return server
		}
	}
	return nil
}

// GetUserServers get user servers list
func (c *Config) GetUserServers(user *User) map[string]*Server {
	servers := make(map[string]*Server, 0)
	for serverKey, server := range *c.Servers {
		if (*server).SSHUsers == nil {
			return servers
		}
	loop:
		for _, sshUser := range *server.SSHUsers {
			if sshUser.AllowUsers == nil {
				break loop
			}

			for _, username := range *sshUser.AllowUsers {
				if user.Username == username {
					break loop
				}
			}
		}
		servers[serverKey] = server
	}
	return servers
}

// GetServerSSHUsers get all allow server' s ssh users
func (c *Config) GetServerSSHUsers(user *User, server *Server) map[string]*SSHUser {
	sshUsers := make(map[string]*SSHUser, 0)
	for sshUserKey, sshUser := range *server.SSHUsers {
		if sshUser.AllowUsers == nil {
			sshUsers[sshUserKey] = sshUser
			continue
		}

		for _, username := range *sshUser.AllowUsers {
			if user.Username == username {
				sshUsers[sshUserKey] = sshUser
			}
		}
	}
	return sshUsers
}

// ReIndexUser ReIndexUser
func (c *Config) ReIndexUser() {
	users := make(map[string]*User, 0)
	i := 1

	for _, user := range *c.Users {
		key := fmt.Sprintf("user%d", i)
		users[key] = user
	}
	(*c).Users = &users
}

// ReIndexServer ReIndexServer
func (c *Config) ReIndexServer() {
	servers := make(map[string]*Server, 0)
	i := 1

	for _, server := range *c.Servers {
		key := fmt.Sprintf("user%d", i)
		servers[key] = server
	}
	(*c).Servers = &servers
}
