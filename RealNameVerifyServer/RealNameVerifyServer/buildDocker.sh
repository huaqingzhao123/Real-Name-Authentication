#!/bin/sh
v="k8s"

chmod +x ./realname
docker build . -t registry.cn-hangzhou.aliyuncs.com/paycenter/realname:${v}
docker push registry.cn-hangzhou.aliyuncs.com/paycenter/realname:${v}