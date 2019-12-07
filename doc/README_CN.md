# gortal 

一个使用 `Go` 语言开发的，超级轻量的堡垒机（跳板机）服务。[English Document](../README.md) | [中文文档](./doc/README_CN.md)

![gortal](./gortal.gif)

## 部署方式

gortal 需要一台拥有公网 IP 的服务器作为跳板机的服务器。  
此服务器需要有外网访问权限，以便能访问你需要访问的目标服务器。  

### Docker

```shell
$ docker push elfgzp/gortal:tagname
$ mkdir -p ~/.gortal/.ssh
$ docker run \
  -p 2222:2222 \
  -v ~/.gortal:/root\
  -v ~/.gortal/.ssh:/root/.ssh\
  --name gortal -d gortal:latest
```

### 二进制文件

在 [Release](https://github.com/TNK-Studio/gortal/releases) 页面下载你需要的版本，解压得到 `gortal` 二进制可执行程序，然后运行。  

```shell
$ ./gortal
starting ssh server on port 2222...
```

## 使用说明  

gortal 服务启动后会在 `2222` 端口开启一个 `sshd` 服务，你也可以通过 `-p` 设置启动端口。  

服务启动后，你只需要使用 `ssh` 命令访问该服务，注意访问的为你跳板机的 ip 或域名，这里演示本地 ip。  

```shell
$ ssh 127.0.0.1 -p 2222
root@127.0.0.1's password:
New Username: root█
Password: ******█
Confirm your password: ******█
Please login again with your new acount.
Shared connection to 127.0.0.1 closed.
```

第一次访问默认用户密码为 `newuser`，然后命令行会提示新建用户，按照提示步骤新建一个跳板机的 `admin` 账户。  

```shell
$ ssh root@127.0.0.1 -p 2222
root@127.0.0.1's password:
Use the arrow keys to navigate: ↓ ↑ → ←
? Please select the function you need:
  ▸ List servers
    Edit users
    Edit servers
    Quit
```

再次使用你的密码登录后就可以使用了。  

