#!/bin/sh
# start a docker running test 
docker run -itd --name mysql-test -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql
