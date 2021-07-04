#!/bin/bash
mysql -uroot -p$MYSQL_ROOT_PASSWORD << EOF
source /usr/local/work/数据库脚本/verifySystem.sql;
source /usr/local/work/存储过程/实名认证系统.sql;