package core

import (
	"fmt"
	_map "github.com/matt-repository/golib/map"
	"github.com/matt-repository/golib/slice"
	"github.com/matt-repository/rpcgenerate/tools/stringx"
	"regexp"
	"sort"
	"strings"
)

const (
	proto3 = "proto3"
)

// SchemaProto ...
type SchemaProto struct {
	Syntax      string
	ServiceName string
	Schema      string
	Imports     sort.StringSlice
	Messages    []*MessageProto
	Enums       EnumCollection
}

// EnumCollection ...
type EnumCollection []*Enum

func (ec EnumCollection) Len() int {
	return len(ec)
}

func (ec EnumCollection) Less(i, j int) bool {
	return ec[i].Name < ec[j].Name
}

func (ec EnumCollection) Swap(i, j int) {
	ec[i], ec[j] = ec[j], ec[i]
}

// Enum ...
type Enum struct {
	Name    string
	Comment string
	Fields  []EnumField
}

// EnumField ...
type EnumField struct {
	Name string
	Tag  int
}

// NewEnumField ...
func NewEnumField(name string, tag int) EnumField {
	name = strings.ToUpper(name)
	re := regexp.MustCompile(`(\W+)`)
	name = re.ReplaceAllString(name, "_")
	return EnumField{name, tag}
}

// MessageProto ...
type MessageProto struct {
	Name      string
	Comment   string
	Fields    []MessageField
	PriFields []MessageField
	Messages  []*MessageProto
}

// MessageFieldConvert ...
func (m *MessageProto) MessageFieldConvert() []MessageField {
	var filedTag int
	curFields := make([]MessageField, 0)
	notCoverList := []string{"version", "del_state", "delete_time"}
	for _, field := range m.Fields {
		isExist := slice.Exists(notCoverList, func(s []string, i int) bool { return s[i] == field.Name })
		if isExist {
			continue
		}
		filedTag++
		field.Tag = filedTag
		field.Name = stringx.From(field.Name).ToCamelWithStartLower()
		if field.IsNull {
			switch field.Typ {
			case "string":
				field.Typ = "google.protobuf.StringValue"
			case "double":
				field.Typ = "google.protobuf.DoubleValue"
			case "int64":
				field.Typ = "google.protobuf.Int64Value"
			case "int32":
				field.Typ = "google.protobuf.Int32Value"
			}
		}
		if field.Comment == "" {
			field.Comment = field.Name
		}
		curFields = append(curFields, field)
	}
	return curFields
}

// MessageField ...
type MessageField struct {
	Typ     string
	Name    string
	Tag     int
	Comment string
	IsKey   bool
	IsNull  bool
}

// newEnumFromStrings ...
func newEnumFromStrings(name, comment string, fields []string) (*Enum, error) {
	enum := &Enum{
		Name:    name,
		Comment: comment,
	}
	for i, field := range fields {
		eField := NewEnumField(field, i)
		isExist := slice.Exists(enum.Fields, func(s []EnumField, i int) bool {
			return s[i].Tag == eField.Tag
		})
		if !isExist {
			enum.Fields = append(enum.Fields, eField)
		}
	}

	return enum, nil
}

func NewProtoSchema(schema string) *SchemaProto {
	return &SchemaProto{
		Syntax:      proto3,
		ServiceName: schema + "Service",
		Schema:      schema,
	}
}

// typesFromColumns ...
func (s *SchemaProto) typesFromColumns(cols []Column, ignoreTables map[string]bool) error {
	messageMap := map[string]*MessageProto{}
	for _, c := range cols {
		if _, ok := ignoreTables[c.TableName]; ok {
			continue
		}
		//驼峰
		messageName := stringx.From(c.TableName).ToCamel()
		//messageName := snaker.SnakeToCamel(c.TableName)
		//messageName = inflect.Singularize(messageName)
		msg, ok := messageMap[messageName]
		if !ok {
			msg = &MessageProto{Name: messageName, Comment: c.TableComment}
			messageMap[messageName] = msg
		}
		fieldType := c.dataTypeConvertProto(s)
		if fieldType == "" {
			return fmt.Errorf("no compatible protobuf type found for `%s`. column: `%s`.`%s`", c.DataType, c.TableName, c.ColumnName)
		}
		err := msg.parseColumn(c, fieldType)
		if err != nil {
			return err
		}
	}
	sort.Sort(s.Enums)
	s.Messages = _map.ToSlice(messageMap, func(k string, v *MessageProto) *MessageProto {
		return v
	})
	return nil
}

// parseColumn ...
func (m *MessageProto) parseColumn(col Column, fieldType string) error {
	field := MessageField{
		Typ:     fieldType,
		Name:    col.ColumnName,
		Tag:     len(m.Fields) + 1,
		Comment: col.ColumnComment,
		IsKey:   col.ColumnKey == "PRI",
		IsNull:  col.IsNullable == "YES",
	}
	m.Fields = append(m.Fields, field)
	if col.ColumnKey == "PRI" {
		m.PriFields = append(m.PriFields, field)
	}

	return nil
}

// appendImport ...
func (s *SchemaProto) appendImport() {
	s.Imports = []string{
		"google/protobuf/wrappers.proto",
	}
	sort.Sort(s.Imports)
}

// ExecTemplate ...
func (s *SchemaProto) ExecTemplate() error {
	return ProtoExecTemplate(fmt.Sprintf("./%s.proto", s.Schema), s)
}
