package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"rpcgenerate/core"
	"strings"

	_ "github.com/go-sql-driver/mysql"
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
	fileType := flag.String("file_type", "proto", "generate file type ,proto|c#_service")

	flag.Parse()

	if *schema == "" {
		fmt.Println(" - please input the database schema ")
		return
	}

	if *fileType != "proto" && *fileType != "c#_service" {
		fmt.Println(" - please input fileType proto|c#_service")
		return
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", *user, *password, *host, *port, *schema)
	db, err := sql.Open(*dbType, connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ignoreTables := strings.Split(*ignoreTableStr, ",")
	ignoreColumns := strings.Split(*ignoreColumnStr, ",")

	switch *fileType {
	case "proto":
		s, err := core.GenerateProto(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *fieldStyle)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	case "c#_service":
		s, err := core.GenerateCSharpService(db, *table, ignoreTables, ignoreColumns, *serviceName, *packageName, *fieldStyle)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	}
}
