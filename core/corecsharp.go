package core

import (
	"fmt"
	_map "github.com/matt-repository/golib/map"
	"github.com/matt-repository/rpcgenerate/tools/stringx"
	"sort"
)

const (
	nameSpace = "GrpcServices"
)

// SchemaCSharp ...
type SchemaCSharp struct {
	Schema      string
	NameSpace   string
	ServiceName string
	EFNameSpace string
	Imports     sort.StringSlice
	Messages    []*MessageTableCSharp
}

// MessageTableCSharp ...
type MessageTableCSharp struct {
	Name      string
	Fields    []MessageFieldCSharp
	PriFields []MessageFieldCSharp
	Messages  []*MessageTableCSharp
}

// MessageFieldCSharp ...
type MessageFieldCSharp struct {
	Typ  string
	Name string
}

func NewCSharpSchema(schema string) *SchemaCSharp {
	return &SchemaCSharp{
		NameSpace:   nameSpace,
		ServiceName: schema + "Service",
		Schema:      schema,
		EFNameSpace: "database." + schema,
	}
}

// typesFromColumns ...
func (s *SchemaCSharp) typesFromColumns(cols []Column, ignoreTableMap map[string]bool) error {
	messageMap := map[string]*MessageTableCSharp{}
	for _, col := range cols {
		if _, ok := ignoreTableMap[col.TableName]; ok {
			continue
		}
		messageName := stringx.From(col.TableName).ToCamel()
		//messageName = inflect.Singularize(messageName)
		m, ok := messageMap[messageName]
		if !ok {
			messageMap[messageName] = &MessageTableCSharp{
				Name: messageName,
			}
			m = messageMap[messageName]
		}

		fieldType := col.dataTypeConvertCSharp()
		if fieldType == "" {
			return fmt.Errorf("no compatible protobuf type found for `%s`. column: `%s`.`%s`", col.DataType, col.TableName, col.ColumnName)
		}
		err := m.parseColumn(col, fieldType)
		if err != nil {
			return err
		}
	}
	s.Messages = _map.ToSlice(messageMap, func(k string, v *MessageTableCSharp) *MessageTableCSharp {
		return v
	})
	return nil
}

// parseColumn ...
func (m *MessageTableCSharp) parseColumn(col Column, fieldType string) error {
	field := MessageFieldCSharp{
		Typ:  fieldType,
		Name: col.ColumnName,
	}
	m.Fields = append(m.Fields, field)
	if col.ColumnKey == "PRI" {
		m.PriFields = append(m.PriFields, field)
	}

	return nil
}

// appendImport ...
func (s *SchemaCSharp) appendImport() {
	s.Imports = []string{
		"System",
		"System.Collections.Generic",
		"System.Linq",
		"Grpc.Core",
		"System.Threading.Tasks",
	}
	s.Imports = append(s.Imports, s.EFNameSpace)
	s.Imports = append(s.Imports, s.Schema)
	sort.Sort(s.Imports)
}

// ExecTemplate ...
func (s *SchemaCSharp) ExecTemplate() error {
	return CsharpExecTemplate(fmt.Sprintf("./%s.cs", s.Schema), s)
}
