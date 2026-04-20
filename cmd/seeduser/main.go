package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"api/internal/config"
	"api/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"crypto/sha256"
	"encoding/hex"
)

var configFile = flag.String("f", "../../etc/zero-api.yaml", "config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	conn := sqlx.NewMysql(c.Mysql.DataSource)
	usersModel := model.NewUsersModel(conn)

	ctx := context.Background()

	if _, err := usersModel.FindOneByAccount(ctx, "demo_user"); err == nil {
		fmt.Println("demo user already exists")
		return
	}

	_, err := usersModel.Insert(ctx, &model.Users{
		Account:      "demo_user",
		Email:        sql.NullString{String: "demo@example.com", Valid: true},
		Mobile:       sql.NullString{String: "13800000000", Valid: true},
		PasswordHash: hashPassword("123456"),
		Nickname:     "演示用户",
		Avatar:       "",
		Status:       1,
		LastLoginAt:  sql.NullTime{},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("demo user seeded: account=demo_user password=123456")
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
