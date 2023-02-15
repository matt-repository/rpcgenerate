package core

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/matt-repository/rpcgenerate/tools/stringx"
	"strings"
)

func GenerateCSharpService(db *sql.DB, table string, ignoreTables, ignoreColumns []string, serviceName, pkg, schema, dbType, nameSpace, efNameSpace string) (*SchemaCSharp, error) {
	s := &SchemaCSharp{}
	dbs, err := dbSchema(db, dbType)
	if nil != err {
		return nil, err
	}
	s.NameSpace = nameSpace
	s.Schema = schema
	if serviceName != "Service" {
		s.ServiceName = serviceName
	} else {
		s.ServiceName = schema + "Service"
	}

	s.EFNameSpace = efNameSpace
	if "" != pkg {
		s.Package = pkg
	}

	cols, err := dbColumns(db, dbs, table, dbType)
	if nil != err {
		return nil, err
	}

	err = typesFromColumnsCSharp(s, cols, ignoreTables, ignoreColumns)

	return s, nil
}

func typesFromColumnsCSharp(s *SchemaCSharp, cols []Column, ignoreTables, ignoreColumns []string) error {
	messageMap := map[string]*MessageCSharp{}
	ignoreMap := map[string]bool{}
	ignoreColumnMap := map[string]bool{}
	for _, ig := range ignoreTables {
		ignoreMap[ig] = true
	}
	for _, ic := range ignoreColumns {
		ignoreColumnMap[ic] = true
	}
	for _, c := range cols {
		if _, ok := ignoreMap[c.TableName]; ok {
			continue
		}
		if _, ok := ignoreColumnMap[c.ColumnName]; ok {
			continue
		}

		messageName := stringx.From(c.TableName).ToCamel()

		//messageName = inflect.Singularize(messageName)

		msg, ok := messageMap[messageName]
		if !ok {
			messageMap[messageName] = &MessageCSharp{Name: messageName, Schema: s.Schema, EFNameSpace: s.EFNameSpace}
			msg = messageMap[messageName]
		}
		err := parseColumnCSharp(msg, c)
		if nil != err {
			return err
		}
	}
	for _, v := range messageMap {
		s.Messages = append(s.Messages, v)
	}

	return nil
}

// parseColumn parses a column and inserts the relevant fields in the Message. If an enumerated type is encountered, an Enum will
// be added to the Schema. Returns an error if an incompatible protobuf data type cannot be found for the database column type.
func parseColumnCSharp(msg *MessageCSharp, col Column) error {
	typ := strings.ToLower(col.DataType)
	var fieldType string
	fieldType = dataTypeConvert(typ)
	if "" == fieldType {
		return fmt.Errorf("no compatible protobuf type found for `%s`. column: `%s`.`%s`", col.DataType, col.TableName, col.ColumnName)
	}
	field := NewMessageFieldCSharp(fieldType, col.ColumnName, col.ColumnKey != "")

	err := msg.AppendFieldCSharp(field)
	if nil != err {
		return err
	}

	return nil
}

func dataTypeConvert(typ string) string {
	fieldType := ""
	switch typ {
	case "char", "nchar", "varchar", "text", "longtext", "mediumtext", "tinytext":
		fieldType = "string"
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		fieldType = "" +
			""
	case "date", "time", "datetime", "timestamp":
		//s.AppendImport("google/protobuf/timestamp.proto")
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

// AppendFieldCSharp appends a message field to a message. If the tag of the message field is in use, an error will be returned.
func (m *MessageCSharp) AppendFieldCSharp(mf MessageFieldCSharp) error {
	m.Fields = append(m.Fields, mf)
	return nil
}

// String returns a string representation of a Schema.
func (s *SchemaCSharp) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("using System;\n")
	buf.WriteString("using System.Collections.Generic;\n")
	buf.WriteString("using System.Linq;\n")
	buf.WriteString("using Grpc.Core;\n")
	buf.WriteString("using System.Threading.Tasks;\n")
	buf.WriteString(fmt.Sprintf("using %s;\n", s.EFNameSpace))
	buf.WriteString(fmt.Sprintf("using %s;\n", s.Package))
	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf("namespace %s\n", s.NameSpace))
	buf.WriteString("{\n")
	buf.WriteString(fmt.Sprintf("%s/// <summary>\n", indent))
	buf.WriteString(fmt.Sprintf("%s/// %s \n", indent, s.ServiceName))
	buf.WriteString(fmt.Sprintf("%s/// </summary>\n", indent))
	buf.WriteString(fmt.Sprintf("%spublic class %s :%s.%s.%sBase \n", indent, s.ServiceName, s.Package, s.ServiceName, s.ServiceName))
	buf.WriteString(fmt.Sprintf("%s{\n", indent))
	buf.WriteString(fmt.Sprintf("%sprivate readonly %sContext _%sContext;\n", indent2, s.Schema, s.Schema))
	buf.WriteString(fmt.Sprintf("%spublic %s(%sContext %sContext)\n", indent2, s.ServiceName, s.Schema, s.Schema))
	buf.WriteString(fmt.Sprintf("%s{\n", indent2))

	buf.WriteString(fmt.Sprintf("%s _%sContext=%sContext;\n", indent3, s.Schema, s.Schema))
	buf.WriteString(fmt.Sprintf("%s}\n", indent2))

	buf.WriteString("\n")

	for _, m := range s.Messages {
		buf.WriteString(fmt.Sprintf("%s// ------------------------------------ \n", indent2))
		buf.WriteString(fmt.Sprintf("%s//%s%sService\n", indent2, indent, m.Name))
		buf.WriteString(fmt.Sprintf("%s// ------------------------------------ \n", indent2))

		buf.WriteString("\n")
		m.GenRpcAddListCSharpService(buf)
		m.GenRpcEditCSharpService(buf)
		m.GenRpcDelCSharpService(buf)
		m.GenRpcGetPageListCSharpService(buf)
	}
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("%s}\n", indent))
	buf.WriteString("}")
	return buf.String()
}

// SchemaCSharp is a representation of a protobuf schema.
type SchemaCSharp struct {
	Package     string
	NameSpace   string
	ServiceName string
	Schema      string //数据库名
	EFNameSpace string
	Messages    []*MessageCSharp
}

type MessageCSharp struct {
	Name        string
	Fields      []MessageFieldCSharp
	Schema      string //数据库名
	EFNameSpace string
	Messages    []*MessageCSharp
}

// MessageFieldCSharp represents the field of a message.
type MessageFieldCSharp struct {
	Typ   string
	Name  string
	IsKey bool
}

// NewMessageFieldCSharp creates a new message field.
func NewMessageFieldCSharp(typ, name string, isKey bool) MessageFieldCSharp {
	return MessageFieldCSharp{typ, name, isKey}
}

func (m MessageCSharp) GenRpcAddListCSharpService(buf *bytes.Buffer) {
	m.rpcStart(buf, "AddList")

	buf.WriteString(fmt.Sprintf("%sif (request.%ss.Count==0)\n", indent3, m.Name))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))
	buf.WriteString(fmt.Sprintf("%sresult.Code = 201;\n", indent4))
	buf.WriteString(fmt.Sprintf("%sresult.Msg = \"Data cannot be empty\";\n", indent4))
	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(result);\n", indent4))
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))

	//赋值
	buf.WriteString(fmt.Sprintf("%stry\n", indent3))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))
	buf.WriteString(fmt.Sprintf("%sforeach (var item in request.%ss)\n", indent4, m.Name))
	buf.WriteString(fmt.Sprintf("%s{\n", indent4))

	buf.WriteString(fmt.Sprintf("%svar model = new %s.%s\n", indent5, m.EFNameSpace, m.Name))
	buf.WriteString(fmt.Sprintf("%s{\n", indent5))
	for _, v := range m.Fields {
		buf.WriteString(fmt.Sprintf("%s%s = item.%s,\n", indent6, v.Name, v.Name))
	}
	buf.WriteString(fmt.Sprintf("%s};\n", indent5))
	buf.WriteString(fmt.Sprintf("%s_%sContext.%s.Add(model);\n", indent5, m.Schema, m.Name))
	buf.WriteString(fmt.Sprintf("%s}\n", indent4))
	buf.WriteString(fmt.Sprintf("%s_%sContext.SaveChanges();\n", indent4, m.Schema))
	buf.WriteString(fmt.Sprintf("%sresult.Code = 200;\n", indent4))
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))
	buf.WriteString(fmt.Sprintf("%scatch (Exception e)\n", indent3))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))
	buf.WriteString(fmt.Sprintf("%sresult.Code = 201;\n", indent4))
	buf.WriteString(fmt.Sprintf("%sresult.Msg = e.Message;\n", indent4))
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))

	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(result);\n", indent3))
	buf.WriteString(fmt.Sprintf("%s}\n", indent2))
}
func (m MessageCSharp) GenRpcEditCSharpService(buf *bytes.Buffer) {
	m.rpcStart(buf, "Edit")
	editList := []string{}
	for _, v := range m.Fields {
		if v.IsKey {
			editList = append(editList, fmt.Sprintf("w.%s == request.%s", v.Name, v.Name))
		}
	}
	editStr := strings.Join(editList, "&&")
	buf.WriteString(fmt.Sprintf("%svar data = _%sContext.%s.FirstOrDefault(w => %s);\n", indent3, m.Schema, m.Name, editStr))
	buf.WriteString(fmt.Sprintf("%sif(data == null)\n", indent3))
	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(new Edit%sReply { Code = 201, Msg = \"Not Exist!\" });\n", indent4, m.Name))
	for _, v := range m.Fields {
		if v.Name != "Id" {
			buf.WriteString(fmt.Sprintf("%sdata.%s = request.%s;\n", indent3, v.Name, v.Name))
		}
	}
	buf.WriteString(fmt.Sprintf("%s_%sContext.SaveChanges();\n", indent3, m.Schema))
	buf.WriteString(fmt.Sprintf("%sresult.Code = 200;\n", indent3))
	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(result);\n", indent3))
	buf.WriteString(fmt.Sprintf("%s}\n", indent2))
}
func (m MessageCSharp) GenRpcDelCSharpService(buf *bytes.Buffer) {
	m.rpcStart(buf, "Del")

	delList := []string{}
	for _, v := range m.Fields {
		if v.IsKey {
			delList = append(delList, fmt.Sprintf("w.%s == request.%s", v.Name, v.Name))
		}
	}
	delStr := strings.Join(delList, "&&")
	buf.WriteString(fmt.Sprintf("%svar data = _%sContext.%s.FirstOrDefault(w =>%s);\n", indent3, m.Schema, m.Name, delStr))
	buf.WriteString(fmt.Sprintf("%sif(data == null)\n", indent3))
	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(new Del%sReply { Code = 201, Msg = \"Not Exist!\" });\n", indent4, m.Name))
	buf.WriteString(fmt.Sprintf("%s_%sContext.%s.Remove(data);\n", indent3, m.Schema, m.Name))
	buf.WriteString(fmt.Sprintf("%s_%sContext.SaveChanges();\n", indent3, m.Schema))
	buf.WriteString(fmt.Sprintf("%sresult.Code = 200;\n", indent3))
	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(result);\n", indent3))
	buf.WriteString(fmt.Sprintf("%s}\n", indent2))
}
func (m MessageCSharp) GenRpcGetPageListCSharpService(buf *bytes.Buffer) {
	m.rpcStart(buf, "GetPageList")
	buf.WriteString(fmt.Sprintf("%svar query = _%sContext.%s.AsQueryable();\n", indent3, m.Schema, m.Name))
	buf.WriteString(fmt.Sprintf("%sif (request.Wheres != null)\n", indent3))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))
	for _, v := range m.Fields {
		switch v.Typ {
		case "string":
			buf.WriteString(fmt.Sprintf("%sif (!string.IsNullOrEmpty(request.Wheres.%s))\n", indent4, v.Name))
		default:
			buf.WriteString(fmt.Sprintf("%sif (request.Wheres.%s > 0)\n", indent4, v.Name))
		}
		buf.WriteString(fmt.Sprintf("%s{\n", indent4))
		buf.WriteString(fmt.Sprintf("%squery = query.Where(w => w.%s == request.Wheres.%s);\n", indent5, v.Name, v.Name))
		buf.WriteString(fmt.Sprintf("%s}\n", indent4))
	}
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))
	//count

	buf.WriteString(fmt.Sprintf("%sresult.Total = query.Count();\n", indent3))

	//分页
	buf.WriteString(fmt.Sprintf("%sif (request.Pagings != null)\n", indent3))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))
	buf.WriteString(fmt.Sprintf("%squery = query.Skip((request.Pagings.PageIndex - 1) * request.Pagings.PageSize).Take(request.Pagings.PageSize);\n", indent4))
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))

	buf.WriteString(fmt.Sprintf("%svar list = query.ToList();\n", indent3))

	//赋值
	buf.WriteString(fmt.Sprintf("%sforeach (var item in list)\n", indent3))
	buf.WriteString(fmt.Sprintf("%s{\n", indent3))

	buf.WriteString(fmt.Sprintf("%svar model = new %sProto.%s\n", indent4, m.Name, m.Name))
	buf.WriteString(fmt.Sprintf("%s{\n", indent4))
	for _, v := range m.Fields {
		buf.WriteString(fmt.Sprintf("%s%s = item.%s,\n", indent5, v.Name, v.Name))
	}
	buf.WriteString(fmt.Sprintf("%s};\n", indent4))
	buf.WriteString(fmt.Sprintf("%sresult.%ss.Add(model);\n", indent4, m.Name))
	buf.WriteString(fmt.Sprintf("%s}\n", indent3))

	buf.WriteString(fmt.Sprintf("%sreturn Task.FromResult(result);\n", indent3))
	buf.WriteString(fmt.Sprintf("%s}\n", indent2))
}

func (m MessageCSharp) rpcStart(buf *bytes.Buffer, funcType string) {
	funcName := funcType + m.Name
	request := funcName + "Request"
	reply := funcName + "Reply"

	buf.WriteString(fmt.Sprintf("%spublic override Task<%s> %s(%s request, ServerCallContext context)\n", indent2, reply, funcName, request))
	buf.WriteString(fmt.Sprintf("%s{\n", indent2))
	buf.WriteString(fmt.Sprintf("%svar result= new %sReply();\n", indent3, funcName))
}
