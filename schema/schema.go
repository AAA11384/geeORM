package schema

import (
	"geeORM/dialect"
	"go/ast"
	"reflect"
)

// Field 结构体表示一个字段，也表示其对应的一个列
// Tag是这个列的约束条件
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema 结构体表示一个table，Model表示的是表对应的结构体指针
// 的reflect.Value的值，即传入时最原始的值。
type Schema struct {
	Model      interface{}       //结构体指针
	Name       string            //表名称
	fieldMap   map[string]*Field //字段map
	Fields     []*Field          //字段slice
	FieldNames []string          //所有字段名称
}

// GetField 使用scheme的map根据field name获取 *Field。
func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// Parse 函数使用一个结构体的指针获得该结构体通过反射解析
// 获得对应的表的Schema结构体的指针。
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	//reflect.ValueOf返回接口中数据的值
	//reflect.Indirect获取指针指向的实例
	//Indirect函数用于对指针的变量使用reflect.ValueOf函数得到的Value结构体
	//TypeOf()/Type入参的类型
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model: dest,
		//modelType.Name() 获取到结构体的名称作为表名
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	//NumField() 获取实例的字段的个数
	for i := 0; i < modelType.NumField(); i++ {
		//通过下标获取到特定字段
		p := modelType.Field(i)
		//是嵌入式字段并且对外暴露
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				//通过p的类型通过d进行转化
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			//获取tag的值
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

// RecordValues 传入结构体实例的指针，解析非空值的字段值，以接口切片返回。
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
