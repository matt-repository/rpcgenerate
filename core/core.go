package core

import (
	"database/sql"
	"github.com/chuckpreslar/inflect"
	"github.com/matt-repository/rpcgenerate/config"
	"github.com/matt-repository/rpcgenerate/tools/stringx"
	"github.com/serenize/snaker"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var (
	mysql     = "mysql"
	sqlserver = "sqlserver"

	mysqlGetSchemaSQL     = "SELECT DATABASE()"
	sqlServerGetSchemaSQL = "Select top 1 Name From Master..SysDataBases Where DbId=(Select Dbid From Master..SysProcesses Where Spid = @@spid)"
	mysqlGetRowsSQl       = "SELECT c.TABLE_NAME, c.COLUMN_NAME, c.IS_NULLABLE, c.DATA_TYPE,c.CHARACTER_MAXIMUM_LENGTH, c.NUMERIC_PRECISION, c.NUMERIC_SCALE, c.COLUMN_TYPE ,c.COLUMN_COMMENT,t.TABLE_COMMENT,c.COLUMN_KEY " +
		"FROM INFORMATION_SCHEMA.COLUMNS as c " +
		"LEFT JOIN  INFORMATION_SCHEMA.TABLES as t on c.TABLE_NAME = t.TABLE_NAME and  c.TABLE_SCHEMA = t.TABLE_SCHEMA " +
		"WHERE c.TABLE_SCHEMA = ?"
	sqlServerGetRowsSQl = "SELECT c.TABLE_NAME, c.COLUMN_NAME, c.IS_NULLABLE, c.DATA_TYPE, c.CHARACTER_MAXIMUM_LENGTH, c.NUMERIC_PRECISION, c.NUMERIC_SCALE,  c.Data_TYPE AS COLUMN_TYPE,'' as COLUMN_COMMENT,'' as TABLE_COMMENT," +
		"'COLUMN_KEY'= CASE  WHEN  d.COLUMN_NAME is null THEN '' ELSE 'PRI' end " +
		"FROM INFORMATION_SCHEMA.COLUMNS as c " +
		"LEFT JOIN INFORMATION_SCHEMA.TABLES as t  on c.TABLE_NAME = t.TABLE_NAME and  c.TABLE_SCHEMA = t.TABLE_SCHEMA " +
		"left join INFORMATION_SCHEMA.KEY_COLUMN_USAGE D on c.TABLE_NAME = D.TABLE_NAME and c.COLUMN_NAME=d.COLUMN_NAME " +
		"WHERE c.TABLE_CATALOG = ?"
)

func dbSchema(db *sql.DB, dbType string) (string, error) {
	var schema string
	switch dbType {
	case mysql:
		mysqlStr := mysqlGetSchemaSQL
		err := db.QueryRow(mysqlStr).Scan(&schema)
		return schema, err
	case sqlserver:
		err := db.QueryRow(sqlServerGetSchemaSQL).Scan(&schema)
		return schema, err
	}
	return schema, nil
}

// dbColumns ...
func dbColumns(c *config.Config) ([]Column, error) {
	q := strings.Builder{}
	switch c.DbType {
	case mysql:
		q.WriteString(mysqlGetRowsSQl)
	case sqlserver:
		q.WriteString(sqlServerGetRowsSQl)
	}
	if len(c.Tables) > 0 {
		tableStr := strings.Join(c.Tables, "','")
		q.WriteString(" AND c.TABLE_NAME IN('" + tableStr + "')")
	}
	q.WriteString(" ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION ")
	rows, err := c.Db.Query(q.String(), c.Schema)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	var cols []Column
	for rows.Next() {
		cs := Column{}
		err = rows.Scan(&cs.TableName, &cs.ColumnName, &cs.IsNullable, &cs.DataType, &cs.CharacterMaximumLength, &cs.NumericPrecision, &cs.NumericScale, &cs.ColumnType, &cs.ColumnComment, &cs.TableComment, &cs.ColumnKey)
		if err != nil {
			log.Fatal(err)
		}
		if cs.TableComment == "" {
			cs.TableComment = stringx.From(cs.TableName).ToCamelWithStartLower()
		}
		cols = append(cols, cs)
	}
	if err = rows.Err(); nil != err {
		return nil, err
	}
	return cols, nil
}

// Column ...
type Column struct {
	Style                  string
	TableName              string
	TableComment           string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnComment          string
	ColumnKey              string
}

// dataTypeConvertCSharp ...
func (c *Column) dataTypeConvertCSharp() string {
	typ := strings.ToLower(c.DataType)
	var fieldType string
	switch typ {
	case "char", "nchar", "varchar", "text", "longtext", "mediumtext", "tinytext":
		fieldType = "string"
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		fieldType = "" + ""
	case "date", "time", "datetime", "timestamp":
		fieldType = "int64"
	case "bool", "bit":
		fieldType = "bool"
	case "tinyint", "smallint", "mediumint", "int":
		fieldType = "int32"
	case "bigint":
		fieldType = "int64"
	case "float", "decimal", "double":
		fieldType = "double"
	case "json":
		fieldType = "string"
	}
	return fieldType
}

// dataTypeConvertProto ...
func (c *Column) dataTypeConvertProto(s *SchemaProto) string {
	typ := strings.ToLower(c.DataType)
	var fieldType string
	switch typ {
	case "char", "nchar", "varchar", "text", "longtext", "mediumtext", "tinytext":
		fieldType = "string"
	case "enum", "set":
		// Parse c.ColumnType to get the enum list
		enumList := regexp.MustCompile(`[enum|set]\((.+?)\)`).FindStringSubmatch(c.ColumnType)
		enums := strings.FieldsFunc(enumList[1], func(c rune) bool {
			cs := string(c)
			return "," == cs || "'" == cs
		})
		enumName := inflect.Singularize(snaker.SnakeToCamel(c.TableName)) + snaker.SnakeToCamel(c.ColumnName)
		enum, err := newEnumFromStrings(enumName, c.ColumnComment, enums)
		if err != nil {
			return ""
		}
		s.Enums = append(s.Enums, enum)
		fieldType = enumName
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		fieldType = "bytes"
	case "date", "time", "datetime", "timestamp":
		//s.appendImport("google/protobuf/timestamp.proto")
		fieldType = "int64"
	case "bool", "bit":
		fieldType = "bool"
	case "tinyint", "smallint", "mediumint", "int":
		fieldType = "int32"
	case "bigint":
		fieldType = "int64"
	case "float", "decimal", "double":
		fieldType = "double"
	case "json":
		fieldType = "string"
	}

	return fieldType
}

// ExecTemplate ...
func ExecTemplate(srcTmpl, destPath string, data interface{}) error {
	t, err := template.ParseFiles(srcTmpl)
	if err != nil {
		return err
	}
	var _file *os.File
	defer _file.Close()
	_file, err = os.Create(destPath)
	if err != nil {
		return err
	}
	err = t.Execute(_file, data)
	return err
}

// ProtoExecTemplate ...
func ProtoExecTemplate(destPath string, data interface{}) error {
	return ExecTemplate("./templates/proto.tmpl", destPath, data)
}

// CsharpExecTemplate ...
func CsharpExecTemplate(destPath string, data interface{}) error {
	return ExecTemplate("./templates/csharp.tmpl", destPath, data)
}

type Schemer interface {
	typesFromColumns(cols []Column, ignoreTableMap map[string]bool) error
	appendImport()
	ExecTemplate() error
}

func Generate(s Schemer) error {
	s.appendImport()
	cols, err := dbColumns(config.GetConfig())
	if err != nil {
		return err
	}
	err = s.typesFromColumns(cols, config.GetConfig().IgnoreTableMap)
	if err != nil {
		return err
	}
	return s.ExecTemplate()
}
