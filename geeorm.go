package geeORM

import (
	"database/sql"
	"errors"
	"geeORM/dialect"
	"geeORM/log"
	"geeORM/session"

	_ "github.com/go-sql-driver/mysql"
)

// Engine 结构体表示一个引擎，使用NewEngine函数创建的Engine
// 是连接数据库取得初始化的，可以通过其NewSession方法得到Session。
type Engine struct {
	s       *sql.DB
	dialect dialect.Dialect
}

// NewEngine 函数输入要操作的数据库类型和链接指令，取得链接后
// 将连接结果和数据库方言包装到Engine结构体中并返回。
func NewEngine(source, info string) (engine *Engine, err error) {
	log.Info(source, info)
	db, err := sql.Open(source, info)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if err = db.Ping(); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	dialect, ok := dialect.GetDialect(source)
	if !ok {
		return nil, errors.New("dialect not exist")
	}
	log.Info("connect successfully")
	return &Engine{
		s:       db,
		dialect: dialect,
	}, nil
}

// Close 使Engine中获得的链接关闭，关闭之后此Engine派生的Session将不可用。
func (e *Engine) Close() (err error) {
	if err = e.s.Close(); err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("connection closed")
	return nil
}

// NewSession 使用Engine的参数得到一个Session，可以用于注册sql语句和执行。
func (e Engine) NewSession() *session.Session {
	return session.New(e.s, e.dialect)
}

type TxFunc func(*session.Session) (interface{}, error)

// Transaction 参数为匿名函数func(s *session.Session) (result interface{}, err error) {}
// 内部包含调用s的很多方法，Transaction会将engine的session给匿名函数，并注册defer自动提交和回滚
// 注意事项：
// 1.mysql涉及到对表结构更改的语句会自动提交事务
// 2.每次事务提交，tx自动置为nil，除非下次继续调用session.Begin
func (e *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return f(s)
}
