
## nginx配置

把`nginx-im-admin.conf`里的配置复制到`nginx-im.conf`里（因为存在upstream的引用），或则把所有upstream提取出来放置到
`http块`的开始位置，例如
```text
http {
    include /etc/nginx/conf.d/upstreams.conf;
    ...
}
```

修改几处变量：

1. {SERVER_ADMIN_NAME}
2. {CERT}\{CERT_KEY} 
3. {OPENIM_IP}