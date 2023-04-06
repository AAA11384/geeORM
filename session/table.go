package session

import (
	"fmt"
	"geeORM/log"
	"geeORM/schema"
	"reflect"
	"strings"
)

// Model 传入结构体指针，在session的scheme中为其建立对应的模型。
func (s *Session) Model(value interface{}) *Session {
	// 如果是nil或者不同结构体才进行更新，用于缓存解析的数据
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// RefTable 获取当前session中的表信息。
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

// CreateTable 传入结构体指针，在数据库创建对应的table。
func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("`%s` %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE `%s` (%s);", table.Name, desc), nil).Exec()
	if err != nil {
		return err
	}
	return nil
}

// DropTable 删除数据库中当前session model中记录的表。
func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

// HasTable 判断当前session model的表在数据库中是否已经存在。
func (s *Session) HasTable(databaseName string) bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	values = append(values, databaseName)
	fmt.Println(values[0].(string), values[1].(string))
	fmt.Println(sql)

	row := s.Raw(sql, values[0].(string), databaseName).QueryRow()
	var tmp int64
	err := row.Scan(&tmp)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	return tmp == 1

}
