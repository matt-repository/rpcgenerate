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
	serviceName := flag.String("service_name", *schema, "the protobuf service name , defaults to the database schema.")
	packageName := flag.String("package", *schema, "the protocol buffer package. defaults to the database schema.")
	ignoreTableStr := flag.String("ignore_tables", "", "a comma spaced list of tables to ignore")
	ignoreColumnStr := flag.String("ignore_columns", "", "a comma spaced list of mysql columns to ignore")
	fieldStyle := flag.String("field_style", "sqlPb", "gen protobuf field style, sql_pb | sqlPb")
	fileType := flag.String("file_type", "proto", "generate file type ,proto|csharp_service")

	flag.Parse()

	//test
	//*dbType = "sqlserver"
	//*host = "localhost"
	//*user = "sa"
	//*schema = "MattTest"
	//*serviceName = "MattTestservice"
	//*fileType = "proto"
	//*packageName = "MattTestProto"
	//*port = 1433
	//*password = "123456"
	//

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
		s, err := core.GenerateProto(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *fieldStyle, *dbType)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	case "csharp_service":
		s, err := core.GenerateCSharpService(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *fieldStyle, *schema, *dbType)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	}
}
