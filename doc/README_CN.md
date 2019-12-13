# gortal

[![Actions Status](https://img.shields.io/github/workflow/status/TNK-Studio/gortal/Build%20release)](https://github.com/TNK-Studio/gortal/actions)[![Docker build](https://img.shields.io/docker/cloud/build/elfgzp/gortal)](https://hub.docker.com/repository/docker/elfgzp/gortal)[![Docker build automated](https://img.shields.io/docker/cloud/automated/elfgzp/gortal)](https://hub.docker.com/repository/docker/elfgzp/gortal)
[![Docker pull](https://img.shields.io/docker/pulls/elfgzp/gortal)](https://hub.docker.com/repository/docker/elfgzp/gortal)[![Release Download](https://img.shields.io/github/downloads/TNK-Studio/gortal/total)](https://github.com/TNK-Studio/gortal/releases)

一个使用 `Go` 语言开发的，超级轻量的跳板机服务。[English Document](../README.md) | [中文文档](./doc/README_CN.md)

![gortal](./gortal.gif)

## 部署方式

gortal 需要一台拥有公网 IP 的服务器作为跳板机的服务器。  
此服务器需要有外网访问权限，以便能访问你需要访问的目标服务器。  

### Docker

```shell
$ docker pull elfgzp/gortal:latest
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

### 第一次使用  

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
    Edit personal info
    Quit
```

再次使用你的密码登录后就可以使用了。  

### 通过跳板机上传或下载文件到服务器  

如果你想通过跳板机上传或下载文件到服务器，你可以通过 `scp` 命令按照以下格式：  

```shell
$ scp -P 2222 ~/Desktop/README.md  gzp@jumpserver:gzp@server2:~/Desktop/README1.md
README.md                                        100% 9279    73.9KB/s   00:00
```

```shell
scp -P 2222 gzp@127.0.0.1:gzp@server2:~/Desktop/video.mp4 ~/Downloads
video.mp4                           100%   10MB  58.8MB/s   00:00
```

注意在 `gzp@jumpserver` 后面用 `:` 加上你需要传输的服务器的 `key` 和 `username`，最后写上目标路径或源文件路径。目前不支持文件夹复制，请压缩文件后在上传或下载。  
