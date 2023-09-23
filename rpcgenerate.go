package main

import (
	"github.com/matt-repository/rpcgenerate/config"
	"github.com/matt-repository/rpcgenerate/core"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-adodb"
	"log"
)

func main() {
	err := config.InitConfig()
	defer config.GetConfig().Db.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	var schemer core.Schemer
	switch config.GetConfig().GenType {
	case config.Proto:
		schemer = core.NewProtoSchema(config.GetConfig().Schema)
	case config.CsharpService:
		schemer = core.NewCSharpSchema(config.GetConfig().Schema)
	}
	err = core.Generate(schemer)
	if err != nil {
		log.Fatal(err)
	}
}
