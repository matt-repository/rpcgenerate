package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"rpcgenerate/core"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-adodb"
)

func main() {
	dbType := flag.String("db", "mysql", "the database type")
	host := flag.String("host", "localhost", "the database host")
	port := flag.Int("port", 3306, "the database port")
	user := flag.String("user", "root", "the database user")
	password := flag.String("password", "root", "the database password")
	schema := flag.String("schema", "", "the database schema")
	table := flag.String("table", "*", "the table schema，multiple tables ',' split. ")
	serviceName := flag.String("service_name", *schema+"Service", "the service name , defaults to the database schema+'Service'.")
	packageName := flag.String("package", *schema, "the protocol buffer package. defaults to the database schema.")
	ignoreTableStr := flag.String("ignore_tables", "", "a comma spaced list of tables to ignore")
	ignoreColumnStr := flag.String("ignore_columns", "", "a comma spaced list of mysql columns to ignore")
	fileType := flag.String("file_type", "proto", "generate file type ,proto|csharp_service")
	efNameSpace := flag.String("ef_namespace", "", "csharp_service entity framework data namespace")
	nameSpace := flag.String("nameSpace", "GrpcServices", "csharp_service namespace")
	flag.Parse()

	//test
	*dbType = "sqlserver"
	*host = "192.168.1.33"
	*user = "sa"
	*schema = "efosbasicsys"
	*serviceName = "AApier"
	*fileType = "proto"
	*packageName = "AApiProto"
	*port = 1433
	*password = "Hietech123"

	if *schema == "" {
		fmt.Println(" - please input the database schema ")
		return
	}

	if *fileType != "proto" && *fileType != "csharp_service" {
		fmt.Println(*fileType)
		fmt.Println(" - please input fileType proto|c#_service")
		return
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", *user, *password, *host, *port, *schema)
	driName := "mysql"
	if *dbType == "sqlserver" {
		var conf []string
		conf = append(conf, "Provider=SQLOLEDB")
		conf = append(conf, fmt.Sprintf("Data Source=%s,%v", *host, *port)) // sqlserver IP 和 服务器名称
		conf = append(conf, fmt.Sprintf("Initial Catalog=%s", *schema))     // 数据库名
		conf = append(conf, fmt.Sprintf("user id=%s", *user))               // 登陆用户名
		conf = append(conf, fmt.Sprintf("password=%s", *password))          // 登陆密码
		connStr = strings.Join(conf, ";")
		driName = "adodb"
	}
	db, err := sql.Open(driName, connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ignoreTables := strings.Split(*ignoreTableStr, ",")
	ignoreColumns := strings.Split(*ignoreColumnStr, ",")

	switch *fileType {
	case "proto":
		s, err := core.GenerateProto(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *dbType)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	case "csharp_service":
		if *efNameSpace == "" {
			fmt.Println(" - please input the entity framework namespace ")
			return
		}

		s, err := core.GenerateCSharpService(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *schema, *dbType, *nameSpace, *efNameSpace)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	}
}
