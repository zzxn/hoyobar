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

## 本地集成测试

集成测试测试以下功能：
- 注册
- 登录
- 发帖
- 列出帖子列表
- 回帖
- 列出回帖列表

方法：
- 确保本地测试环境已配置好，且数据均已清空
- 运行服务 `go run main.go` 等待启动成功
- 运行集成测试，`python3 ./script/test.py`
    - 需要 `pip3 install grequests`

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
