# gortal 

A lightweight jumpserver written by Go, use to manage & jump to any servers.

一个超级轻量的堡垒机（跳板机）服务。  

## 部署方式

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