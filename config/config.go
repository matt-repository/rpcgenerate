package config

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/matt-repository/golib/slice"
	"strings"
)

type GenType = string

const (
	Proto         GenType = "proto"
	CsharpService GenType = "csharp_service"
)

type DbType = string

const (
	Mysql     DbType = "mysql"
	Sqlserver DbType = "sqlserver"
)

const (
	SqlServerDriver = "adodb"
)

type Config struct {
	DbType         DbType
	Schema         string
	GenType        GenType
	Conn           string
	Tables         []string
	IgnoreTableMap map[string]bool
	Db             *sql.DB
}

var config *Config

func GetConfig() *Config {
	return config
}

func InitConfig() error {
	// " root@123456@tcp(localhost:3306)/test"
	//  "Provider=SQLOLEDB;Data Source=localhost,3306;Initial Catalog=test;user id=admin;password=%123456"
	dbType := flag.String("db", Mysql, "the database type")
	schema := flag.String("schema", "", "the database schema")
	genType := flag.String("type", "proto", "generate file type ,proto|csharp_service")
	conn := flag.String("conn", "", "the connect string")
	table := flag.String("table", "", "the table schema，multiple tables ',' split. ")
	ignoreTable := flag.String("ignore_table", "", "a comma spaced list of tables to ignore")
	flag.Parse()

	if *dbType != Mysql && *dbType != Sqlserver {
		return fmt.Errorf(" - please input the database type mysql|sqlserver ")
	}

	if *schema == "" {
		return fmt.Errorf(" - please input the database schema ")
	}

	if *genType != Proto && *genType != CsharpService {
		return fmt.Errorf(" - please input type proto|csharp_service")
	}

	if *conn == "" {
		return fmt.Errorf(" - please input the database connect string ")
	}

	var ignoreTableMap map[string]bool
	var tables []string
	if *ignoreTable != "" {
		// ignore tables map
		ignoreTables := strings.Split(*ignoreTable, ",")
		ignoreTableMap = slice.ToMap(ignoreTables, func(v string) (key string, value bool) {
			return v, true
		})
	}
	if *table != "" {
		tables = strings.Split(*table, ",")
	}

	var driveName string
	switch *dbType {
	case Mysql:
		driveName = Mysql
	case Sqlserver:
		driveName = SqlServerDriver
	}
	// database
	db, err := sql.Open(driveName, *conn)
	if err != nil {
		return err
	}

	config = &Config{
		DbType:         *dbType,
		Schema:         *schema,
		GenType:        *genType,
		Conn:           *conn,
		Tables:         tables,
		IgnoreTableMap: ignoreTableMap,
		Db:             db,
	}
	return nil

}
