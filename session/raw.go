package session

import (
	"database/sql"
	"fmt"
	"geeORM/clause"
	"geeORM/dialect"
	"geeORM/log"
	"geeORM/schema"
	"strings"
)

// Session 结构体用于sql语句的生成和执行，是这个包中的关键
// 如果您想要执行sql语句必须通过Session的方法执行。
type Session struct {
	dialect  dialect.Dialect
	tx       *sql.Tx         //开启事务时不为nil
	refTable *schema.Schema  //当前会话中缓存的表信息
	db       *sql.DB         //数据库链接
	sql      strings.Builder //暂时存放sql语句
	sqlVars  []interface{}   //暂时存放sql语句对应的参数
	clause   clause.Clause   //用于为session生成sql语句
}

// CommonDB 是一个容纳*sql.db或者*sql.tx的接口
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// Clear 函数重置Session结构体的sql，sqlVars以及clause
// 为下一个sql语句的拼接和执行做准备
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		dialect: dialect,
		db:      db,
	}
}

// Raw 函数将sql语句和对应的参数直接放入Session中，一般是对
// clause.Build返回的参数调用此方法。
func (s *Session) Raw(sql string, sqlVars ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, sqlVars...)
	return s
}

// Exec 将当前Session中的sql以及sqlVars拼接并执行，在执行之后会清除sql信息
// 为下一个sql做准备。
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()

	log.Infof("real command : "+s.sql.String(), s.sqlVars...)
	fmt.Println(s.sqlVars)

	if s.sqlVars != nil && s.sqlVars[0] != nil {
		log.Info("Exec : ", s.sqlVars)
		if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
			log.Error(err.Error())
			return nil, err
		}
	} else {
		if result, err = s.DB().Exec(s.sql.String()); err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}
	return result, nil
}

// QueryRow 使用Session中的sql信息在数据库中查询单行结果
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 使用Session中的sql信息在数据库中查询多行结果
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err.Error())
	}
	return rows, nil
}

func (s *Session) GiveTx() *sql.Tx {
	return s.tx
}
