package clause

import (
	"geeORM/log"
	"strings"
)

// Clause 结构体用于暂时存放不同语句的sqlStr和参数，使用Set放入
// 使用Build将结构体中已存在的sqlStr和参数拼接并返回
type Clause struct {
	Sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

// Set 调用generate中的对应方法解析vars，并将结果存放到clause中。
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.Sql == nil {
		c.Sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.Sql[name] = sql
	c.sqlVars[name] = vars
}

// Build 将orders中的所有信息按照顺序拼接起来，分别返回sqlStr和参数。
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.Sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	log.Info("Build", vars)
	return strings.Join(sqls, " "), vars
}
