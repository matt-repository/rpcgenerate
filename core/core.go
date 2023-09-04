package core

import (
	"database/sql"
	"github.com/matt-repository/rpcgenerate/tools/stringx"
	"log"
	"strings"
)

var (
	mysql                 = "mysql"
	sqlserver             = "sqlserver"
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

func dbColumns(db *sql.DB, schema, table, dbType string) ([]Column, error) {
	tableArr := strings.Split(table, ",")
	q := strings.Builder{}
	switch dbType {
	case mysql:
		q.WriteString(mysqlGetRowsSQl)
	case sqlserver:
		q.WriteString(sqlServerGetRowsSQl)
	}
	if table != "" && table != "*" {
		tableStr := strings.Join(tableArr, "','")
		q.WriteString(" AND c.TABLE_NAME IN('" + tableStr + "')")
	}
	q.WriteString(" ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION ")
	rows, err := db.Query(q.String(), schema)
	defer rows.Close()
	if nil != err {
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
