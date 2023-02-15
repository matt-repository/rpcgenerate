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
	protoServiceName := flag.String("proto_service_name", *schema+"er", "the proto service name , defaults to the database schema+'er'.")
	packageName := flag.String("package", *schema, "the protocol buffer package. defaults to the database schema.")
	ignoreTableStr := flag.String("ignore_tables", "", "a comma spaced list of tables to ignore")
	ignoreColumnStr := flag.String("ignore_columns", "", "a comma spaced list of mysql columns to ignore")
	fileType := flag.String("file_type", "proto", "generate file type ,proto|csharp_service")
	efNameSpace := flag.String("ef_namespace", "", "csharp_service entity framework data namespace")
	nameSpace := flag.String("nameSpace", "GrpcServices", "csharp_service namespace")

	flag.Parse()

	//test
	//*dbType = "sqlserver"
	//*host = "localhost"
	//*user = "root"
	//*schema = "123456"
	//*serviceName = "testservice"
	//*fileType = "proto"
	//*packageName = "testProto"
	//*port = 1433
	//*password = "123456"

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
	if *dbType == "sqlserver" {
		var conf []string
		conf = append(conf, "Provider=SQLOLEDB")
		conf = append(conf, fmt.Sprintf("Data Source=%s,1433", *host))  // sqlserver IP 和 服务器名称
		conf = append(conf, fmt.Sprintf("Initial Catalog=%s", *schema)) // 数据库名
		conf = append(conf, fmt.Sprintf("user id=%s", *user))           // 登陆用户名
		conf = append(conf, fmt.Sprintf("password=%s", *password))      // 登陆密码
		connStr = strings.Join(conf, ";")
	}
	db, err := sql.Open("adodb", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ignoreTables := strings.Split(*ignoreTableStr, ",")
	ignoreColumns := strings.Split(*ignoreColumnStr, ",")

	switch *fileType {
	case "proto":
		s, err := core.GenerateProto(db, *table, ignoreTables, ignoreColumns, *protoServiceName, *packageName, *dbType)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	case "csharp_service":
		if *efNameSpace == "" {
			fmt.Println(" - please input the ef namespace ")
			return
		}

		s, err := core.GenerateCSharpService(db, *table, ignoreTables, ignoreColumns, *serviceName, *protoServiceName, *packageName, *schema, *dbType, *nameSpace, *efNameSpace)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	}
}
