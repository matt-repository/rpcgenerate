#### Thanks for you,give a star . 💕💕

 Generates a rpc file from your mysql database. currently contains c#、proto

### Uses

##### Tips:  If your operating system is windows, the default encoding of windows command line is "GBK", you need to change it to "UTF-8", otherwise the generated file will be messed up. 



#### Use from the command line:

`go install github.com/matt-repository/rpcgenerate@latest`

```
$ rpcgenerate -h

Usage of rpcgenerate:
  -db string
        the database type (default "mysql|sqlserver")
  -field_style string
        gen protobuf field style, sql_pb | sqlPb (default "sqlPb")
  -go_package string
        the protocol buffer go_package. defaults to the database schema.
  -host string
        the database host (default "localhost")
  -ignore_columns string
        a comma spaced list of mysql columns to ignore
  -ignore_tables string
        a comma spaced list of tables to ignore
  -package string
        the protocol buffer package. defaults to the database schema.
  -password string
        the database password (default "root")
  -port int
        the database port (default 3306)
  -schema string
        the database schema
  -service_name string
        the service name , defaults to the database schema+'Service'
  -table string
        the table schema，multiple tables ',' split.  (default "*")
  -user string
        the database user (default "root")
  -file_type string 
        generate file type ,proto|csharp_service
  -ef_namespace string 
        csharp_service entity framework data namespace      
  -nameSpace string 
        csharp_service namespace       
```

```
$ rpcgenerate -db mysql -host localhost -user root -password 123456 -package pb -port 3306 -schema test -service_name tester  > test.proto
$ rpcgenerate -db mysql -host localhost -user root -password 123456 -package pb -port 3306 -schema test -service_name testService  -file_type csharp_service -table 'person' -ef_namespace "database.test" >test.cs 

```



#### Use as an imported library

```sh
$ go get -u github.com/matt-repository/rpcgenerate@latest
```

#### Thanks for 
    sql2pb : https://github.com/Mikaelemmmm/sql2pb
    schemabuf : https://github.com/mcos/schemabuf
#### Have a problem adding wechat: fq943609
