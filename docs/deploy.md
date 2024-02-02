## Docker Deployment

*以下步骤需要在Linux system上操作*

### 1. Clone the Repository and Initialize
```shell
git clone https://github.com/openimsdk/openim-docker openim-docker && cd openim-docker && make init
```

### 2. Setting OPENIM_IP
```shell
# 宿主机内网ip地址
export OPENIM_IP="Internal IP"
```

### 3. 修改Minio的外网地址
*需要把下面的`{OPENIM_IP}`替换成之前部署配置地址（`宿主机内网ip地址`）*
```shell
sed -i 's|apiURL: "http://{OPENIM_IP}:10002"|apiURL: "https://qaimapi.btcmetaswap.com/api
"|g' openim-server/config/config.yaml

sed -i 's|signEndpoint: "http://{OPENIM_IP}:10005"|signEndpoint: "https://qaimapi.btcmetaswap.com/im-minio-api
"|g' openim-server/config/config.yaml
```

### 4. Start Services and View Logs

把`docker-compose.yaml`上传到`openim-docker`目录

```shell
docker compose up -d
docker ps
docker compose logs -f openim-chat
docker compose logs -f openim-server
```

### 5. Quick Verification
查看各容器container运行状态 或 安装dweb im测试
 

## nginx配置

上传nginx-im.conf，并修改几处变量：

1. {SERVER_NAME}
2. {CERT}\{CERT_KEY}
3. {OPENIM_IP} 是第2步设置的值