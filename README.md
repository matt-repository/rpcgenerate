#### 谢谢你的star. 💕💕

自动生成CURD proto文件和c# CURD rpc service 文件

### 用法
1、将templates 文件夹和 可执行文件(rpcgenerate) 放在同一目录 

2、执行以下命令

#### 命令行使用::

```
$ rpcgenerate -h
Usage of ./rpcgenerate:
    -conn string
        the connect string
  -db string
        the database type,mysql|sqlserver (default "mysql")
  -ignore_table string
        ignore table,multiple tables ',' split
  -schema string
        the database schema
  -table string
        table,multiple tables ',' split, empty  is all
  -type string
        generate file type ,proto|csharp_service (default "proto")
```

```
$ ./rpcgenerate -db mysql -conn "root:123456@tcp(localhost:3306)/test" -schema test -type csharp_service 
$ ./rpcgenerate -db mysql -conn "root:123456@tcp(localhost:3306)/test" -schema test -type proto
$ ./rpcgenerate -db sqlserver -conn "Provider=SQLOLEDB;Data Source=localhost,1433;Initial Catalog=test;user id=sa;password=123456;" -schema test -type csharp_service
$ ./rpcgenerate -db sqlserver -conn "Provider=SQLOLEDB;Data Source=localhost,1433;Initial Catalog=test;user id=sa;password=123456;" -schema test -type proto
```


#### 有问题添加微信: fq943609
