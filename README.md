#### 谢谢你的star. 💕💕

自动生成CURD proto文件和c# CURD rpc service 文件

### 用法

#### 命令行使用::

`go install github.com/matt-repository/rpcgenerate@latest`

```
$ rpcgenerate -h
Usage of ./rpcgenerate:
  -conn string
        the connect string
  -db string
        the database type (default "mysql")
  -ignore_table string
        a comma spaced list of tables to ignore
  -schema string
        the database schema
  -table string
        the table schema，multiple tables ',' split. 
  -type string
        generate file type ,proto|csharp_service (default "proto")

```

```
$ ./rpcgenerate -db mysql -conn "root:123456@tcp(localhost:3306)/test" -schema test -type csharp_service 
$ ./rpcgenerate -db mysql -conn "root:123456@tcp(localhost:3306)/test" -schema test -type proto
$ ./rpcgenerate -db sqlserver -conn "Provider=SQLOLEDB;Data Source=localhost,1433;Initial Catalog=test;user id=sa;password=123456;" -schema test -type csharp_service
$ ./rpcgenerate -db sqlserver -conn "Provider=SQLOLEDB;Data Source=localhost,1433;Initial Catalog=test;user id=sa;password=123456;" -schema test -type proto



```


#### golang导入

```sh
$ go get -u github.com/matt-repository/rpcgenerate@latest
```

#### 有问题添加微信: fq943609
