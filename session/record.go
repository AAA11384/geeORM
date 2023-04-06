package session

import (
	"errors"
	"geeORM/clause"
	"geeORM/log"
	"reflect"
)

// Insert 传入表对应的结构体的实例的指针，会通过反射获取信息
// 并生成sql语句将信息插入到表中。不可以在之前使用其它参数，
// 此函数会最终执行sql语句。
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		//对这个结构体建立模型,并获取表结构
		//注意，此处必须填写第二个参数，即使已经预设了&User为model
		//因为我们要对每一个参数执行它的钩子函数，这其实是一个方法，函数的第一个参数必须是自身，即实例本身的指针
		//如果是一个结构体模型，其内部值为0，并不会改变每个参数中的值
		table := s.Model(value).RefTable()
		s.CallMethod(BeforeInsert, value)
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		//获取各字段的值
		recordValues = append(recordValues, table.RecordValues(value))
	}

	//设置参数
	s.clause.Set(clause.VALUES, recordValues...)
	//获取sqlStr以及参数
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	//执行语句
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

// Find 传入结构体切片的指针，自动将信息填充到切片中。可以在前方使用Where，OrderBy以及Limit参数，
// 此函数会最终执行sql语句。
func (s *Session) Find(values ...interface{}) error {
	s.CallMethod(BeforeQuery, nil)
	//获取当前values的值
	destSlice := reflect.Indirect(reflect.ValueOf(values[0]))
	//获取数组中元素的类型。Type.Elem函数当Type不为 Array, Chan, Map, Pointer, Slice时panic。
	destType := destSlice.Type().Elem()
	//reflect.New使用类型创建一个value。 value.Elem获取指针的值或者接口的动态类型。
	//.interface()  将一个value转为interface包含的值
	//需要将 []user -> *User
	//这里使用了reflect.New(destType).Elem() 确保传入的是一个值 而不是指针，
	//尽管Parse会使用reflect.Indirect，是为了保证统一性
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	//拼接sql语句和参数
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	log.Info(sql, vars)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	//将查询结果分别赋值给参数
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		//将字段展开
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		//获取结果
		if err := rows.Scan(values...); err != nil {
			return err
		}
		//执行钩子函数
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		//将值存入参数结构体
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// Update 函数接收一个map或者多个kv对，修改当前Session中
// Schema指定的的表一个或者多个字段的值，k是字段名，v是期望修改之后的值。
// 可以结合Where参数。此函数会执行sql语句。
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.refTable.Name, m)
	sqlStr, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	res, err := s.Raw(sqlStr, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return res.RowsAffected()
}

// Delete 函数删除Session.Schema指向的表中符合条件的行
// 可以在前方使用Where参数，此函数会执行sql语句。
func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.refTable.Name)
	sqlStr, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	res, err := s.Raw(sqlStr, vars...).Exec()
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return res.RowsAffected()
}

// Count 函数统计Session.Schema指定的表中符合条件的行数
// 可以在前方使用Where参数，此函数会执行sql语句。
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.refTable.Name)
	sqlStr, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sqlStr, vars...).QueryRow()
	var tem int64
	if err := row.Scan(&tem); err != nil {
		log.Error(err.Error())
		return 0, err
	}
	return tem, nil
}

// Limit 函数设置Limit参数，此函数不会执行sql语句
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Where 设置where参数，此函数不会执行sql语句
// Example:
// s.Where("age >= ?", "18").Exec()
func (s *Session) Where(sqlStr string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, sqlStr), args...)...)
	return s
}

// OrderBy 设置顺序参数，此函数不会执行sql语句。
func (s *Session) OrderBy(field string) *Session {
	s.clause.Set(clause.ORDERBY, field)
	return s
}

// First 传入结构体实例的指针，向其中填入第一个符合条件的值。
// 此函数会执行sql语句
func (s *Session) First(value interface{}) error {
	//将参数转化为值
	dest := reflect.Indirect(reflect.ValueOf(value))
	//按照值取得类型，新建该类型的切片 reflect.SliceOf(Type)
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	//使用Limit和Find 设置数量并执行 destSlice.Addr().Interface() Find的参数是地址的interface
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	//将第一个结果写回dest
	dest.Set(destSlice.Index(0))
	return nil
}
