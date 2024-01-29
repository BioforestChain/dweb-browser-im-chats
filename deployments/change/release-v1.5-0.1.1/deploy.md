## Docker Deployment

*以下步骤需要在Linux system上操作*

### 1. 修改Minio的外网地址
a. *把下面的`{OPENIM_DOCKER_PATH}`替换成`openim-docker`所在路径*    
b. *把下面的`{OPENIM_IP}`替换成之前部署配置地址（`宿主机内网ip地址`）*
```shell
cd {OPENIM_DOCKER_PATH}/openim-docker

sed -i 's|apiURL: "http://{OPENIM_IP}:10002"|apiURL: "https://qaimapi.btcmetaswap.com/api
"|g' openim-server/config/config.yaml

sed -i 's|signEndpoint: "http://{OPENIM_IP}:10005"|signEndpoint: "https://qaimapi.btcmetaswap.com/im-minio-api
"|g' openim-server/config/config.yaml
```

### 2. ReStart Services and View Logs

把`docker-compose.yaml`上传到`openim-docker`目录

```shell
docker compose down
docker compose up -d
docker ps
docker compose logs -f dweb-im
docker compose logs -f openim-server
```

### 3. Quick Verification
查看各容器container运行状态 或 安装dweb im测试
 

## nginx配置

上传nginx-im.conf，并修改几处变量：

1. {SERVER_NAME}
2. {CERT}\{CERT_KEY}
3. {OPENIM_IP} 是第2步设置的值