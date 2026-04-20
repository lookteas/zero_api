// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type MysqlConf struct {
	DataSource string
}

type CycleConf struct {
	TotalPoints int64
}

type Config struct {
	rest.RestConf
	Mysql MysqlConf
	Cycle CycleConf
}
