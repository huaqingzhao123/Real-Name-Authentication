FROM alpine

# 设置时区为上海
RUN apk add --no-cache tzdata && ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime\
    && echo "Asia/Shanghai" > /etc/timezone

COPY realname /realname
ENTRYPOINT [ "/realname" ]