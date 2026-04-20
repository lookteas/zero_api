# Zero API

`apps/api` 是 Zero 项目的后端服务，基于 `go-zero` 搭建。

## 配置数据库

需要填写配置文件：`apps/api/etc/zero-api.yaml`

示例模板见：`apps/api/etc/zero-api.example.yaml`

### 关键配置

```yaml
Mysql:
  DataSource: "your_user:your_password@tcp(127.0.0.1:3306)/zero_app?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
```

### 注意事项

- 数据库名默认使用 `zero_app`
- 必须开启 `parseTime=true`
- 字符集建议使用 `utf8mb4`
- 若未填写 `DataSource`，系统会回退到演示数据模式，方便前端继续联调

## 当前已接数据库的模块

- `topics`
- `daily_tasks`
- `daily_logs`
- `review_items`
- `review_records`

## 初始化数据库

```bash
cd apps/api
go run ./cmd/initdb -f ./etc/zero-api.yaml
```

## 初始化主题数据

可执行 `docs/sql/002_seed_topics.sql` 中的 SQL。

如果你愿意，后续也可以继续补一个专门的种子命令。

## 启动服务

```bash
cd apps/api
go run zero.go -f etc/zero-api.yaml
```

默认地址：`http://localhost:8888/api/v1`

## 当前行为说明

- 已配置数据库时：优先走 MySQL 查询与写入
- 未配置数据库时：自动返回演示占位数据，方便前端继续开发和联调
