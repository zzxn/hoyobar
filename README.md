# HoYoBar

## 本地测试环境搭建

在Docker中运行Redis和MySQL

```bash
docker run -itd --name redis-test -p 6379:6379 redis
docker run -itd --name mysql-test -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql
```

在MySQL容器中创建数据库：

```
mysql -u root -p 
password

mysql> CREATE DATABASE `hoyobar_test`;
```

## 运行方式

```bash
go run .
```

## 密码规则

密码包含 数字,英文,字符中的两种以上，长度6-20

## 进度

### 功能

[x] 项目框架搭建 TESTED

[x] 用户登录注册 TESTED

[x] 帖子发布、列表、查看 TESTED

[x] 帖子回复

### 优化

[x] 用户表分表

[ ] 利用Redis缓存优化性能
    [x] user
    [ ] post
